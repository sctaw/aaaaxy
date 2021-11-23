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

package aaaaxy

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/divVerent/aaaaxy/internal/audiowrap"
	"github.com/divVerent/aaaaxy/internal/credits"
	"github.com/divVerent/aaaaxy/internal/demo"
	"github.com/divVerent/aaaaxy/internal/engine"
	"github.com/divVerent/aaaaxy/internal/flag"
	"github.com/divVerent/aaaaxy/internal/font"
	"github.com/divVerent/aaaaxy/internal/image"
	"github.com/divVerent/aaaaxy/internal/input"
	"github.com/divVerent/aaaaxy/internal/log"
	m "github.com/divVerent/aaaaxy/internal/math"
	"github.com/divVerent/aaaaxy/internal/noise"
	"github.com/divVerent/aaaaxy/internal/sound"
	"github.com/divVerent/aaaaxy/internal/timing"
	"github.com/divVerent/aaaaxy/internal/version"
	"github.com/divVerent/aaaaxy/internal/vfs"
)

var (
	vsync                 = flag.Bool("vsync", true, "enable waiting for vertical synchronization")
	fullscreen            = flag.Bool("fullscreen", true, "enable fullscreen mode")
	windowScaleFactor     = flag.Float64("window_scale_factor", 0, "window scale factor in device pixels per game pixel (0 means auto integer scaling)")
	runnableWhenUnfocused = flag.Bool("runnable_when_unfocused", false, "keep running the game even when not focused")
)

func LoadConfig() (*flag.Config, error) {
	return engine.LoadConfig()
}

func setWindowSize() {
	logicalF := *windowScaleFactor
	log.Infof("requested logical scale factor: %v", logicalF)
	dscale := ebiten.DeviceScaleFactor()
	log.Infof("device scale factor: %v", dscale)
	var physicalF float64
	if logicalF <= 0 {
		screenw, screenh := ebiten.ScreenSizeInFullscreen()
		log.Infof("screen size: %vx%v", screenw, screenh)
		// Reserve 128 device independent pixels for system controls.
		maxw, maxh := screenw-128, screenh-128
		log.Infof("max size: %vx%v", maxw, maxh)
		// Compute max scaling factors.
		maxwf, maxhf := float64(maxw)*dscale/engine.GameWidth, float64(maxh)*dscale/engine.GameHeight
		log.Infof("max physical scale factors: %v, %v", maxwf, maxhf)
		physicalF = math.Min(maxwf, maxhf)
	} else {
		physicalF = logicalF * dscale
	}
	log.Infof("requested physical scale factor: %v", physicalF)
	// Make output pixels an integer multiple of input pixels (looks better).
	physicalF = math.Floor(physicalF)
	if physicalF < 1 {
		physicalF = 1
	}
	log.Infof("chosen physical scale factor: %v", physicalF)
	// Convert back to logical scale factor as ebiten needs that.
	logicalF = physicalF / dscale
	log.Infof("chosen logical pixel scale factor: %v", logicalF)
	w, h := m.Rint(engine.GameWidth*logicalF), m.Rint(engine.GameHeight*logicalF)
	log.Infof("chosen window size: %vx%v", w, h)
	ebiten.SetWindowSize(w, h)
}

func InitEbiten() error {
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	ebiten.SetFullscreen(*fullscreen)
	ebiten.SetInitFocused(true)
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetScreenTransparent(false)
	ebiten.SetVsyncEnabled(*vsync)
	ebiten.SetWindowDecorated(true)
	ebiten.SetWindowResizable(true)
	setWindowSize()
	ebiten.SetWindowTitle("AAAAXY")

	err := vfs.Init()
	if err != nil {
		return fmt.Errorf("could not initialize VFS: %v", err)
	}
	err = version.Init()
	if err != nil {
		return fmt.Errorf("could not initialize version: %v", err)
	}
	err = input.Init()
	if err != nil {
		return fmt.Errorf("could not initialize input: %v", err)
	}
	err = font.Init()
	if err != nil {
		return fmt.Errorf("could not initialize fonts: %v", err)
	}
	err = credits.Precache()
	if err != nil {
		return fmt.Errorf("could not precache credits: %v", err)
	}
	err = image.Precache()
	if err != nil {
		return fmt.Errorf("could not precache images: %v", err)
	}
	err = audiowrap.Init()
	if err != nil {
		return fmt.Errorf("could not initialize audio: %v", err)
	}
	err = sound.Precache()
	if err != nil {
		return fmt.Errorf("could not precache sounds: %v", err)
	}
	err = engine.Precache()
	if err != nil {
		return fmt.Errorf("could not precache engine: %v", err)
	}
	err = noise.Init()
	if err != nil {
		return fmt.Errorf("could not initialize noise: %v", err)
	}
	err = demo.Init()
	if err != nil {
		return fmt.Errorf("could not initialize demo: %v", err)
	}
	err = initDumping()
	if err != nil {
		return fmt.Errorf("could not initialize dumping: %v", err)
	}

	if slowDumping() || demo.Timedemo() {
		ebiten.SetMaxTPS(ebiten.UncappedTPS)
	} else {
		ebiten.SetMaxTPS(engine.GameTPS)
	}

	// Pause when unfocused, except when recording demos.
	ebiten.SetRunnableOnUnfocused(*runnableWhenUnfocused || (demo.Playing() && dumping()))

	return nil
}

func BeforeExit() error {
	timing.PrintReport()
	err := finishDumping()
	if err != nil {
		return fmt.Errorf("could not finish dumping: %v", err)
	}
	err = demo.BeforeExit()
	if err != nil {
		return fmt.Errorf("could not finalize demo: %v", err)
	}
	return nil
}
