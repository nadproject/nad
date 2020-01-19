#!/usr/bin/env bash
# shellcheck disable=SC1090
# dev.sh builds and starts development environment
set -eux -o pipefail

# clean up background processes
dir=$(dirname "${BASH_SOURCE[0]}")
basePath="$dir/../.."
serverPath=$(realpath "$basePath/pkg/server")

# load env
set -a
dotenvPath="$serverPath/.env.dev"
source "$dotenvPath"
set +a

# run server
task="go run -ldflags \"-X github.com/nadproject/nad/pkg/server/build.CSSFiles=watt,wat\" main.go start"


#cd "$serverPath" && eval "${task}"
(
  cd "$basePath/pkg/watcher" && \
  go run main.go \
    --task="$task" \
    --context="$serverPath" \
    "$serverPath" \
)
