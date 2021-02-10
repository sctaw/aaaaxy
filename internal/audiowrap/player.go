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

package audiowrap

import (
	"bytes"
	"io"
	"time"

	"github.com/divVerent/aaaaaa/internal/flag"
	ebiaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

var (
	volume = flag.Float64("volume", 1.0, "global volume (0..1)")
	// TODO: add a way to simulate audio and write to disk, syncing with the frame clock (i.e. each frame renders exactly 1/60 sec of audio).
	// Also a way to don't actually render audio (but still advance clock) would be nice.
)

type Player struct {
	ebi *ebiaudio.Player
	dmp *dumper
}

func NewPlayer(src io.Reader) (*Player, error) {
	dmp, src := newDumperWithTee(src)
	ebi, err := ebiaudio.NewPlayer(ebiaudio.CurrentContext(), src)
	if err != nil {
		return nil, err
	}
	return &Player{
		ebi: ebi,
		dmp: dmp,
	}, nil
}

func NewPlayerFromBytes(src []byte) *Player {
	dmp := newDumper(bytes.NewReader(src))
	ebi := ebiaudio.NewPlayerFromBytes(ebiaudio.CurrentContext(), src)
	return &Player{
		ebi: ebi,
		dmp: dmp,
	}
}

func (p *Player) Close() error {
	if p.dmp != nil {
		p.dmp.Close()
	}
	return p.ebi.Close()
}

func (p *Player) Current() time.Duration {
	if p.dmp != nil {
		return p.dmp.Current()
	}
	return p.ebi.Current()
}

func (p *Player) IsPlaying() bool {
	if p.dmp != nil {
		return p.dmp.IsPlaying()
	}
	return p.ebi.IsPlaying()
}

func (p *Player) Pause() {
	if p.dmp != nil {
		p.dmp.Pause()
	}
	p.ebi.Pause()
}

func (p *Player) Play() {
	if p.dmp != nil {
		p.dmp.Play()
	}
	p.ebi.Play()
}

func (p *Player) SetVolume(vol float64) {
	if p.dmp != nil {
		p.dmp.SetVolume(vol * *volume)
	}
	p.ebi.SetVolume(vol * *volume)
}
