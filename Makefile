.PHONY: all build 
all: build

build:
	cd example && go build -o ../bin/example
	