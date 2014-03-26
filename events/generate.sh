#!/bin/bash

set -e

go install code.google.com/p/gogoprotobuf/protoc-gen-gogo

protoc --plugin=$GOPATH/bin/protoc-gen-gogo --gogo_out=. http.proto