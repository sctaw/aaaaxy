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

package input

import (
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"

	m "github.com/divVerent/aaaaxy/internal/math"
)

type ImpulseState struct {
	Held    bool `json:",omitempty"`
	JustHit bool `json:",omitempty"`
}

func (i *ImpulseState) Empty() bool {
	return !i.Held && !i.JustHit
}

func (i *ImpulseState) OrEmpty() ImpulseState {
	if i == nil {
		return ImpulseState{}
	}
	return *i
}

func (i *ImpulseState) UnlessEmpty() *ImpulseState {
	if i.Empty() {
		return nil
	}
	return i
}

type InputMap int

func (i InputMap) ContainsAny(o InputMap) bool {
	return i&o != 0
}

type impulse struct {
	ImpulseState
	Name string

	keys        map[ebiten.Key]InputMap
	padControls padControls
}

const (
	NoInput InputMap = 0

	// Allocated input bits.
	DOSKeyboardWithEscape    InputMap = 1
	NESKeyboardWithEscape    InputMap = 2
	FPSKeyboardWithEscape    InputMap = 4
	ViKeyboardWithEscape     InputMap = 8
	Gamepad                  InputMap = 16
	DOSKeyboardWithBackspace InputMap = 32
	NESKeyboardWithBackspace InputMap = 64
	FPSKeyboardWithBackspace InputMap = 128
	ViKeyboardWithBackspace  InputMap = 256

	// Computed helpers values.
	AnyKeyboardWithEscape    = DOSKeyboardWithEscape | NESKeyboardWithEscape | FPSKeyboardWithEscape | ViKeyboardWithEscape
	AnyKeyboardWithBackspace = DOSKeyboardWithBackspace | NESKeyboardWithBackspace | FPSKeyboardWithBackspace | ViKeyboardWithBackspace
	DOSKeyboard              = DOSKeyboardWithEscape | DOSKeyboardWithBackspace
	NESKeyboard              = NESKeyboardWithEscape | NESKeyboardWithBackspace
	FPSKeyboard              = FPSKeyboardWithEscape | FPSKeyboardWithBackspace
	ViKeyboard               = ViKeyboardWithEscape | ViKeyboardWithBackspace
	AnyKeyboard              = AnyKeyboardWithEscape | AnyKeyboardWithBackspace
	AnyInput                 = AnyKeyboard | Gamepad
)

var (
	Left       = (&impulse{Name: "Left", keys: leftKeys, padControls: leftPad}).register()
	Right      = (&impulse{Name: "Right", keys: rightKeys, padControls: rightPad}).register()
	Up         = (&impulse{Name: "Up", keys: upKeys, padControls: upPad}).register()
	Down       = (&impulse{Name: "Down", keys: downKeys, padControls: downPad}).register()
	Jump       = (&impulse{Name: "Jump", keys: jumpKeys, padControls: jumpPad}).register()
	Action     = (&impulse{Name: "Action", keys: actionKeys, padControls: actionPad}).register()
	Exit       = (&impulse{Name: "Exit", keys: exitKeys, padControls: exitPad}).register()
	Fullscreen = (&impulse{Name: "Fullscreen", keys: fullscreenKeys /* no padControls */}).register()

	impulses = []*impulse{}

	inputMap InputMap

	// Wait for first frame to detect initial gamepad situation.
	firstUpdate = true

	// Current mouse/finger hover pos, if any.
	hoverPos *m.Pos

	// Last mouse/finger click/release pos, if any.
	clickPos *m.Pos
)

func (i *impulse) register() *impulse {
	impulses = append(impulses, i)
	return i
}

func (i *impulse) update() {
	keyboardHeld := i.keyboardPressed()
	gamepadHeld := i.gamepadPressed()
	held := keyboardHeld | gamepadHeld
	if held != 0 && !i.Held {
		i.JustHit = true
		// Whenever a new key is pressed, update the flag whether we're actually
		// _using_ the gamepad. Used for some in-game text messages.
		inputMap &= held
		if inputMap == NoInput {
			inputMap = held
		}
		// Hide mouse pointer if using another input device in the menu.
		mouseCancel()
	} else {
		i.JustHit = false
	}
	i.Held = held != 0
}

func Init() error {
	gamepadInit()
	return nil
}

func Update(screenWidth, screenHeight, gameWidth, gameHeight int) {
	gamepadScan()
	if firstUpdate {
		// At first, assume gamepad whenever one is present.
		if len(gamepads) > 0 {
			inputMap = Gamepad
		} else {
			inputMap = AnyKeyboard
		}
		firstUpdate = false
	}
	for _, i := range impulses {
		i.update()
	}
	clickPos, hoverPos = nil, nil
	mouseUpdate(screenWidth, screenHeight, gameWidth, gameHeight)
	easterEggUpdate()
}

func SetWantClicks(want bool) {
	mouseSetWantClicks(want)
}

func EasterEggJustHit() bool {
	return easterEgg.justHit || snesEasterEgg.justHit
}

func KonamiCodeJustHit() bool {
	return konamiCode.justHit || snesKonamiCode.justHit || kbdKonamiCode.justHit
}

func Map() InputMap {
	return inputMap
}

type ExitButtonID int

const (
	Escape ExitButtonID = iota
	Backspace
	Start
)

func ExitButton() ExitButtonID {
	if inputMap.ContainsAny(Gamepad) {
		return Start
	}
	if runtime.GOOS != "js" {
		// On JS, the Esc key is kinda "reserved" for leaving fullsreeen.
		// Thus we never recommend it, even if the user used it before.
		if inputMap.ContainsAny(AnyKeyboardWithEscape) {
			return Escape
		}
	}
	return Backspace
}

func HoverPos() (m.Pos, bool) {
	if hoverPos == nil {
		return m.Pos{}, false
	}
	return *hoverPos, true
}

func ClickPos() (m.Pos, bool) {
	if clickPos == nil {
		return m.Pos{}, false
	}
	return *clickPos, true
}

type MouseStatus int

const (
	NoMouse MouseStatus = iota
	HoveringMouse
	ClickingMouse
)

func Mouse() (m.Pos, MouseStatus) {
	if clickPos != nil {
		return *clickPos, ClickingMouse
	}
	if hoverPos != nil {
		return *hoverPos, HoveringMouse
	}
	return m.Pos{}, NoMouse
}

// Demo code.

type DemoState struct {
	InputMap          InputMap      `json:",omitempty"`
	Left              *ImpulseState `json:",omitempty"`
	Right             *ImpulseState `json:",omitempty"`
	Up                *ImpulseState `json:",omitempty"`
	Down              *ImpulseState `json:",omitempty"`
	Jump              *ImpulseState `json:",omitempty"`
	Action            *ImpulseState `json:",omitempty"`
	Exit              *ImpulseState `json:",omitempty"`
	HoverPos          *m.Pos        `json:",omitempty"`
	ClickPos          *m.Pos        `json:",omitempty"`
	EasterEggJustHit  bool          `json:",omitempty"`
	KonamiCodeJustHit bool          `json:",omitempty"`
}

func LoadFromDemo(state *DemoState) {
	if state == nil {
		state = &DemoState{}
	}
	inputMap = state.InputMap
	Left.ImpulseState = state.Left.OrEmpty()
	Right.ImpulseState = state.Right.OrEmpty()
	Up.ImpulseState = state.Up.OrEmpty()
	Down.ImpulseState = state.Down.OrEmpty()
	Jump.ImpulseState = state.Jump.OrEmpty()
	Action.ImpulseState = state.Action.OrEmpty()
	Exit.ImpulseState = state.Exit.OrEmpty()
	hoverPos = state.HoverPos
	clickPos = state.ClickPos
	easterEgg.justHit = state.EasterEggJustHit
	snesEasterEgg.justHit = state.EasterEggJustHit
	konamiCode.justHit = state.KonamiCodeJustHit
	snesKonamiCode.justHit = state.KonamiCodeJustHit
	kbdKonamiCode.justHit = state.KonamiCodeJustHit
}

func SaveToDemo() *DemoState {
	return &DemoState{
		InputMap:          inputMap,
		Left:              Left.ImpulseState.UnlessEmpty(),
		Right:             Right.ImpulseState.UnlessEmpty(),
		Up:                Up.ImpulseState.UnlessEmpty(),
		Down:              Down.ImpulseState.UnlessEmpty(),
		Jump:              Jump.ImpulseState.UnlessEmpty(),
		Action:            Action.ImpulseState.UnlessEmpty(),
		Exit:              Exit.ImpulseState.UnlessEmpty(),
		HoverPos:          hoverPos,
		ClickPos:          clickPos,
		EasterEggJustHit:  EasterEggJustHit(),
		KonamiCodeJustHit: KonamiCodeJustHit(),
	}
}
