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

typedef struct optimizely_sdk_feature_variable_boolean {
	char *err;
	_Bool rv;
} optimizely_sdk_feature_variable_boolean;
*/
import "C"

import (
	"errors"
	"fmt"
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
	if optlyErr != nil {
		s := optlyErr.Error()
		c_str := C.CString(s) // this allocates a string, caller must free it
		return c_str
	}
	return nil
}

//export optimizely_sdk_free
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
	optimizely_sdk_close(handle)
	optlyClients.lock.Lock()
	delete(optlyClients.m, handle)
	optlyClients.lock.Unlock()
}

//export optimizely_sdk_is_feature_enabled
func optimizely_sdk_is_feature_enabled(handle int32, feature_name *C.char, attribs C.struct_optimizely_user_attributes, errno *C.int) int32 {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		optlyErr = errors.New("no client exists with the specified handle id")
		return -1
	}
	fmt.Printf("in optimizely_sdk_is_feature_enabled, errno is: %v derefrenced: %v\n", errno, &errno)

	// TODO loop through the attributes in the attributes and initialize the UserContext Attributes map
	u := entities.UserContext{ID: C.GoString(attribs.id)}

	enabled, err := optlyClient.IsFeatureEnabled(C.GoString(feature_name), u)

	if err != nil {
		fmt.Printf("errors is not nil, it is: %v\n", err)
		optlyErr = err
		//		*errno = 1
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

	//var errno *C.int
	/*
		errno := C.malloc(C.sizeof_int)
		rv := optimizely_sdk_is_feature_enabled(handle, feature_name, attribs, (*C.int)(errno))
	*/

	var errno2 C.int
	errno2 = 0
	//fmt.Printf("in optimizelySdkIsFeatureEnabled, errno is: %v derefrenced: %v\n", errno2, &errno2)
	rv := optimizely_sdk_is_feature_enabled(handle, feature_name, attribs, &errno2)

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(user))
	return rv
}

/*
//export optimizely_sdk_get_feature_variable_string
func optimizely_sdk_get_feature_variable_string(handle int32, feature_name *C.char, variable_key *C.char, attribs C.struct_optimizely_user_attributes) *C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		optlyErr = errors.New("no client exists with the specified handle id")
		return nil
	}

	u := entities.UserContext{ID: C.GoString(attribs.id)}
	s, err := optlyClient.GetFeatureVariableString(C.GoString(feature_name), C.GoString(variable_key), u)
	if err != nil {
		optlyErr = err
		return nil
	}

	return C.CString(s) // caller must free string
}

func optimizelySdkGetFeatureVariableString(handle int32, featureName string, variableKey string, userCtx entities.UserContext) string {
	feature_name := C.CString(featureName)
	variable_key := C.CString(variableKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	s := optimizely_sdk_get_feature_variable_string(handle, feature_name, variable_key, attribs)
	str := C.GoString(s)

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(variable_key))
	C.free(unsafe.Pointer(user))
	C.free(unsafe.Pointer(s))

	return str
}
*/

//export optimizely_sdk_get_feature_variable_string
func optimizely_sdk_get_feature_variable_string(handle int32, feature_name *C.char, variable_key *C.char, attribs C.struct_optimizely_user_attributes, err **C.char) *C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return nil
	}

	// TODO get rest of the attributes
	u := entities.UserContext{ID: C.GoString(attribs.id)}
	s, e := optlyClient.GetFeatureVariableString(C.GoString(feature_name), C.GoString(variable_key), u)
	if e != nil {
		*err = C.CString(e.Error())
		return nil
	}
	*err = nil

	return C.CString(s) // caller must free
}

func optimizelySdkGetFeatureVariableString(handle int32, featureKey string, variableKey string, userCtx entities.UserContext) (string, error) {
	feature_name := C.CString(featureKey)
	variable_key := C.CString(variableKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	s := optimizely_sdk_get_feature_variable_string(handle, feature_name, variable_key, attribs, &err)
	if err != nil {
		e = errors.New(C.GoString(err))
	} else {
		e = nil
	}

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(variable_key))
	C.free(unsafe.Pointer(user))

	str := C.GoString(s)
	C.free(unsafe.Pointer(s))

	return str, e
}

//export optimizely_sdk_get_feature_variable_boolean
func optimizely_sdk_get_feature_variable_boolean(handle int32, feature_name *C.char, variable_key *C.char, attribs C.struct_optimizely_user_attributes, err **C.char) bool {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return false
	}

	// TODO get rest of the attributes
	u := entities.UserContext{ID: C.GoString(attribs.id)}
	b, e := optlyClient.GetFeatureVariableBoolean(C.GoString(feature_name), C.GoString(variable_key), u)
	if e != nil {
		*err = C.CString(e.Error())
		return false
	}
	*err = nil

	return b
}

func optimizelySdkGetFeatureVariableBoolean(handle int32, featureKey string, variableKey string, userCtx entities.UserContext) (bool, error) {
	feature_name := C.CString(featureKey)
	variable_key := C.CString(variableKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	rv := optimizely_sdk_get_feature_variable_boolean(handle, feature_name, variable_key, attribs, &err)
	if err != nil {
		e = errors.New(C.GoString(err))
	} else {
		e = nil
	}

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(variable_key))
	C.free(unsafe.Pointer(user))

	return rv, e
}

//export optimizely_sdk_get_feature_variable_double
func optimizely_sdk_get_feature_variable_double(handle int32, feature_name *C.char, variable_key *C.char, attribs C.struct_optimizely_user_attributes, err **C.char) float64 {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return 0
	}

	// TODO get rest of the attributes
	u := entities.UserContext{ID: C.GoString(attribs.id)}
	d, e := optlyClient.GetFeatureVariableDouble(C.GoString(feature_name), C.GoString(variable_key), u)
	if e != nil {
		*err = C.CString(e.Error())
		return 0
	}
	*err = nil

	return d
}

func optimizelySdkGetFeatureVariableDouble(handle int32, featureKey string, variableKey string, userCtx entities.UserContext) (float64, error) {
	feature_name := C.CString(featureKey)
	variable_key := C.CString(variableKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	rv := optimizely_sdk_get_feature_variable_double(handle, feature_name, variable_key, attribs, &err)
	if err != nil {
		e = errors.New(C.GoString(err))
	} else {
		e = nil
	}

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(variable_key))
	C.free(unsafe.Pointer(user))

	return rv, e
}

//func optimizelySdkGetFeatureVariableInteger(featureKey string, variableKey string, userContext entities.UserContext) : int, error

//export optimizely_sdk_get_feature_variable_integer
func optimizely_sdk_get_feature_variable_integer(handle int32, feature_name *C.char, variable_key *C.char, attribs C.struct_optimizely_user_attributes, err **C.char) int {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return 0
	}

	// TODO get rest of the attributes
	u := entities.UserContext{ID: C.GoString(attribs.id)}
	i, e := optlyClient.GetFeatureVariableInteger(C.GoString(feature_name), C.GoString(variable_key), u)
	if e != nil {
		*err = C.CString(e.Error())
		return 0
	}
	*err = nil

	return i
}

func optimizelySdkGetFeatureVariableInteger(handle int32, featureKey string, variableKey string, userCtx entities.UserContext) (int, error) {
	feature_name := C.CString(featureKey)
	variable_key := C.CString(variableKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	rv := optimizely_sdk_get_feature_variable_integer(handle, feature_name, variable_key, attribs, &err)
	if err != nil {
		e = errors.New(C.GoString(err))
	} else {
		e = nil
	}

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(variable_key))
	C.free(unsafe.Pointer(user))

	return rv, e
}

//export optimizely_sdk_get_variation
func optimizely_sdk_get_variation(handle int32, experiment_key *C.char, attribs C.struct_optimizely_user_attributes, err **C.char) *C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return nil
	}

	// TODO get rest of the attributes
	u := entities.UserContext{ID: C.GoString(attribs.id)}
	s, e := optlyClient.GetVariation(C.GoString(experiment_key), u)
	if e != nil {
		*err = C.CString(e.Error())
		return nil
	}
	*err = nil

	return C.CString(s)
}

func optimizelySdkGetVariation(handle int32, experimentKey string, variableKey string, userCtx entities.UserContext) (string, error) {
	experiment_key := C.CString(experimentKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	s := optimizely_sdk_get_variation(handle, experiment_key, attribs, &err)
	if err != nil {
		e = errors.New(C.GoString(err))
	} else {
		e = nil
	}

	C.free(unsafe.Pointer(experiment_key))
	C.free(unsafe.Pointer(user))

	str := C.GoString(s)
	C.free(unsafe.Pointer(s))

	return str, e // caller must free string
}

//export optimizely_sdk_get_feature_variable
func optimizely_sdk_get_feature_variable(handle int32, feature_name *C.char, variable_key *C.char, attribs C.struct_optimizely_user_attributes, variable_type **C.char, err **C.char) *C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return nil
	}

	// TODO get rest of the attributes
	u := entities.UserContext{ID: C.GoString(attribs.id)}
	s, varType, e := optlyClient.GetFeatureVariable(C.GoString(feature_name), C.GoString(variable_key), u)
	if e != nil {
		*err = C.CString(e.Error())
		return nil
	}
	*variable_type = C.CString(string(varType))
	*err = nil

	return C.CString(s) // caller must free
}

func optimizelySdkGetFeatureVariable(handle int32, featureKey string, variableKey string, userCtx entities.UserContext) (string, string /*entities.VariableType*/, error) {
	feature_name := C.CString(featureKey)
	variable_key := C.CString(variableKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var var_type *C.char
	var e error
	s := optimizely_sdk_get_feature_variable(handle, feature_name, variable_key, attribs, &var_type, &err)
	if err != nil {
		e = errors.New(C.GoString(err))
	} else {
		e = nil
	}

	C.free(unsafe.Pointer(feature_name))
	C.free(unsafe.Pointer(variable_key))
	C.free(unsafe.Pointer(user))

	str := C.GoString(s)
	C.free(unsafe.Pointer(s))

	str2 := C.GoString(var_type)
	C.free(unsafe.Pointer(var_type))

	return str, str2, e
}

// func (o *OptimizelyClient) Activate(experimentKey string, userContext entities.UserContext) (result string, err error) {

//export optimizely_sdk_activate
func optimizely_sdk_activate(handle int32, experiment_key *C.char, attribs C.struct_optimizely_user_attributes, err **C.char) *C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return nil
	}

	// TODO get rest of the attributes
	u := entities.UserContext{ID: C.GoString(attribs.id)}
	s, e := optlyClient.Activate(C.GoString(experiment_key), u)
	if e != nil {
		*err = C.CString(e.Error())
		return nil
	}
	*err = nil

	return C.CString(s) // caller must free
}

func optimizelySdkActivate(handle int32, experimentKey string, userCtx entities.UserContext) (string, error) {
	experiment_key := C.CString(experimentKey)
	user := C.CString(userCtx.ID)
	attribs := C.struct_optimizely_user_attributes{
		id:                  user,
		user_attribute_list: nil,
	}

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	s := optimizely_sdk_activate(handle, experiment_key, attribs, &err)
	if err != nil {
		e = errors.New(C.GoString(err))
	} else {
		e = nil
	}

	C.free(unsafe.Pointer(experiment_key))

	str := C.GoString(s)
	C.free(unsafe.Pointer(s))

	return str, e
}

func optimizely_sdk_close(handle int32) {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if ok {
		optlyClient.Close()
	}
}

func main() {
}
