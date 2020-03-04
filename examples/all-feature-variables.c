/*
 * =====================================================================================
 *
 *       Filename: enabled-features.c
 *
 *    Description: Demo of the Optimizely SDK in C, get all features
 *
 *        Version: 1.0
 *        Created: 02/04/2020 15:31:46
 *       Revision: none
 *       Compiler: gcc
 *
 *         Author: Ola Nordstrom (ola.nordstrom@optimizely.com)
 *   Organization: Optimizely 
 *
 * =====================================================================================
 */
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include <optimizely/optimizely-sdk.h>

int main(int argc, char *argv[])
{
	char *sdkkey = getenv("OPTIMIZELY_SDKKEY"); // "YOUR SDK KEY";
	char *feature_name = getenv("OPTIMIZELY_FEATURE_NAME"); // "SOME FEATURE NAME";
	char *user_id = getenv("OPTIMIZELY_END_USER_ID"); // "OPTIMIZELY END USER ID";

	optimizely_user_attributes attrib = {0};

	attrib.id = user_id;

	if (sdkkey == NULL) {
		printf("no SDKKEY available\n");
		return -1;
	}
	int handle = optimizely_sdk_client(sdkkey);
	if (handle == -1) {
		fprintf(stderr, "failed to initialize Optimizely SDK\n");
		return 1;
	}
	char *err = NULL;
	int len;
	int enabled;
//func optimizely_sdk_get_all_feature_variables(handle int32, feature_key *C.char, attribs C.struct_optimizely_user_attributes, enabled *C.int, count *C.int, err **C.char) **C.char {
	char **features = optimizely_sdk_get_all_feature_variables(handle, feature_name, attrib, &enabled, &len, &err);
	if (err != NULL) {
		fprintf(stderr, "failed: %s\n", err);
		return 1;
	}
	optimizely_sdk_delete_client(handle); // cleanup

	printf("len: %d, enabled: %d\n", len, enabled);
	for (int i = 0; i < len; i++) {
		printf("feature %d: %s\n", i, features[i]);
	}

	return 0;
}
