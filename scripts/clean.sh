#!/bin/bash
# This can be bash because it is not used in the docker build

# from https://stackoverflow.com/questions/242538/unix-shell-script-find-out-which-directory-the-script-file-resides?rq=1
# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
SCRIPTS_DIR=$(dirname "$SCRIPT")

PROJECT_ROOT=$(cd $SCRIPTS_DIR/.. && pwd)

cd $PROJECT_ROOT

for name in bin lib tests/node_modules; do
    if [ -d $name ]; then
        rm -rf $name
    fi
done
