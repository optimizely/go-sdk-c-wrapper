CC = gcc
CFLAGS = -I.. -Wall
LIBS = optimizely-sdk.so
EXES = is-feature-enabled

.PHONY: all debug clean

all: $(EXES)

debug: CFLAGS += -g
debug: all

is-feature-enabled: is-feature-enabled.c
	$(CC) $(CFLAGS) $(LIBS) $? -o $@

clean:
	-rm $(EXES)
