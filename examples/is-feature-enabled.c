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

/*
 * type UserContext struct {
 * 	ID         string
 * 	Attributes map[string]interface{}
 * }
 */

/*
union optimizely_attribute {
    _Bool bdata;
    char *sdata;
    float fdata;
} optimizely_attribute;

typedef struct optimizely_user_attribute {
    _Bool type; // 1 = bool, 2 = char, 3 = float
    union optimizely_attribute attr;
} optimzely_user_attribute;

typedef struct optimizely_user_attributes{
    char *id;
    struct optimizely_user_attribute *user_attribute_list;
} optimizely_user_attributes;

typedef struct optimizely_user_attributes{
    char *id;
    struct optimizely_user_attribute *user_attribute_list;
} optimizely_user_attributes;
*/

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
	int enabled = optimizely_sdk_is_feature_enabled(handle, feature_name, attrib);

	printf("the feature: %s is enabled: %d\n", feature_name, enabled);

	optimizely_sdk_delete_client(handle); // cleanup

	return 0;
}
