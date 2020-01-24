#!/usr/bin/env bash
# run_server_test.sh runs server test files sequentially
# https://stackoverflow.com/questions/23715302/go-how-to-run-tests-for-multiple-packages

set -eux

dir=$(dirname "${BASH_SOURCE[0]}")
basePath="$dir/../../pkg/cli"

# clear tmp dir in case not properly torn down
rm -rf "$basePath/tmp"

# run test
pushd "$basePath"

go test -a ./... \
  -p 1\
  --tags "fts5"

popd
