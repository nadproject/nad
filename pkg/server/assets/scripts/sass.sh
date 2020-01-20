#!/usr/bin/env bash
set -eux

dir=$(dirname "${BASH_SOURCE[0]}")
serverDir="$dir/../.."
outputDir="$serverDir/static"
inputDir="$dir/../scss"

rm -rf "${outputDir:?}/*"

task="npx node-sass \
  --output-tyle compressed \
  --source-map true \
  --output $outputDir"

# compile first then watch
eval "$task $inputDir"
eval "$task --watch $inputDir"

# hash="$(shasum -a 256 "$outputDir")"
# hash="${hash:0:12}"
# mv "$outputDir/nad.css" "$outputDir/nad-$hash.css"
