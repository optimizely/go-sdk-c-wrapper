# Optimizely C SDK

This repository houses an *Unsupported* C SDK for use with Optimizely Full Stack and Optimizely Rollouts.

## Installation

Download then build the SDK as shown below.

### Install from source:

```
cd src && make
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
optimizely_sdk_is_feature_enabled(handle, "<feature name>", "oeu1383080393924r0.5047421827912331");
```

For a full example see [examples/is-feature-enabled.c](https://github.com/optimizely/c-sdk/blob/master/examples/is-feature-enabled.c).

## Credits

This software is used with additional code that is separately downloaded by you. These components are subject to their own license terms which you should review carefully.

Golang (c) 2009 The Go Authors License https://github.com/golang/go
