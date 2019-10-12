#!/usr/bin/env bash
set -eux

function remove_notice {
  sed -i -e '/\/\* Copyright/,/\*\//d' "$1"

  # remove leading newline
  sed -i '/./,$!d' "$1"
}

function add_notice {
  ed "$1" <<END
0i
$2

.
w
q
END
}

gpl="/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of NAD.
 *
 * NAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * NAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with NAD.  If not, see <https://www.gnu.org/licenses/>.
 */"

agpl="/* Copyright (C) 2019 Monomax Software Pty Ltd
 *
 * This file is part of NAD.
 *
 * NAD is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * NAD is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with NAD.  If not, see <https://www.gnu.org/licenses/>.
 */"

pkgPath="$GOPATH/src/github.com/nadproject/nad/pkg"
serverPath="$GOPATH/src/github.com/nadproject/nad/pkg/server"
browserPath="$GOPATH/src/github.com/nadproject/nad/browser"

gplFiles=$(find "$pkgPath" "$browserPath" -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.tsx" -o -name "*.scss" -o -name "*.css"  \) ! -path "**/vendor/*" ! -path "**/node_modules/*" ! -path "$serverPath/*")

for file in $gplFiles; do
  remove_notice "$file"
  add_notice "$file" "$gpl"
done

webPath="$GOPATH"/src/github.com/nadproject/nad/web
jslibPath="$GOPATH/src/github.com/nadproject/nad/jslib/src"
agplFiles=$(find "$serverPath" "$webPath" "$jslibPath" -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.tsx" -o -name "*.scss" -o -name "*.css" \) ! -path "**/vendor/*" ! -path "**/node_modules/*" ! -path "**/dist/*")

for file in $agplFiles; do
  remove_notice "$file"
  add_notice "$file" "$agpl"
done
