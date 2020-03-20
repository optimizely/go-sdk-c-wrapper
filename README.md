# Optimizely C SDK

This repository contains a C SDK for use with Optimizely Full Stack and Optimizely Rollouts. This SDK is currently in Alpha.

## Installation

Download then build the SDK as shown below.

### Install from source:

```makefile
make
```

## Usage

### Instantiation

Include the headerfile and initialize the SDK with an SDK Key. The returned handle should be used in subsequent calls.
```c_cpp
#include <optimizely/optimizely-sdk.h>
. . .
int handle = optimizely_sdk_client("<sdk key>");
```

See API for more details.

### Feature Rollouts

To see if a feature has been enabled initialize the SDK then call `is_feature_enabled` function.
```c_cpp
...
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

## API

**Important:** All strings and string arrays returned by the API must be free'd by the caller.

```c_cpp
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

// reintializes the SDK and creates a new handle table
int optimizely_sdk_init();

// creates a new optimizely SDK and returns a handle or returns -1 if the sdk could not be initialized
int optimizely_sdk_client(char* sdkkey);

// removes the passed in client handle
void optimizely_sdk_delete_client(int handle);

// checks to see if feature_name is enabled, non zero return means the feature is enabled
int optimizely_sdk_is_feature_enabled(int handle, char* feature_name,
                                      optimizely_user_attributes* attributes, char** error);

// returns the string feature variable value
char* optimizely_sdk_get_feature_variable_string(int handle, char* feature_name, char* variable_key, 
                                                 optimizely_user_attributes* attributes, char** error);

// returns the boolean feature variable value
int optimizely_sdk_get_feature_variable_boolean(int handle, char* feature_name, char* variable_key,
                                                optimizely_user_attributes* attributes, char** error);

// returns the double feature variable value
double optimizely_sdk_get_feature_variable_double(int handle, char* feature_name, char* variable_key,
                                                  optimizely_user_attributes* attributes, char** error);

// returns the integer feature variable value
int optimizely_sdk_get_feature_variable_integer(int handle, char* feature_name, char* variable_key,
                                                optimizely_user_attributes* attributes, char** error);

// returns the variation for the specified experiment key and user attributes
char* optimizely_sdk_get_variation(int handle, char* experiment_key,
                                   optimizely_user_attributes* attributes, char** error);

// returns the feature variable for the specified feature_name and variable_key
// the variable_type receives a string specifying the variable type
char* optimizely_sdk_get_feature_variable(int handle, char* feature_name,
                                          char* variable_key, optimizely_user_attributes* attributes,
					  char** variable_type, char** error);

// activates the specified experiment_key
char* optimizely_sdk_activate(int handle, char* experiment_key,
                              optimizely_user_attributes* attributes, char** error);

// returns a list of the enabled features, the count receives the number of features returned
char** optimizely_sdk_get_enabled_features(int handle, optimizely_user_attributes* attributes,
                                           int* count, char** error);

// returns a list of the enabled feature variables, count contains the feature variable count
// the caller must free all returned feature name strings
// this only returns the names of the features 
// to get the value call optimizely_sdk_get_feature_variable_<type>()
char** optimizely_sdk_get_all_feature_variables(int handle, char* feature_key,
                                                optimizely_user_attributes* attributes,
                                                int* enabled, int* count, char** error);

// tracks the specified event_key
char* optimizely_sdk_track(int handle, char* feature_key, optimizely_user_attributes* attributes,
                           float* value, char** error);
```

## Credits

This software is built with the following software.

* Golang (c) 2009 The Go Authors License, [BSD 3-Clause](https://golang.org/LICENSE)
