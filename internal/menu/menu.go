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

package menu

import (
	"fmt"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/divVerent/aaaaxy/internal/engine"
	"github.com/divVerent/aaaaxy/internal/exitstatus"
	"github.com/divVerent/aaaaxy/internal/flag"
	_ "github.com/divVerent/aaaaxy/internal/game" // Load entities.
	"github.com/divVerent/aaaaxy/internal/input"
	"github.com/divVerent/aaaaxy/internal/log"
	"github.com/divVerent/aaaaxy/internal/music"
	"github.com/divVerent/aaaaxy/internal/offscreen"
	"github.com/divVerent/aaaaxy/internal/playerstate"
	"github.com/divVerent/aaaaxy/internal/sound"
	"github.com/divVerent/aaaaxy/internal/timing"
)

var (
	saveState = flag.Int("save_state", 0, "number of save state slot")
)

const (
	blurSize     = 1
	blurFrames   = 32
	darkenFactor = 0.75
)

type MenuScreen interface {
	Init(m *Controller) error
	Update() error
	Draw(screen *ebiten.Image)
}

type Controller struct {
	initialized   bool
	World         engine.World
	Screen        MenuScreen
	moveSound     *sound.Sound
	activateSound *sound.Sound
	blurFrame     int
	needReloadMap bool
}

func (c *Controller) Update() error {
	defer timing.Group()()

	timing.Section("once")
	if !c.initialized {
		err := c.InitGame(loadGame)
		if err != nil {
			return err
		}
		c.moveSound, err = sound.Load("menu_move.ogg")
		if err != nil {
			return err
		}
		c.activateSound, err = sound.Load("menu_activate.ogg")
		if err != nil {
			return err
		}
		input.CancelHover()
		c.initialized = true
	}

	timing.Section("global_hotkeys")

	input.SetHideControls(c.World.TimerStopped)
	if c.World.ForceCredits {
		c.World.ForceCredits = false
		c.blurFrame = 0
		return c.SwitchToScreen(&CreditsScreen{Fancy: true})
	} else if input.Exit.JustHit && c.Screen == nil && !c.World.TimerStopped {
		if c.World.PlayerState.LastCheckpoint() != "" || c.World.PlayerState.Frames() > 0 {
			c.World.TimerStarted = true
		}
		music.Switch("")
		if c.World.TimerStarted {
			c.World.PlayerState.AddEscape()
		}
		c.World.PreDespawn()
		c.blurFrame = 0
		return c.SwitchToScreen(&MainScreen{})
	}
	if input.Fullscreen.JustHit {
		c.toggleFullscreen()
	}

	timing.Section("screen")
	if c.Screen != nil {
		if c.blurFrame < blurFrames {
			c.blurFrame++
		}
		input.SetWantClicks(true)
		err := c.Screen.Update()
		if err != nil {
			return err
		}
	} else {
		c.blurFrame = 0
		input.SetWantClicks(false)
	}

	performQualityAdjustment()

	return nil
}

func (c *Controller) toggleFullscreen() error {
	fs := !ebiten.IsFullscreen()
	flag.Set("fullscreen", fs)
	ebiten.SetFullscreen(fs)
	input.CancelHover() // Fullscreen toggle changes mouse position; ignore hover events for that.
	return nil
}

func (c *Controller) UpdateWorld() error {
	// Increment the frame counter.
	// Except when on the credits screen - that time does not count.
	if c.World.TimerStarted && !c.World.TimerStopped {
		c.World.PlayerState.AddFrame()
	}

	if c.Screen != nil {
		// Game is paused while in menu.
		return nil
	}
	return c.World.Update()
}

func (c *Controller) Draw(screen *ebiten.Image) {
	defer timing.Group()()

	timing.Section("screen")
	if c.Screen != nil {
		c.Screen.Draw(screen)
	}
}

func (c *Controller) DrawWorld(screen *ebiten.Image) {
	f := float64(c.blurFrame) / blurFrames

	dest := screen
	if offscreen.AvoidReuse() && f != 0 {
		dest = offscreen.New("GameUnblurred", engine.GameWidth, engine.GameHeight)
	}

	c.World.Draw(dest, f)

	if f != 0 {
		// If a menu screen is active, just draw the previous saved bitmap, but blur it.
		darken := darkenFactor*f + 1.0*(1-f)
		engine.BlurImage("BlurGame", dest, screen, blurSize, darken, 0.0, f)
		if offscreen.AvoidReuse() {
			offscreen.Dispose(dest)
		}
	}
}

type resetFlag int

const (
	loadGame resetFlag = iota
	resetGame
)

// initGame reinitializes just the game.
func (c *Controller) initGame(f resetFlag) error {
	// Stop the timer.
	c.World.TimerStarted = false

	// Initialize the world.
	err := c.World.Init(*saveState)
	if err != nil {
		return fmt.Errorf("could not initialize world: %w", err)
	}

	// Load the saved state.
	if f == loadGame {
		err := c.World.Load()
		if err != nil {
			return err
		}
	}

	// Now we're good.
	c.needReloadMap = false

	return nil
}

// InitGame is called by menu screens to load/reset the game.
func (c *Controller) InitGame(f resetFlag) error {
	err := c.initGame(f)
	if err != nil {
		return err
	}

	// Go to the game screen.
	c.Screen = nil
	return nil
}

// SwitchSaveState switches to a given save state.
func (c *Controller) SwitchSaveState(state int) error {
	// Save the game first.
	err := c.World.Save()
	if err != nil {
		return fmt.Errorf("could not save game: %w", err)
	}

	// Now select the new state.
	*saveState = state
	err = engine.SaveConfig()
	if err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}

	// And finally restart the game from there.
	return c.InitGame(loadGame)
}

// SwitchToGame switches to the game without teleporting.
func (c *Controller) SwitchToGame() error {
	if c.needReloadMap {
		err := c.initGame(loadGame)
		if err != nil {
			return err
		}
	}
	c.Screen = nil
	return nil
}

// SwitchToCheckpoint switches to a specific checkpoint.
func (c *Controller) SwitchToCheckpoint(cp string) error {
	if c.needReloadMap {
		err := c.initGame(loadGame)
		if err != nil {
			return err
		}
	}
	if cp != c.World.PlayerState.LastCheckpoint() {
		c.World.PlayerState.AddTeleport()
	}
	err := c.World.RespawnPlayer(cp, true)
	if err != nil {
		return fmt.Errorf("could not respawn player: %w", err)
	}
	c.World.TimerStarted = true
	c.Screen = nil
	return nil
}

// SwitchToScreen is called by menu screens to go to a different menu screen.
func (c *Controller) SwitchToScreen(screen MenuScreen) error {
	c.Screen = screen
	return c.Screen.Init(c)
}

// SaveConfigAndSwitchToScreen is called by menu screens to go to a different menu screen.
func (c *Controller) SaveConfigAndSwitchToScreen(screen MenuScreen) error {
	err := engine.SaveConfig()
	if err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}
	c.Screen = screen
	return c.Screen.Init(c)
}

// QuitGame is called by menu screens to end the game.
func (c *Controller) QuitGame() error {
	categories, _ := (c.World.PlayerState.SpeedrunCategories() | playerstate.AnyPercentSpeedrun).Describe()
	log.Infof("on track for %v", categories)
	err := c.World.Save()
	if err != nil {
		return fmt.Errorf("could not save game: %w", err)
	}
	err = engine.SaveConfig()
	if err != nil {
		return fmt.Errorf("could not save config: %w", err)
	}
	return exitstatus.RegularTermination
}

// ActivateSound plays the sound effect to activate something.
func (c *Controller) ActivateSound(err error) error {
	if err == nil {
		c.activateSound.Play()
	}
	return err
}

// MoveSound plays the sound effect to activate something.
func (c *Controller) MoveSound(err error) error {
	if err == nil {
		c.moveSound.Play()
	}
	return err
}

func (c *Controller) QueryMouseItem(item interface{}, count int) Direction {
	mousePos, mouseState := input.Mouse()
	if mouseState == input.NoMouse {
		return NotClicked
	}
	if idx, dir := ItemClicked(mousePos, count); dir != NotClicked {
		v := reflect.ValueOf(item).Elem()
		prev := v.Int()
		if int64(idx) != prev {
			v.SetInt(int64(idx))
			c.MoveSound(nil)
		}
		if mouseState == input.ClickingMouse {
			return dir
		}
	}
	return NotClicked
}

func (c *Controller) PaletteChanged() error {
	// Reinitialize world when going back to game so palette change applies
	// fully. While under menu blur, some stuff will be slightly glitchy (e.g.
	// gradient), but that's better than black screen.
	c.needReloadMap = true
	return nil
}
