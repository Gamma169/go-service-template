#!/bin/bash

# This can be bash because it is not used in the docker build

# from https://stackoverflow.com/questions/242538/unix-shell-script-find-out-which-directory-the-script-file-resides?rq=1
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTS_DIR=$(dirname "$SCRIPT")

PROJECT_ROOT=$(cd $SCRIPTS_DIR/.. && pwd)

pushd ${PROJECT_ROOT}/tests > /dev/null

yarn install
yarn test

status_code=$?

if [ -f yarn-error.log ]; then
    rm yarn-error.log
fi

if [ $status_code -ne 0 ]; then
    exit $status_code
fi

popd > /dev/null
