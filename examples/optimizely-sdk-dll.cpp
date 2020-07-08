#include "optimizely-sdk-dll.h"

void OptimizelyInit()
{
	optimizely_sdk_init();
}

GoInt32 OptimizelyClient(char* p0)
{
	return optimizely_sdk_client(p0);
}

void OptimizelyDeleteClient(GoInt32 p0)
{
	optimizely_sdk_delete_client(p0);
}

GoInt32 OPTIMIZELY_SDK OptimizelyIsFeatureEnabled(GoInt32 p0, char* p1, optimizely_user_attributes* p2, char** p3)
{
	return optimizely_sdk_is_feature_enabled(p0, p1, p2, p3);
}

char* OPTIMIZELY_SDK OptimizelyGetFeatureVariableString(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4)
{
	return optimizely_sdk_get_feature_variable_string(p0, p1, p2, p3, p4);
}

GoUint8 OPTIMIZELY_SDK OptimizelyGetFeatureVariableBoolean(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4)
{
	return optimizely_sdk_get_feature_variable_boolean(p0, p1, p2, p3, p4);
}

GoFloat64 OPTIMIZELY_SDK OptimizelyGetFeatureVariableDouble(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4)
{
	return optimizely_sdk_get_feature_variable_double(p0, p1, p2, p3, p4);
}

GoInt OPTIMIZELY_SDK OptimizelyGetFeatureVariableInteger(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4)
{
	return optimizely_sdk_get_feature_variable_integer(p0, p1, p2, p3, p4);
}

char* OPTIMIZELY_SDK OptimizelyGetVariation(GoInt32 p0, char* p1, optimizely_user_attributes* p2, char** p3)
{
	return optimizely_sdk_get_variation(p0, p1, p2, p3);
}

char* OPTIMIZELY_SDK OptimizelyGetFeatureVariable(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4, char** p5)
{
	return  optimizely_sdk_get_feature_variable(p0, p1, p2, p3, p4, p5);
}

char* OPTIMIZELY_SDK OptimizelyActivate(GoInt32 p0, char* p1, optimizely_user_attributes* p2, char** p3)
{
	return  optimizely_sdk_activate(p0, p1, p2, p3);
}

char** OPTIMIZELY_SDK OptimizelyGetEnabledFeatures(GoInt32 p0, optimizely_user_attributes* p1, int* p2, char** p3)
{
	return  optimizely_sdk_get_enabled_features(p0, p1, p2, p3);
}

char** OPTIMIZELY_SDK OptimizelyGetAllFeatureVariables(GoInt32 p0, char* p1, optimizely_user_attributes* p2, int* p3, int* p4, char** p5)
{
	return  optimizely_sdk_get_all_feature_variables(p0, p1, p2, p3, p4, p5);
}

char* OPTIMIZELY_SDK OptimizelyTrack(GoInt32 p0, char* p1, optimizely_user_attributes* p2, float* p3, char** p4)
{
	return  optimizely_sdk_track(p0, p1, p2, p3, p4);
}
