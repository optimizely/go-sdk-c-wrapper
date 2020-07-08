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
	"github.com/optimizely/go-sdk/pkg/logging"
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
	logger       = logging.GetLogger("CSDK", "")
)

func init() {
	// it is safe to initialize multiple times
	optimizely_sdk_init()
}

//export optimizely_sdk_init
func optimizely_sdk_init() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	optlyClients = &optimizelyClientMap{
		lock:       new(sync.RWMutex),
		m:          make(map[int32]*client.OptimizelyClient),
		randSource: r,
	}
}

//export optimizely_sdk_client
func optimizely_sdk_client(sdkkey *C.char) int32 {
	optlyClients.lock.Lock()
	c, err := optly.Client(C.GoString(sdkkey))
	if err != nil {
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
	if attrCount > 0 {
		attrList := (*[1 << 30]C.struct_optimizely_user_attribute)(unsafe.Pointer(listPtr.user_attribute_list))[:attrCount:attrCount]

		for i := 0; i < int(attrCount); i++ {
			name := C.GoString(attrList[i].name)
			logger.Debug(fmt.Sprintf(`convertUserContext i:"%d" is attr.name "%s".`, i, name))
			t := attrList[i].var_type
			switch t {
			case 1: // string
				value := C.GoString((*C.char)(attrList[i].data))
				logger.Debug(fmt.Sprintf(`value: "%s".`, name))
				u.Attributes[name] = value //C.GoString((*C.char)(attrList[i].data))
			case 2: // bool
				value := ((*C.int)(attrList[i].data))
				if *value == 0 {
					u.Attributes[name] = false
				} else {
					u.Attributes[name] = true
				}
				logger.Debug(fmt.Sprintf(`value: "%v".`, name))
			case 3: // float
				value := ((*C.float)(attrList[i].data))
				u.Attributes[name] = *value
				logger.Debug(fmt.Sprintf(`value: "%v".`, name))
			case 4: // int
				value := ((*C.int)(attrList[i].data))
				u.Attributes[name] = *value
				logger.Debug(fmt.Sprintf(`value: "%v".`, name))
			default:
				logger.Debug(fmt.Sprintf(`Unknown type specified: %d, skipping`, t))
			}
			logger.Debug(fmt.Sprintf(`map: %v`, u.Attributes))
		}
	}
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
	}

	u, e := convertUserContext(attribs)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return -1
	}

	enabled, e := optlyClient.IsFeatureEnabled(C.GoString(feature_name), *u)
	if e != nil {
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

//export optimizely_sdk_get_feature_variable_string
func optimizely_sdk_get_feature_variable_string(handle int32, feature_name *C.char, variable_key *C.char, attribs *C.struct_optimizely_user_attributes, err **C.char) *C.char {
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

	s, e := optlyClient.GetFeatureVariableString(C.GoString(feature_name), C.GoString(variable_key), *u)
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
	s := optimizely_sdk_get_feature_variable_string(handle, feature_name, variable_key, &attribs, &err)
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
func optimizely_sdk_get_feature_variable_boolean(handle int32, feature_name *C.char, variable_key *C.char, attribs *C.struct_optimizely_user_attributes, err **C.char) bool {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return false
	}

	u, e := convertUserContext(attribs)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return false
	}

	b, e := optlyClient.GetFeatureVariableBoolean(C.GoString(feature_name), C.GoString(variable_key), *u)
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
	attribsRef := &attribs

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	rv := optimizely_sdk_get_feature_variable_boolean(handle, feature_name, variable_key, attribsRef, &err)
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
func optimizely_sdk_get_feature_variable_double(handle int32, feature_name *C.char, variable_key *C.char, attribs *C.struct_optimizely_user_attributes, err **C.char) float64 {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return 0
	}

	u, e := convertUserContext(attribs)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return 0
	}

	d, e := optlyClient.GetFeatureVariableDouble(C.GoString(feature_name), C.GoString(variable_key), *u)
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
	attribsRef := &attribs

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	rv := optimizely_sdk_get_feature_variable_double(handle, feature_name, variable_key, attribsRef, &err)
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

//export optimizely_sdk_get_feature_variable_integer
func optimizely_sdk_get_feature_variable_integer(handle int32, feature_name *C.char, variable_key *C.char, attribs *C.struct_optimizely_user_attributes, err **C.char) int {
	optlyClients.lock.RLock()
	optlyClient, ok := optlyClients.m[handle]
	optlyClients.lock.RUnlock()
	if !ok {
		cstr := C.CString("no client exists with the specified handle id") // this allocates a string, caller must free it
		*err = cstr
		return 0
	}

	u, e := convertUserContext(attribs)
	if e != nil {
		cstr := C.CString(e.Error())
		*err = cstr
		return 0
	}

	i, e := optlyClient.GetFeatureVariableInteger(C.GoString(feature_name), C.GoString(variable_key), *u)
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
	attribsRef := &attribs

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	rv := optimizely_sdk_get_feature_variable_integer(handle, feature_name, variable_key, attribsRef, &err)
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
func optimizely_sdk_get_variation(handle int32, experiment_key *C.char, attribs *C.struct_optimizely_user_attributes, err **C.char) *C.char {
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

	s, e := optlyClient.GetVariation(C.GoString(experiment_key), *u)
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
	attribsRef := &attribs

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	s := optimizely_sdk_get_variation(handle, experiment_key, attribsRef, &err)
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
func optimizely_sdk_get_feature_variable(handle int32, feature_name *C.char, variable_key *C.char, attribs *C.struct_optimizely_user_attributes, variable_type **C.char, err **C.char) *C.char {
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

	s, varType, e := optlyClient.GetFeatureVariable(C.GoString(feature_name), C.GoString(variable_key), *u)
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
	attribsRef := &attribs

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var var_type *C.char
	var e error
	s := optimizely_sdk_get_feature_variable(handle, feature_name, variable_key, attribsRef, &var_type, &err)
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
func optimizely_sdk_activate(handle int32, experiment_key *C.char, attribs *C.struct_optimizely_user_attributes, err **C.char) *C.char {
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

	s, e := optlyClient.Activate(C.GoString(experiment_key), *u)
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
	attribsRef := &attribs

	// TODO loop through the user_context and create the rest of the attribs
	var err *C.char
	var e error
	s := optimizely_sdk_activate(handle, experiment_key, attribsRef, &err)
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
func optimizely_sdk_get_enabled_features(handle int32, attribs *C.struct_optimizely_user_attributes, count *C.int, err **C.char) **C.char {
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

	featureList, e := optlyClient.GetEnabledFeatures(*u)
	if e != nil {
		*err = C.CString(e.Error())
		return nil
	}
	*err = nil

	cArr := C.malloc(C.size_t(len(featureList)) * C.size_t(unsafe.Sizeof(uintptr(0))))

	a := (*[1<<30 - 1]*C.char)(cArr) // a is a pointer to a the c array

	for idx, substring := range featureList {
		a[idx] = C.CString(substring)
	}

	*count = C.int(len(featureList)) // return the count
	return (**C.char)(cArr)          // caller must free
}

// this only returns the names, not the values
//export optimizely_sdk_get_all_feature_variables
func optimizely_sdk_get_all_feature_variables(handle int32, feature_key *C.char, attribs *C.struct_optimizely_user_attributes, enabled *C.int, count *C.int, err **C.char) **C.char {
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

	bEnabled, varMap, e := optlyClient.GetAllFeatureVariablesWithDecision(C.GoString(feature_key), *u)
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
	attribRef := &attribs
	var err *C.char
	var e error
	s := optimizely_sdk_get_feature_variable_string(handle, feature_name, variable_key, attribRef, &err)
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
