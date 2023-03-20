// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build ios
// +build ios

package vfs

import (
	"fmt"
	"path/filepath"

	"github.com/divVerent/aaaaxy/internal/log"
)

func pathForReadRaw(kind StateKind, name string) (string, error) {
	return pathForWrite(kind, name)
}

func pathForWriteRaw(kind StateKind, name string) (string, error) {
	switch kind {
	case Config:
		return "", fmt.Errorf("NOT YET IMPLEMENTED: %d", kind)
	case SavedGames:
		return "", fmt.Errorf("NOT YET IMPLEMENTED: %d", kind)
	default:
		return "", fmt.Errorf("searched for unsupported state kind: %d", kind)
	}
}
