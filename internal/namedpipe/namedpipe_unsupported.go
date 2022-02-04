// Copyright 2021 Google LLC
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

//go:build wasm
// +build wasm

package namedpipe

import (
	"fmt"
)

type Fifo struct {
	nothing bool
}

func New(bufCount, _ int) (*Fifo, error) {
	return nil, fmt.Errorf("named pipes are not supported on this OS")
}

func (f *Fifo) Path() string {
	return ""
}

func (f *Fifo) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("named pipes are not supported on this OS")
}

func (f *Fifo) Close() error {
	return fmt.Errorf("named pipes are not supported on this OS")
}
