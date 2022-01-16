#!/bin/sh
# This must be sh because it is used in the alpine docker container


# from https://stackoverflow.com/questions/242538/unix-shell-script-find-out-which-directory-the-script-file-resides?rq=1
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTS_DIR=$(dirname "$SCRIPT")

PROJECT_ROOT=$(cd $SCRIPTS_DIR/.. && pwd)

LIB_DIR=$PROJECT_ROOT/lib
if [ ! -d $LIB_DIR/src ]; then
    mkdir -p $LIB_DIR/src
fi

GOPATH=$LIB_DIR

cd $PROJECT_ROOT

if [ $# -eq 0 ]; then
    GOPATH=$GOPATH go get -d ./src/...
elif [ $# -eq 1 ]; then
    GOPATH=$GOPATH go get $1
else
    echo "Usage: get_deps.sh [package name]"
    echo "If [package name] is not provided, will get all deps needed in src"
    exit 1
fi
