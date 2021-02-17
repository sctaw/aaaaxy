// Copyright 2021 Google LLC
//
// Licensed under the Apache License, SaveGameVersion 2.0 (the "License");
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

package level

import (
	"fmt"
	"log"
	"strings"

	"github.com/fardog/tmx"
	"github.com/mitchellh/hashstructure/v2"

	m "github.com/divVerent/aaaaaa/internal/math"
	"github.com/divVerent/aaaaaa/internal/vfs"
)

// Level is a parsed form of a loaded level.
type Level struct {
	Player              *Spawnable
	Checkpoints         map[string]*Spawnable
	CheckpointLocations *CheckpointLocations
	SaveGameVersion     int
	Hash                uint64

	tiles []LevelTile
	width int
}

// Tile returns the tile at the given position.
func (l *Level) Tile(pos m.Pos) *LevelTile {
	i := pos.X + pos.Y*l.width
	t := &l.tiles[i]
	if !t.Valid {
		return nil
	}
	return t
}

// setTile sets the tile at the given position.
func (l *Level) setTile(pos m.Pos, t *LevelTile) {
	i := pos.X + pos.Y*l.width
	l.tiles[i] = *t
}

// ForEachTile iterates over all tiles in the level.
func (l *Level) ForEachTile(f func(pos m.Pos, t *LevelTile)) {
	for i := range l.tiles {
		f(m.Pos{X: i % l.width, Y: i / l.width}, &l.tiles[i])
	}
}

// LevelTile is a single tile in the level.
type LevelTile struct {
	Tile      Tile
	WarpZones []*WarpZone
	Valid     bool
}

// WarpZone represents a warp tile. Whenever anything enters this tile, it gets
// moved to "to" and the direction transformed by "transform". For the game to
// work, every warpZone must be paired with an exact opposite elsewhere. This
// is ensured at load time. Warpzones can be temporarily toggled by name; this
// state is lost on checkpoint restore.
type WarpZone struct {
	Name         string
	InitialState bool
	PrevTile     m.Pos
	ToTile       m.Pos
	Transform    m.Orientation
}

// SaveGameData is a not-yet-hashed SaveGame.
type SaveGameData struct {
	State        map[EntityID]PersistentState
	LevelVersion int
	LevelHash    uint64
}

// SaveGame is the data structure we save game state with.
// It contains all needed (in addition to loading the level) to reset to the last visited checkpoint.
type SaveGame struct {
	SaveGameData
	Hash uint64
}

// SaveGame returns the current state as a SaveGame.
func (l *Level) SaveGame() (SaveGame, error) {
	save := SaveGame{
		SaveGameData: SaveGameData{
			State:        map[EntityID]PersistentState{},
			LevelVersion: l.SaveGameVersion,
			LevelHash:    l.Hash,
		},
	}
	saveOne := func(s *Spawnable) {
		if len(s.PersistentState) > 0 {
			save.State[s.ID] = s.PersistentState
		}
	}
	l.ForEachTile(func(_ m.Pos, tile *LevelTile) {
		for _, s := range tile.Tile.Spawnables {
			saveOne(s)
		}
	})
	saveOne(l.Player)
	var err error
	save.Hash, err = hashstructure.Hash(save.SaveGameData, hashstructure.FormatV2, nil)
	if err != nil {
		return SaveGame{}, err
	}
	return save, nil
}

// LoadGame loads the given SaveGame into the map.
// Note that when this returns an error, the SaveGame might have been partially loaded.
func (l *Level) LoadGame(save SaveGame) error {
	saveHash, err := hashstructure.Hash(save.SaveGameData, hashstructure.FormatV2, nil)
	if err != nil {
		return err
	}
	if saveHash != save.Hash {
		return fmt.Errorf("someone tampered with the save game")
	}
	if save.LevelVersion != l.SaveGameVersion {
		return fmt.Errorf("save game does not match level version: got %v, want %v", save.LevelVersion, l.SaveGameVersion)
	}
	if save.LevelHash != l.Hash {
		log.Printf("Save game does not match level hash: got %v, want %v; trying to load anyway", save.LevelHash, l.Hash)
	}
	loadOne := func(s *Spawnable) {
		// Do not reallocate the map! Works better with already loaded entities.
		for key := range s.PersistentState {
			delete(s.PersistentState, key)
		}
		for key, value := range save.State[s.ID] {
			s.PersistentState[key] = value
		}
	}
	l.ForEachTile(func(_ m.Pos, tile *LevelTile) {
		for _, s := range tile.Tile.Spawnables {
			loadOne(s)
		}
	})
	loadOne(l.Player)
	return nil
}

func Load(filename string) (*Level, error) {
	r, err := vfs.Load("maps", filename+".tmx")
	if err != nil {
		return nil, fmt.Errorf("could not open map: %v", err)
	}
	defer r.Close()
	t, err := tmx.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("invalid map: %v", err)
	}
	if t.Orientation != "orthogonal" {
		return nil, fmt.Errorf("unsupported map: got orientation %q, want orthogonal", t.Orientation)
	}
	// t.RenderOrder doesn't matter.
	// t.Width, t.Height used later.
	if t.TileWidth != TileSize || t.TileHeight != TileSize {
		return nil, fmt.Errorf("unsupported map: got tile size %dx%d, want %dx%d", t.TileWidth, t.TileHeight, TileSize, TileSize)
	}
	// t.HexSideLength doesn't matter.
	// t.StaggerAxis doesn't matter.
	// t.StaggerIndex doesn't matter.
	// t.BackgroundColor doesn't matter.
	// t.NextObjectID doesn't matter.
	if len(t.TileSets) != 1 {
		return nil, fmt.Errorf("unsupported map: got %d embedded tilesets, want 1", len(t.TileSets))
	}
	// t.Properties used later.
	if len(t.Layers) != 1 {
		return nil, fmt.Errorf("unsupported map: got %d layers, want 1", len(t.Layers))
	}
	// t.ObjectGroups used later.
	if len(t.ImageLayers) != 0 {
		return nil, fmt.Errorf("unsupported map: got %d image layers, want 0", len(t.ImageLayers))
	}
	for i, ts := range t.TileSets {
		if ts.Source != "" {
			r, err := vfs.Load("tiles", ts.Source)
			if err != nil {
				return nil, fmt.Errorf("could not open tileset: %v", err)
			}
			defer r.Close()
			decoded, err := tmx.DecodeTileset(r)
			if err != nil {
				return nil, fmt.Errorf("could not decode tileset: %v", err)
			}
			decoded.FirstGlobalID = ts.FirstGlobalID
			t.TileSets[i] = *decoded
		}
	}
	for _, ts := range t.TileSets {
		if ts.TileWidth != TileSize || ts.TileHeight != TileSize {
			return nil, fmt.Errorf("unsupported tileset: got tile size %dx%d, want %dx%d", ts.TileWidth, ts.TileHeight, TileSize, TileSize)
		}
		// ts.Spacing, ts.Margin, ts.TileCount, ts.Columns doesn't matter (we only support multi image tilesets).
		if ts.ObjectAlignment != "topleft" {
			return nil, fmt.Errorf("unsupported tileset: got objectalignment %q, want topleft", ts.ObjectAlignment)
		}
		// ts.Properties doesn't matter.
		if (ts.TileOffset != tmx.TileOffset{}) {
			return nil, fmt.Errorf("unsupported tileset: got a tile offset")
		}
		if ts.Image.Source != "" {
			return nil, fmt.Errorf("unsupported tileset: got single image, want image collection")
		}
		// ts.TerrainTypes doesn't matter (editor only).
		// ts.Tiles used later.
	}
	layer := &t.Layers[0]
	if layer.X != 0 || layer.Y != 0 {
		return nil, fmt.Errorf("unsupported map: layer has been shifted")
	}
	// layer.Width, layer.Height used later.
	// layer.Opacity, layer.Visible not used (we allow it though as it may help in the editor).
	if layer.OffsetX != 0 || layer.OffsetY != 0 {
		return nil, fmt.Errorf("unsupported map: layer has an offset")
	}
	// layer.Properties not used.
	// layer.RawData not used.
	tds, err := layer.TileDefs(t.TileSets)
	if err != nil {
		return nil, fmt.Errorf("invalid map layer: %v", err)
	}
	saveGameVersion, err := t.Properties.Int("save_game_version")
	if err != nil {
		return nil, fmt.Errorf("unsupported map: could not read save_game_version: %v", err)
	}
	level := Level{
		Checkpoints:     map[string]*Spawnable{},
		SaveGameVersion: int(saveGameVersion),
		tiles:           make([]LevelTile, layer.Width*layer.Height),
		width:           layer.Width,
	}
	for i, td := range tds {
		if td.Nil {
			continue
		}
		// td.Tile.Probability not used (editor only).
		// td.Tile.Properties used later.
		// td.Tile.Image used later.
		if len(td.Tile.Animation) != 0 {
			return nil, fmt.Errorf("unsupported tileset: got an animation")
		}
		if len(td.Tile.ObjectGroup.Objects) != 0 {
			return nil, fmt.Errorf("unsupported tileset: got objects in a tile")
		}
		// td.Tile.RawTerrainType not used (editor only).
		pos := m.Pos{X: i % layer.Width, Y: i / layer.Width}
		orientation := m.Identity()
		if td.HorizontallyFlipped {
			orientation = m.FlipX().Concat(orientation)
		}
		if td.VerticallyFlipped {
			orientation = m.FlipY().Concat(orientation)
		}
		if td.DiagonallyFlipped {
			orientation = m.FlipD().Concat(orientation)
		}
		properties := map[string]string{}
		for _, prop := range td.Tile.Properties {
			properties[prop.Name] = prop.Value
		}
		solid := properties["solid"] != "false"
		opaque := properties["opaque"] != "false"
		imgSrc := td.Tile.Image.Source
		imgSrcByOrientation := map[m.Orientation]string{}
		for propName, propValue := range properties {
			if oStr := strings.TrimPrefix(propName, "img."); oStr != propName {
				o, err := m.ParseOrientation(oStr)
				if err != nil {
					return nil, fmt.Errorf("invalid map: could not parse orientation tile: %v", err)
				}
				if o == m.Identity() && propValue != td.Tile.Image.Source {
					return nil, fmt.Errorf("invalid tileset: unrotated image isn't same as img: got %q, want %q", propValue, td.Tile.Image.Source)
				}
				imgSrcByOrientation[o] = propValue
			}
		}
		level.setTile(pos, &LevelTile{
			Tile: Tile{
				Solid:                 solid,
				Opaque:                opaque,
				LevelPos:              pos,
				ImageSrc:              imgSrc,
				ImageSrcByOrientation: imgSrcByOrientation,
				Orientation:           orientation,
			},
			Valid: true,
		})
	}
	type RawWarpZone struct {
		StartTile, EndTile m.Pos
		Orientation        m.Orientation
		InitialState       bool
	}
	warpZones := map[string][]RawWarpZone{}
	for _, og := range t.ObjectGroups {
		// og.Name, og.Color not used (editor only).
		if og.X != 0 || og.Y != 0 {
			return nil, fmt.Errorf("unsupported map: object group has been shifted")
		}
		// og.Width, og.Height not used.
		// og.Opacity, og.Visible not used (we allow it though as it may help in the editor).
		if og.OffsetX != 0 || og.OffsetY != 0 {
			return nil, fmt.Errorf("unsupported map: object group has an offset")
		}
		// og.DrawOrder not used (we use our own z index).
		// og.Properties not used.
		for _, o := range og.Objects {
			// o.ObjectID used later.
			properties := map[string]string{}
			if o.Name != "" {
				properties["name"] = o.Name
			}
			// o.X, o.Y, o.Width, o.Height used later.
			if o.Rotation != 0 {
				return nil, fmt.Errorf("unsupported map: object %v has a rotation (maybe implement this?)", o.ObjectID)
			}
			var tile *tmx.Tile
			if o.GlobalID != 0 {
				tile = t.TileSets[0].TileWithID(o.GlobalID.TileID(&t.TileSets[0]))
				if tile.Type == "" {
					properties["type"] = "Sprite"
				} else {
					properties["type"] = tile.Type
				}
				properties["image_dir"] = "tiles"
				properties["image"] = tile.Image.Source
				for _, prop := range tile.Properties {
					properties[prop.Name] = prop.Value
				}
			}
			// o.Visible not used (we allow it though as it may help in the editor).
			if o.Polygons != nil {
				return nil, fmt.Errorf("unsupported map: object %v has polygons", o.ObjectID)
			}
			if o.Polylines != nil {
				return nil, fmt.Errorf("unsupported map: object %v has polylines", o.ObjectID)
			}
			if o.Image.Source != "" {
				properties["type"] = "Sprite"
				properties["image_dir"] = "sprites"
				properties["image"] = o.Image.Source
			}
			if o.Type != "" {
				properties["type"] = o.Type
			}
			for _, prop := range o.Properties {
				properties[prop.Name] = prop.Value
			}
			// o.RawExtra not used.
			entRect := m.Rect{
				Origin: m.Pos{
					X: int(o.X),
					Y: int(o.Y),
				},
				Size: m.Delta{
					DX: int(o.Width),
					DY: int(o.Height),
				},
			}
			startTile := entRect.Origin.Div(TileSize)
			endTile := entRect.OppositeCorner().Div(TileSize)
			orientation := m.Identity()
			if orientationProp := properties["orientation"]; orientationProp != "" {
				orientation, err = m.ParseOrientation(orientationProp)
				if err != nil {
					return nil, fmt.Errorf("invalid orientation: %v", err)
				}
			}
			if properties["type"] == "WarpZone" {
				// WarpZones must be paired by name.
				name := properties["name"]
				initialState := properties["initial_state"] != "false" // Default enabled.
				warpZones[name] = append(warpZones[name], RawWarpZone{
					StartTile:    startTile,
					EndTile:      endTile,
					Orientation:  orientation,
					InitialState: initialState,
				})
				continue
			}
			ent := Spawnable{
				ID:         EntityID(o.ObjectID),
				EntityType: properties["type"],
				LevelPos:   startTile,
				RectInTile: m.Rect{
					Origin: entRect.Origin.Sub(
						startTile.Mul(TileSize).Delta(m.Pos{})),
					Size: entRect.Size,
				},
				Orientation:     orientation,
				Properties:      properties,
				PersistentState: PersistentState{},
			}
			if properties["type"] == "Player" {
				level.Player = &ent
				level.Checkpoints[""] = &ent
				// Do not link to tiles.
				continue
			}
			if properties["type"] == "Checkpoint" {
				level.Checkpoints[properties["name"]] = &ent
				// These do get linked.
			}
			for y := startTile.Y; y <= endTile.Y; y++ {
				for x := startTile.X; x <= endTile.X; x++ {
					pos := m.Pos{X: x, Y: y}
					levelTile := level.Tile(pos)
					if levelTile == nil {
						log.Panicf("Invalid entity location: outside map bounds: %v in %v", pos, ent)
					}
					levelTile.Tile.Spawnables = append(levelTile.Tile.Spawnables, &ent)
				}
			}
		}
	}
	for warpname, warppair := range warpZones {
		if len(warppair) != 2 {
			return nil, fmt.Errorf("unpaired WarpZone %q: got %d, want 2", warpname, len(warppair))
		}
		for a := 0; a < 2; a++ {
			from := warppair[a]
			to := warppair[1-a]
			// Warp orientation: right = direction to walk the warp, down = orientation (for mirroring).
			// Transform is identity transform iff the warps are reverse in right and identical in down.
			// T = to * flipx * from^-1
			// T' = from * flipx * to^-1
			// T T' = id
			transform := to.Orientation.Concat(m.FlipX()).Concat(from.Orientation.Inverse())
			fromCenter2 := from.StartTile.Add(from.EndTile.Delta(m.Pos{}))
			toCenter2 := to.StartTile.Add(to.EndTile.Delta(m.Pos{}))
			for fromy := from.StartTile.Y; fromy <= from.EndTile.Y; fromy++ {
				for fromx := from.StartTile.X; fromx <= from.EndTile.X; fromx++ {
					fromPos := m.Pos{X: fromx, Y: fromy}
					prevPos := fromPos.Add(from.Orientation.Apply(m.West()))
					fromPos2 := fromPos.Add(fromPos.Delta(m.Pos{}))
					toPos2 := toCenter2.Add(transform.Apply(fromPos2.Delta(fromCenter2)))
					toPos := toPos2.Div(2).Add(to.Orientation.Apply(m.West()))
					levelTile := level.Tile(fromPos)
					if levelTile == nil {
						log.Panicf("Invalid WarpZone location: outside map bounds: %v in %v", fromPos, warppair)
					}
					toTile := level.Tile(toPos)
					if toTile == nil {
						log.Panicf("Invalid WarpZone destination location: outside map bounds: %v in %v", toPos, warppair)
					}
					levelTile.WarpZones = append(levelTile.WarpZones, &WarpZone{
						Name:         warpname,
						InitialState: from.InitialState,
						PrevTile:     prevPos,
						ToTile:       toPos,
						Transform:    transform,
					})
				}
			}
		}
	}
	level.CheckpointLocations, err = level.LoadCheckpointLocations(filename)
	if err != nil {
		log.Printf("could not load checkpoint locations: %v", err)
	}
	level.Hash, err = hashstructure.Hash(&level, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, fmt.Errorf("could not hash level: %v", err)
	}
	return &level, nil
}
