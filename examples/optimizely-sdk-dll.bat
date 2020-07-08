g++ -c -DOPTIMIZELY_SDK_DLL optimizely-sdk-dll.cpp
g++ -shared -o optimizely-sdk-dll.dll optimizely-sdk-dll.o -Wl,--out-implib,optimizely-sdk-dll.lib -L. -l:optimizely-sdk.so
