#pragma once

// Visual Studio doesn't understand __SIZE_TYPE__
#ifndef __SIZE_TYPE__
typedef uintptr_t __SIZE_TYPE__;
#endif

// Complex number is not supported atm.
#ifndef _Complex
#define _Complex
#endif

#include "../optimizely/optimizely-sdk.h"

#ifdef __cplusplus
extern "C" {
#endif

#ifdef OPTIMIZELY_SDK_DLL
#define OPTIMIZELY_SDK __declspec(dllexport)
#else
#define OPTIMIZELY_SDK __declspec(dllimport)
#endif

OPTIMIZELY_SDK void OptimizelyInit();

OPTIMIZELY_SDK GoInt32 OptimizelyClient(char* p0);

OPTIMIZELY_SDK void OptimizelyDeleteClient(GoInt32 p0);

OPTIMIZELY_SDK GoInt32 OptimizelyIsFeatureEnabled(GoInt32 p0, char* p1, optimizely_user_attributes* p2, char** p3);

OPTIMIZELY_SDK char* OptimizelyGetFeatureVariableString(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4);

OPTIMIZELY_SDK GoUint8 OptimizelyGetFeatureVariableBoolean(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4);

OPTIMIZELY_SDK GoFloat64 OptimizelyGetFeatureVariableDouble(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4);

OPTIMIZELY_SDK GoInt OptimizelyGetFeatureVariableInteger(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4);

OPTIMIZELY_SDK char* OptimizelyGetVariation(GoInt32 p0, char* p1, optimizely_user_attributes* p2, char** p3);

OPTIMIZELY_SDK char* OptimizelyGetFeatureVariable(GoInt32 p0, char* p1, char* p2, optimizely_user_attributes* p3, char** p4, char** p5);

OPTIMIZELY_SDK char* OptimizelyActivate(GoInt32 p0, char* p1, optimizely_user_attributes* p2, char** p3);

OPTIMIZELY_SDK char** OptimizelyGetEnabledFeatures(GoInt32 p0, optimizely_user_attributes* p1, int* p2, char** p3);

// this only returns the names, not the values

OPTIMIZELY_SDK char** OptimizelyGetAllFeatureVariables(GoInt32 p0, char* p1, optimizely_user_attributes* p2, int* p3, int* p4, char** p5);

OPTIMIZELY_SDK char* OptimizelyTrack(GoInt32 p0, char* p1, optimizely_user_attributes* p2, float* p3, char** p4);

#ifdef __cplusplus
}
#endif
