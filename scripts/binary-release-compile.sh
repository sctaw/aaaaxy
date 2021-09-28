#!/bin/sh
# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -ex

: ${GO:=go}

GOOS=$($GO env GOOS)
GOARCH=$($GO env GOARCH)
GOEXE=$($GO env GOEXE)
zip="$PWD/aaaaxy-$GOOS-$GOARCH-$(scripts/version.sh gittag).zip"

exec 3>&1
exec >&2

case "$GOOS" in
	darwin)
		appdir=packaging/
		app=AAAAXY.app
		prefix=packaging/AAAAXY.app/Contents/MacOS/
		;;
	js)
		appdir=.
		app="aaaaxy-$GOOS-$GOARCH$GOEXE aaaaxy.html wasm_exec.js"
		prefix=
		;;
	*)
		appdir=.
		app=aaaaxy-$GOOS-$GOARCH$GOEXE
		prefix=
		;;
esac

make clean
make BUILDTYPE=release PREFIX="$prefix"

rm -f "$zip"
7za a -tzip -mx=9 "$zip" \
	README.md LICENSE CONTRIBUTING.md \
	licenses
(
	cd "$appdir"
	7za a -tzip -mx=9 "$zip" \
		$app
)

make clean

echo >&3 "$zip"
