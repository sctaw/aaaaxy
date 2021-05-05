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

package main

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/divVerent/aaaaaa/internal/aaaaaa"
	"github.com/divVerent/aaaaaa/internal/flag"
)

var (
	cpuprofile = flag.String("debug_cpuprofile", "", "write CPU profile to file")
	memprofile = flag.String("debug_memprofile", "", "write memory profile to file")
)

func main() {
	flag.Parse(aaaaaa.LoadConfig)
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("could not create CPU profile: %v", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("could not start CPU profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}
	err := aaaaaa.InitEbiten()
	if err != nil {
		log.Panicf("could not initialize game: %v", err)
	}
	game := &aaaaaa.Game{}
	err = ebiten.RunGame(game)
	aaaaaa.BeforeExit()
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatalf("could not create memory profile: %v", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("could not write memory profile: %v", err)
		}
	}
	if err != nil && err != aaaaaa.RegularTermination {
		log.Fatal(err)
	}
}
