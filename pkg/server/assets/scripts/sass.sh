#!/usr/bin/env bash
set -eux

dir=$(dirname "${BASH_SOURCE[0]}")
serverDir="$dir/../.."
outputDir="$serverDir/static"

rm -rf "${outputDir:?}/*"

npx node-sass \
  --output-tyle compressed \
  --source-map true \
  --output "$outputDir" \
  --watch scss

# hash="$(shasum -a 256 "$outputDir")"
# hash="${hash:0:12}"
# mv "$outputDir/nad.css" "$outputDir/nad-$hash.css"
