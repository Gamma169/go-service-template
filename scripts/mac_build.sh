#! /usr/bin/env bash

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

P_ROOT=$(cd $SCRIPTS_DIR/.. && pwd)
GOPATH=$P_ROOT/lib

if [ $# -ne 1 ]; then
    echo "Usage: build [service-name]"
    exit 1
fi

if [ ! -d $P_ROOT/src/$1 ]; then
    echo "no such directory: $1"
    exit 1
fi
SERVICE_NAME=$1

cd $P_ROOT/src/$SERVICE_NAME
GOPATH=$GOPATH go build

if [ -d .git ]; then
    git log -1 | awk 'NR==1{print $2}' > $P_ROOT/git_sha.txt
fi

BIN_DIR=$P_ROOT/bin
if [ ! -d $BIN_DIR ]; then
    mkdir $BIN_DIR
fi
mv ./$SERVICE_NAME $BIN_DIR