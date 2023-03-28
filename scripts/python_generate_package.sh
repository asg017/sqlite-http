#!/bin/bash

set -eo pipefail

export PACKAGE_NAME_BASE="sqlite-http"
export EXTENSION_NAME="http0"
export VERSION=$(cat VERSION)

export LOADABLE_PATH=$1
export PYTHON_LOADABLE=$2
export OUTPUT_WHEELS=$3
export RENAME_WHEELS_ARGS=$4

cp $LOADABLE_PATH $PYTHON_LOADABLE
rm $OUTPUT_WHEELS/sqlite_http* || true
pip3 wheel python/sqlite_http/ -w $OUTPUT_WHEELS
python3 scripts/rename-wheels.py $OUTPUT_WHEELS $RENAME_WHEELS_ARGS
echo "✅ generated python wheel"

envsubst < python/version.py.tmpl > python/sqlite_http/sqlite_http/version.py
echo "✅ generated python/sqlite_http/sqlite_http/version.py"

envsubst < python/version.py.tmpl > python/datasette_sqlite_http/datasette_sqlite_http/version.py
echo "✅ generated python/datasette_sqlite_http/datasette_sqlite_http/version.py"