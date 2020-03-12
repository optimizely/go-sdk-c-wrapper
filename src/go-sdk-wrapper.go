/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

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
    int var_type; // 1 = string, 2 = bool, 3 = float, 4 = int
    void *data;
} optimzely_user_attribute;

typedef struct optimizely_user_attributes{
    char *id;
    int num_attributes;
    struct optimizely_user_attribute *user_attribute_list;
} optimizely_user_attributes;
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

// returns the UserContext as a go object
func convertUserContext(attribs *C.struct_optimizely_user_attributes) (*entities.UserContext, error) {
	if attribs == nil {
		return nil, errors.New("convertUserContext called with nil attribs")
	}

	listPtr := (*C.optimizely_user_attributes)(unsafe.Pointer(attribs))

	u := entities.UserContext{ID: C.GoString(attribs.id), Attributes: map[string]interface{}{}}

	attrCount := (int)(listPtr.num_attributes)
	fmt.Printf("attrCount: %d\n", attrCount)
	if attrCount > 0 {

		//p := (*[1 << 30]C._profile)(unsafe.Pointer(profiles.profile))[:numProfiles:numProfiles]
		//pola := (*[1 << 30]C.struct_optimizely_user_attribute)(unsafe.Pointer(listPtr.user_attribute_list))[:2:2]
		pola := (*[1 << 30]C.struct_optimizely_user_attribute)(unsafe.Pointer(listPtr.user_attribute_list))[:attrCount:attrCount]
		//pola := (*[1 << 30]C.struct_optimizely_user_attribute)(unsafe.Pointer(attribola))

		for i := 0; i < int(attrCount); i++ {
			name := C.GoString(pola[i].name)
			fmt.Printf("i: %d attr.name: %s\n", i, name)

			t := pola[i].var_type
			fmt.Printf("t: %d\n", t)
			switch t {
			case 1: // string
				value := C.GoString((*C.char)(pola[i].data))
				fmt.Println("value:", value)
				u.Attributes[name] = value //C.GoString((*C.char)(pola[i].data))
			case 2: // bool
				//value := ((*C.Bool)(pola[i].data))
				value := ((*C.int)(pola[i].data))
				if *value == 0 {
					u.Attributes[name] = false
				} else {
					u.Attributes[name] = true
				}
				fmt.Println("value:", *value)
			case 3: // float
				value := ((*C.float)(pola[i].data))
				u.Attributes[name] = *value
				fmt.Println("value:", *value)
			case 4: // int
				value := ((*C.int)(pola[i].data))
				u.Attributes[name] = *value
				fmt.Println("value:", *value)
			default:
				fmt.Printf("Unknown type specified: %d, skipping\n", t)
			}
			fmt.Println("map:", u.Attributes)
		}
	}
	fmt.Printf("Here's the map:\n%+v\n", u.Attributes)
	return &u, nil
}

//export optimizely_sdk_is_feature_enabled
func optimizely_sdk_is_feature_enabled(handle int32, feature_name *C.char, attribs *C.struct_optimizely_user_attributes, err **C.char) int32 {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id")
		*err = cstr
		return -1
	} else {

		fmt.Println("The handle is valid", handle)
	}

	u, e := convertUserContext(attribs)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return -1
	}

	enabled, e := optlyClient.IsFeatureEnabled(C.GoString(feature_name), *u)

	if e != nil {
		fmt.Printf("errors is not nil, it is: %v\n", e)
		cstr := C.CString(e.Error())
		*err = cstr
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

	var s *C.char
	rv := optimizely_sdk_is_feature_enabled(handle, feature_name, &attribs, &s)

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

//export optimizely_sdk_get_enabled_features
func optimizely_sdk_get_enabled_features(handle int32, attribs C.struct_optimizely_user_attributes, count *C.int, err **C.char) **C.char {
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
	featureList, e := optlyClient.GetEnabledFeatures(u)
	if e != nil {
		*err = C.CString(e.Error())
		return nil
	}
	*err = nil

	cArr := C.malloc(C.size_t(len(featureList)) * C.size_t(unsafe.Sizeof(uintptr(0))))

	a := (*[1<<30 - 1]*C.char)(cArr) // a is a pointer to a the c array

	// Confused? See
	// https://stackoverflow.com/questions/48756732/what-does-1-30c-yourtype-do-exactly-in-cgo
	// https://stackoverflow.com/questions/41492071/how-do-i-convert-a-go-array-of-strings-to-a-c-array-of-strings
	// https://stackoverflow.com/questions/53238602/accessing-c-array-in-golang

	for idx, substring := range featureList {
		a[idx] = C.CString(substring)
	}

	*count = C.int(len(featureList)) // return the count
	return (**C.char)(cArr)          // caller must free
}

// this only returns the names, not the values
//export optimizely_sdk_get_all_feature_variables
func optimizely_sdk_get_all_feature_variables(handle int32, feature_key *C.char, attribs *C.struct_optimizely_user_attributes, attribola *C.struct_optimizely_user_attribute, enabled *C.int, count *C.int, err **C.char) **C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return nil
	}

	u, e := convertUserContext(attribs)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return nil
	}

	bEnabled, varMap, e := optlyClient.GetAllFeatureVariables(C.GoString(feature_key), *u)
	if e != nil {
		*err = C.CString(e.Error())
		return nil
	}
	if bEnabled {
		*enabled = 1
	}

	/* now allocate the number of necessary structs and set the data */
	cArr := C.malloc(C.size_t(len(varMap)) * C.size_t(unsafe.Sizeof(uintptr(0))))

	a := (*[1<<30 - 1]*C.char)(cArr) // a is a pointer to a the c array

	// Confused? See
	// https://stackoverflow.com/questions/48756732/what-does-1-30c-yourtype-do-exactly-in-cgo
	// https://stackoverflow.com/questions/41492071/how-do-i-convert-a-go-array-of-strings-to-a-c-array-of-strings
	// https://stackoverflow.com/questions/53238602/accessing-c-array-in-golang

	i := 0
	for key := range varMap {
		a[i] = C.CString(key)
		i++
	}

	*count = C.int(len(varMap)) // return the count
	return (**C.char)(cArr)     // caller must free
}

//export optimizely_sdk_track
func optimizely_sdk_track(handle int32, event_key *C.char, attribs *C.struct_optimizely_user_attributes, value *C.float, err **C.char) *C.char {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return nil
	}

	u, e := convertUserContext(attribs)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return nil
	}

	eventTags := map[string]interface{}{}
	if value != nil {
		eventTags["value"] = *value
	}

	e = optlyClient.Track(C.GoString(event_key), *u, eventTags)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return nil
	}

	return nil
}

func optimizelySdkGetFeatureVariableStringTODO(handle int32, featureKey string, variableKey string, userCtx entities.UserContext) (string, error) {
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
