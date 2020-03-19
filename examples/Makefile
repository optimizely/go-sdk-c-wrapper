CC := gcc
CLANG_SECURITY_FLAGS := -fstack-protector -D_FORTIFY_SOURCE=2 -O1 -fPIE -Wformat -Wformat-security -Werror=format-security
CFLAGS := -I.. -Wall $(CLANG_SECURITY_FLAGS)
LIBS := optimizely-sdk.so
EXES := is-feature-enabled variable-boolean all-feature-variables enabled-features

.PHONY: all debug clean

all: $(EXES)

debug: CFLAGS += -g
debug: all

is-feature-enabled: is-feature-enabled.c
	$(CC) $(CFLAGS) $(LIBS) $? -o $@

variable-boolean: variable-boolean.c
	$(CC) $(CFLAGS) $(LIBS) $? -o $@

all-feature-variables: all-feature-variables.c
	$(CC) $(CFLAGS) $(LIBS) $? -o $@

enabled-features: enabled-features.c
	$(CC) $(CFLAGS) $(LIBS) $? -o $@

clean:
	-rm $(EXES)
