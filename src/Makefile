LIB:=optimizely-sdk.so
HEADER:=optimizely-sdk.h
SRC:=go-sdk-wrapper.go
CC:=go build
CC_OPTS:=-buildmode=c-shared

.PHONY: all
all: ${LIB} install

${LIB}: ${SRC}
	${CC} ${CC_OPTS} -o $@ $^

.PHONY: test
test:
	go test -v

.PHONY: install
install:
	cp ${HEADER} ../optimizely
	cp ${LIB} ../examples

# Note: cross compiling doesn't work for native code generation
#GOOS=linux ${CC} ${CC_OPTS} -o $@ $^

.PHONY: clean
clean:
	rm -rf *.so *.h
