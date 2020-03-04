CC = gcc
CFLAGS = -I.. -Wall
LIBS = optimizely-sdk.so
EXES = is-feature-enabled variable-boolean

.PHONY: all debug clean

all: $(EXES)

debug: CFLAGS += -g
debug: all

is-feature-enabled: is-feature-enabled.c
	$(CC) $(CFLAGS) $(LIBS) $? -o $@

variable-boolean: variable-boolean.c
	$(CC) $(CFLAGS) $(LIBS) $? -o $@

clean:
	-rm $(EXES)
