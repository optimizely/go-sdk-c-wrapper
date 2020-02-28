package main

import (
	"testing"
)

func testClientInit(t *testing.B) {
	if optimizely_sdk_init() != 0 {
		t.Fatal("Failed to initialize the SDK")
	}

	c1 := optimizelySdkClient("some fake key")
	c2 := optimizelySdkClient("some fake key")
	if c1 != c2 {
		t.Fatal("different client instances not returned")
	}
}

func testClientDelete(t *testing.B) {
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
