// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// distributed under the License is distributed on an "AS IS" BASIS,
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// you may not use this file except in compliance with the License.
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package menu

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/divVerent/aaaaaa/internal/engine"
	"github.com/divVerent/aaaaaa/internal/flag"
	"github.com/divVerent/aaaaaa/internal/font"
	"github.com/divVerent/aaaaaa/internal/input"
	m "github.com/divVerent/aaaaaa/internal/math"
)

type ResetScreenItem int

const (
	ResetNothing ResetScreenItem = iota
	ResetConfig
	ResetGame
	BackToMain
	ResetCount
)

const resetFrames = 300

type ResetScreen struct {
	Menu       *Menu
	Item       ResetScreenItem
	ResetFrame int
}

func (s *ResetScreen) Init(m *Menu) error {
	s.Menu = m
	return nil
}

func (s *ResetScreen) Update() error {
	if s.Item == ResetGame {
		s.ResetFrame++
	} else {
		s.ResetFrame = 0
	}
	if input.Down.JustHit {
		s.Item++
		s.Menu.MoveSound(nil)
	}
	if input.Up.JustHit {
		s.Item--
		s.Menu.MoveSound(nil)
	}
	s.Item = ResetScreenItem(m.Mod(int(s.Item), int(ResetCount)))
	if input.Exit.JustHit {
		return s.Menu.ActivateSound(s.Menu.SwitchToScreen(&SettingsScreen{}))
	}
	if input.Jump.JustHit || input.Action.JustHit {
		switch s.Item {
		case ResetNothing:
			return s.Menu.ActivateSound(s.Menu.SwitchToScreen(&SettingsScreen{}))
		case ResetConfig:
			flag.ResetToDefaults()
			return s.Menu.ActivateSound(s.Menu.SwitchToScreen(&SettingsScreen{}))
		case ResetGame:
			if s.ResetFrame >= resetFrames {
				return s.Menu.ActivateSound(s.Menu.InitGame(resetGame))
			}
		case BackToMain:
			return s.Menu.ActivateSound(s.Menu.SwitchToScreen(&MainScreen{}))
		}
	}
	return nil
}

func (s *ResetScreen) Draw(screen *ebiten.Image) {
	h := engine.GameHeight
	x := engine.GameWidth / 2
	fgs := color.NRGBA{R: 255, G: 255, B: 85, A: 255}
	bgs := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	fgn := color.NRGBA{R: 170, G: 170, B: 170, A: 255}
	bgn := color.NRGBA{R: 85, G: 85, B: 85, A: 255}
	font.MenuBig.Draw(screen, "Reset", m.Pos{X: x, Y: h / 4}, true, fgs, bgs)
	fg, bg := fgn, bgn
	if s.Item == ResetNothing {
		fg, bg = fgs, bgs
	}
	font.Menu.Draw(screen, "Reset Nothing", m.Pos{X: x, Y: 21 * h / 32}, true, fg, bg)
	fg, bg = fgn, bgn
	if s.Item == ResetConfig {
		fg, bg = fgs, bgs
	}
	font.Menu.Draw(screen, "Reset and Lose Settings", m.Pos{X: x, Y: 23 * h / 32}, true, fg, bg)
	var resetText string
	var dx, dy int
	if s.ResetFrame >= resetFrames && s.Item == ResetGame {
		fg, bg = color.NRGBA{R: 170, G: 0, B: 0, A: 255}, color.NRGBA{R: 0, G: 0, B: 0, A: 255}
		resetText = "Reset and Lose SAVED GAME"
	} else {
		fg, bg = fgn, bgn
		if s.Item == ResetGame {
			fg, bg = color.NRGBA{R: 255, G: 85, B: 85, A: 255}, color.NRGBA{R: 170, G: 0, B: 0, A: 255}
		}
		if s.Item == ResetGame {
			resetText = fmt.Sprintf("Reset and Lose Saved Game (think about it for %d sec)", (resetFrames-s.ResetFrame+engine.GameTPS-1)/engine.GameTPS)
		} else {
			resetText = "Reset and Lose Saved Game"
		}
		dx = rand.Intn(3) - 1
		dy = rand.Intn(3) - 1
	}
	font.Menu.Draw(screen, resetText, m.Pos{X: x + dx, Y: 25*h/32 + dy}, true, fg, bg)
	fg, bg = fgn, bgn
	if s.Item == BackToMain {
		fg, bg = fgs, bgs
	}
	font.Menu.Draw(screen, "Main Menu", m.Pos{X: x, Y: 27 * h / 32}, true, fg, bg)
}
