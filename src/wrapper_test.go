package main

import (
	"fmt"
	"testing"
)

func TestClientInit(t *testing.T) {
	if optimizely_sdk_init() != 0 {
		t.Fatal("Failed to initialize the SDK")
	}

	c1 := optimizelySdkClient("some fake key")
	c2 := optimizelySdkClient("some fake key")
	fmt.Printf("c1: %v c2: %v\n", c1, c2)
	if c1 == c2 {
		t.Fatal("different client instances not returned")
	}
}

func TestClientDelete(t *testing.T) {
	if optimizely_sdk_init() != 0 {
		t.Fatal("Failed to initialize the SDK")
	}

	h := optimizelySdkClient("some fake key")
	optimizely_sdk_delete_client(h)
	rv := optimizelySdkIsFeatureEnabled(h, "", "")
	if rv != -1 {
		t.Fatal("invalid handle error not returned")
	}

}
