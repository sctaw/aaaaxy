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

// +build !statik

package vfs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
)

var (
	localAssetDirs []string
)

// Init initializes the VFS. Must run after loading the assets.
func init() {
	localAssetDirs = []string{"assets"}
	content, err := ioutil.ReadDir("third_party")
	if err != nil {
		log.Panicf("Could not find third party directory: %v", err)
	}
	for _, info := range content {
		localAssetDirs = append(localAssetDirs, filepath.Join("third_party", info.Name(), "assets"))
	}
	log.Printf("Local asset search path: %v", localAssetDirs)
}

// load loads a file from the VFS.
func load(vfsPath string) (ReadSeekCloser, error) {
	// Note: this must be consistent with statik-vfs.sh.
	var err error
	for _, dir := range localAssetDirs {
		var r ReadSeekCloser
		r, err = os.Open(path.Join(dir, vfsPath))
		if err != nil {
			continue
		}
		return r, nil
	}
	return nil, fmt.Errorf("could not open local:%v: %w", vfsPath, err)
}

// readDir lists all files in a directory. Returns their VFS paths!
func readDir(vfsPath string) ([]string, error) {
	var results []string
	for _, dir := range localAssetDirs {
		content, err := ioutil.ReadDir(path.Join(dir, vfsPath))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("could not scan local:%v:%v: %v", vfsPath, dir, err)
			}
			continue
		}
		for _, info := range content {
			results = append(results, filepath.Join(vfsPath, info.Name()))
		}
	}
	sort.Strings(results)
	return results, nil
}
