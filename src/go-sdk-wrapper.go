package main

/*
#include <stdlib.h>
//#include "user-attributes.h"

union optimizely_attribute {
    _Bool bdata;
    char *sdata;
    float fdata;
};

typedef struct optimizely_user_attribute {
    char *name;
    _Bool type; // 1 = bool, 2 = char, 3 = float
    union optimizely_attribute attr;
} optimzely_user_attribute;

typedef struct optimizely_user_attributes{
    char *id;
    struct optimizely_user_attribute *user_attribute_list;
} optimizely_user_attributes;
*/
import "C"

import (
	"errors"
	optly "github.com/optimizely/go-sdk"
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/entities"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

type optimizelyClientMap struct {
	lock       *sync.RWMutex
	m          map[int32]*client.OptimizelyClient
	randSource *rand.Rand
}

var (
	optlyClients *optimizelyClientMap
	optlyErr     error // track the last error
)

func init() {
	// it is safe to initialize multiple times
	optimizely_sdk_init()
}

//export optimizely_sdk_init
func optimizely_sdk_init() uint32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	optlyClients = &optimizelyClientMap{
		lock:       new(sync.RWMutex),
		m:          make(map[int32]*client.OptimizelyClient),
		randSource: r,
	}
	return 0
}

// important: caller must free error string
//export optimizely_sdk_get_error
func optimizely_sdk_get_error() *C.char {
	s := optlyErr.Error()
	c_str := C.CString(s) // this allocates a string, caller must free it
	return c_str
}

// export optimizely_sdk_free
func optimizely_sdk_free(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export optimizely_sdk_client
func optimizely_sdk_client(sdkkey *C.char) int32 {
	optlyClients.lock.Lock()
	c, err := optly.Client(C.GoString(sdkkey))
	if err != nil {
		optlyErr = err
		return -1
	}
	handle := optlyClients.randSource.Int31()
	if _, ok := optlyClients.m[handle]; ok {
		// try one more time
		handle = optlyClients.randSource.Int31()
		if _, ok := optlyClients.m[handle]; ok {
			panic("unable to insert into handle map")
		}
	}
	optlyClients.m[handle] = c
	optlyClients.lock.Unlock()
	return handle
}

func optimizelySdkClient(sdkkey string) int32 {
	s := C.CString(sdkkey)
	rv := optimizely_sdk_client(s)
	C.free(unsafe.Pointer(s))
	return rv
}

//export optimizely_sdk_delete_client
func optimizely_sdk_delete_client(handle int32) {
	optlyClients.lock.Lock()
	delete(optlyClients.m, handle)
	optlyClients.lock.Unlock()
}

//export optimizely_sdk_is_feature_enabled
func optimizely_sdk_is_feature_enabled(handle int32, feature_name *C.char, attributes_list C.struct_optimizely_user_attributes) int32 {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		optlyErr = errors.New("no client exists with the specified handle id")
		return -1
	}

	// TODO loop through the attributes in the attributes_list and initialize the UserContext Attributes map
	u := entities.UserContext{ID: C.GoString(attributes_list.id)}

	enabled, err := optlyClient.IsFeatureEnabled(C.GoString(feature_name), u)

	if err != nil {
		optlyErr = err
		return -1
	}

	if enabled {
		return 1
	} else {
		return 0
	}
}

func optimizelySdkIsFeatureEnabled(handle int32, featureName string, userCtx entities.UserContext) int32 {
	feature_name := C.CString(featureName)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	rv := optimizely_sdk_is_feature_enabled(handle, feature_name, attribs)

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(user))
	return rv
}

//export optimizely_sdk_get_feature_variable_string
func optimizely_sdk_get_feature_variable_string(handle int32, feature_name *C.char, variable_key *C.char, user *C.char) *C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		optlyErr = errors.New("no client exists with the specified handle id")
		return nil
	}

	u := entities.UserContext{ID: C.GoString(user)}
	s, err := optlyClient.GetFeatureVariableString(C.GoString(feature_name), C.GoString(variable_key), u)
	if err != nil {
		optlyErr = err
		return nil
	}

	return C.CString(s) // caller must free string
}

func main() {
}
