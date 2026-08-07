#!/bin/sh
SCRIPT_NAME=`basename $0 | cut -f1 -d'.'`
WORK_DIR=`dirname "$0"`
cd $WORK_DIR

TAR_FILE="$1"

set -x
exec >$WORK_DIR/${SCRIPT_NAME}.log 2>&1
PROG_DIR=`dirname ${WORK_DIR}`

PWD=`pwd`
echo "args: $*"
TAR_FILE="$1"

${PROG_DIR}/audiostag --path "$TAR_FILE" --extract  

echo "done"
