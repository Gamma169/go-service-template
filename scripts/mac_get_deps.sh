#! /usr/bin/env bash

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PROJECT_ROOT=$(cd $SCRIPTS_DIR/.. && pwd)

LIB_DIR=$PROJECT_ROOT/lib
if [ ! -d $LIB_DIR/src ]; then
    mkdir -p $LIB_DIR/src
fi

GOPATH=$LIB_DIR

cd $PROJECT_ROOT
GOPATH=$GOPATH go get -d ./src/...
