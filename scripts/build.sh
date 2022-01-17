#!/bin/sh
# This must be sh because it is used in the alpine docker container

# from https://stackoverflow.com/questions/242538/unix-shell-script-find-out-which-directory-the-script-file-resides?rq=1
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTS_DIR=$(dirname "$SCRIPT")

P_ROOT=$(cd $SCRIPTS_DIR/.. && pwd)
GOPATH=$P_ROOT/lib

if [ $# -ne 1 ]; then
    echo "Usage: build.sh [service-name]"
    exit 1
fi

if [ ! -d $P_ROOT/src/$1 ]; then
    echo "no such directory: $1"
    exit 1
fi
SERVICE_NAME=$1

cd $P_ROOT/src/$SERVICE_NAME
GOPATH=$GOPATH go build

BIN_DIR=$P_ROOT/bin
if [ ! -d $BIN_DIR ]; then
    mkdir $BIN_DIR
fi
mv ./$SERVICE_NAME $BIN_DIR