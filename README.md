# Optimizely C SDK

This repository houses an *Unsupported* C SDK for use with Optimizely Full Stack and Optimizely Rollouts.

## Installation

Download then build the SDK as shown below.

### Install from source:

```
make
```

## Usage

### Instantiation

Include the headerfile and initialize the SDK with an SDK Key. The returned handle should be used in subsequent calls.
```
#include <optimizely/optimizely-sdk.h>
. . .
int handle = optimizely_sdk_client("<sdk key>");
```

### Feature Rollouts

To see if a feature has been enabled initialize the SDK then call `is_feature_enabled` function.
```
. . .
int handle = optimizely_sdk_client(sdkkey);
if (handle == -1) {
	fprintf(stderr, "failed to initialize Optimizely SDK\n");
	return 1;
}
char *err = NULL;
int enabled = optimizely_sdk_is_feature_enabled(handle, feature_name, &attrib, &err);
. . .
```

For a full example see [examples/is-feature-enabled.c](https://github.com/optimizely/c-sdk/blob/master/examples/is-feature-enabled.c).

## Credits

This software is built with the following software.

* Golang (c) 2009 The Go Authors License, [BSD 3-Clause](https://golang.org/LICENSE)
