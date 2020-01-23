#!/usr/bin/env bash
# test.sh runs server tests. It is to be invoked by other scripts that set
# appropriate env vars.
set -eux

dir=$(dirname "${BASH_SOURCE[0]}")
pushd "$dir/../../pkg/server"

# export NAD_TEST_EMAIL_TEMPLATE_DIR="$dir/../../pkg/server/mailer/templates/src"

function run_test {
  # go test ./... -cover -p 1 
  go test ./controllers -cover -p 1 
}

if [ "${WATCH-false}" == true ]; then
  set +e
  while inotifywait --exclude .swp -e modify -r .; do run_test; done;
  set -e
else
  run_test
fi

popd
