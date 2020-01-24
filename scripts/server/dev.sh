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

# run asset pipeline in the background
(cd "$serverPath/assets/" && "$serverPath/assets/scripts/styles/dev.sh" ) &

# run server
moduleName="github.com/nadproject/nad"
ldflags="-X '$moduleName/pkg/server/buildinfo.CSSFiles=nad.css' -X '$moduleName/pkg/server/buildinfo.JSFiles=nad.js' -X '$moduleName/pkg/server/buildinfo.Version=dev' "
task="go run -ldflags \"$ldflags\" main.go start"

(
  cd "$basePath/pkg/watcher" && \
  go run main.go \
    --task="$task" \
    --context="$serverPath" \
    --ignore="$serverPath/assets/node_modules" \
    "$serverPath" \
)
