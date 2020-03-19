/*
 * =====================================================================================
 *
 *       Filename: is-feature-enabled.c
 *
 *    Description: Demo of the Optimizely SDK in C, check to see if a feature is enabled
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
		return 1;
	}

	int handle = optimizely_sdk_client(sdkkey);
	if (handle == -1) {
		fprintf(stderr, "failed to initialize Optimizely SDK\n");
		return 1;
	}
	char *err = NULL;
	int enabled = optimizely_sdk_is_feature_enabled(handle, feature_name, &attrib, &err);
	if (err != NULL) {
		fprintf(stderr, "failed, error: %s\n", err);
		free(err);
	}

	printf("the feature: %s is enabled: %d\n", feature_name, enabled);

	optimizely_sdk_delete_client(handle); // cleanup

	return 0;
}
