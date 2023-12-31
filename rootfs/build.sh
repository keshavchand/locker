#!/bin/sh

go build -tags netgo -ldflags '-extldflags "-static"' -o main test_programs.go
