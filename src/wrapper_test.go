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

import (
	"fmt"
	"github.com/optimizely/go-sdk/pkg/entities"
	"os"
	"testing"
)

func TestClientInit(t *testing.T) {
	c1 := optimizelySdkClient(os.Getenv("OPTIMIZELY_SDKKEY"))
	c2 := optimizelySdkClient(os.Getenv("OPTIMIZELY_SDKKEY"))
	fmt.Printf("c1: %v c2: %v\n", c1, c2)
	if c1 == c2 {
		t.Fatal("different client instances not returned")
	}
}

func TestClientDelete(t *testing.T) {
	h := optimizelySdkClient(os.Getenv("OPTIMIZELY_SDKKEY"))
	optimizely_sdk_delete_client(h)

	u := entities.UserContext{ID: "i'm a user Id"}
	rv := optimizelySdkIsFeatureEnabled(h, "", u) // this should not work
	if rv != -1 {
		t.Fatal("invalid handle error not returned")
	}
}

func TestGetFeatureVariableBoolen(t *testing.T) {
	h := optimizelySdkClient(os.Getenv("OPTIMIZELY_SDKKEY"))

	u := entities.UserContext{ID: os.Getenv("OPTIMIZELY_END_USER_ID")}
	rv, err := optimizelySdkGetFeatureVariableBoolean(h, os.Getenv("OPTIMIZELY_FEATURE_NAME"), "boolvar", u)
	if err != nil {
		t.Fatal(err)
	}

	// free all the sturct
	optimizely_sdk_delete_client(h)
	t.Log("optimizelySdkGetFeatureVariableBoolean: ", rv)
}

func TestGetFeatureVariableString(t *testing.T) {
	h := optimizelySdkClient(os.Getenv("OPTIMIZELY_SDKKEY"))

	u := entities.UserContext{ID: os.Getenv("OPTIMIZELY_END_USER_ID")}
	rv, err := optimizelySdkGetFeatureVariableString(h, os.Getenv("OPTIMIZELY_FEATURE_NAME"), "stringvar", u)
	if err != nil {
		t.Fatal(err)
	}

	// free all the sturct
	optimizely_sdk_delete_client(h)
	t.Log("optimizelySdkGetFeatureVariableDouble: ", rv)
}

func TestGetFeatureVariableDouble(t *testing.T) {
	h := optimizelySdkClient(os.Getenv("OPTIMIZELY_SDKKEY"))

	u := entities.UserContext{ID: os.Getenv("OPTIMIZELY_END_USER_ID")}
	rv, err := optimizelySdkGetFeatureVariableDouble(h, os.Getenv("OPTIMIZELY_FEATURE_NAME"), "doublevar", u)
	if err != nil {
		t.Fatal(err)
	}

	// free all the sturct
	optimizely_sdk_delete_client(h)
	t.Log("optimizelySdkGetFeatureVariableDouble: ", rv)
}

func TestGetFeatureVariableInteger(t *testing.T) {
	h := optimizelySdkClient(os.Getenv("OPTIMIZELY_SDKKEY"))

	u := entities.UserContext{ID: os.Getenv("OPTIMIZELY_END_USER_ID")}
	rv, err := optimizelySdkGetFeatureVariableInteger(h, os.Getenv("OPTIMIZELY_FEATURE_NAME"), "integervar", u)
	if err != nil {
		t.Fatal(err)
	}

	// free all the sturct
	optimizely_sdk_delete_client(h)
	t.Log("optimizelySdkGetFeatureVariableInteger: ", rv)
}
