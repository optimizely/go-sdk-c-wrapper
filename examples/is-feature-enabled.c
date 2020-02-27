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
 *         Author: Ola Nordstrom (ola.nordstrom@optimizely.com),
 *   Organization: Optimizely 
 *
 * =====================================================================================
 */
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

#include <optimizely/optimizely-sdk.h>

int main() {
    /* initialization variables */
    char *sdkkey = "YOUR SDK KEY";
    /* the feature we're checking */
    char *feature_name = "SOME USER ID";

    /* the optimizely end user id */
    char *user_id = "SOME USER ID";

	int handle = optimizely_sdk_client(sdkkey);
	int enabled = optimizely_sdk_is_feature_enabled(handle, feature_name, user_id);
	printf("the feature: %s is enabled: %d\n", feature_name, enabled);

    return 0;
}
