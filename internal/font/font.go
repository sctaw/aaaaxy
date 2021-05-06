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

package font

import (
	"fmt"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/gofont/gosmallcaps"

	"github.com/divVerent/aaaaaa/internal/flag"
)

var (
	pinFontsToCache       = flag.Bool("pin_fonts_to_cache", true, "Pin all fonts to glyph cache.")
	pinFontsToCacheHarder = flag.Bool("pin_fonts_to_cache_harder", false, "Do a dummy draw command to pin fonts to glyph cache harder.")
)

// Face is an alias to font.Face so users do not need to import the font package.
type Face struct {
	font.Face
}

func makeFace(f font.Face) Face {
	face := Face{Face: f}
	all = append(all, face)
	return face
}

// cacheChars are all characters the game uses. ASCII plus all Unicode our map file contains.
var cacheChars = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~τπö¾"

// We always keep the game character set in cache.
// This has to be repeated regularly as ebiten expires unused cache entries.
func KeepInCache(dst *ebiten.Image) {
	if *pinFontsToCacheHarder {
		for _, f := range all {
			f.precache(dst, cacheChars)
		}
	}
	if *pinFontsToCache {
		for _, f := range all {
			f.recache(cacheChars)
		}
	}
}

var (
	all            = []Face{}
	ByName         = map[string]Face{}
	Centerprint    Face
	CenterprintBig Face
	DebugSmall     Face
	MenuBig        Face
	Menu           Face
	MenuSmall      Face
)

func Init() error {
	// Load the fonts.
	regular, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return fmt.Errorf("could not load goitalic font: %v", err)
	}
	italic, err := truetype.Parse(goitalic.TTF)
	if err != nil {
		return fmt.Errorf("could not load goitalic font: %v", err)
	}
	mono, err := truetype.Parse(gomono.TTF)
	if err != nil {
		return fmt.Errorf("could not load gomono font: %v", err)
	}
	smallcaps, err := truetype.Parse(gosmallcaps.TTF)
	if err != nil {
		return fmt.Errorf("could not load gosmallcaps font: %v", err)
	}

	ByName["Small"] = makeFace(truetype.NewFace(regular, &truetype.Options{
		Size:       10,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	ByName["Regular"] = makeFace(truetype.NewFace(regular, &truetype.Options{
		Size:       16,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	ByName["Italic"] = makeFace(truetype.NewFace(italic, &truetype.Options{
		Size:       16,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	ByName["Mono"] = makeFace(truetype.NewFace(mono, &truetype.Options{
		Size:       16,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	ByName["SmallCaps"] = makeFace(truetype.NewFace(smallcaps, &truetype.Options{
		Size:       16,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	Centerprint = makeFace(truetype.NewFace(italic, &truetype.Options{
		Size:       16,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	CenterprintBig = makeFace(truetype.NewFace(smallcaps, &truetype.Options{
		Size:       24,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	DebugSmall = makeFace(truetype.NewFace(mono, &truetype.Options{
		Size:       5,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	Menu = makeFace(truetype.NewFace(smallcaps, &truetype.Options{
		Size:       16,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	MenuBig = makeFace(truetype.NewFace(smallcaps, &truetype.Options{
		Size:       24,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))
	MenuSmall = makeFace(truetype.NewFace(smallcaps, &truetype.Options{
		Size:       12,
		Hinting:    font.HintingFull,
		SubPixelsX: 1,
		SubPixelsY: 1,
	}))

	return nil
}
