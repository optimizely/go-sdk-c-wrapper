CC = gcc
CFLAGS = -I.. -Wall
LIBS = optimizely-sdk.so
EXES = is-feature-enabled variable-boolean all-feature-variables

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

clean:
	-rm $(EXES)
