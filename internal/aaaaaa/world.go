package aaaaaa

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// tilePos represents a tile position.
type tilePos struct {
	c, r int
}

// tileDelta represents a move between two tiles.
type tileDelta struct {
	dc, dr int
}

// World represents the current game state including its entities.
type World struct {
	// tiles are all tiles currently loaded.
	tiles map[tilePos]Tile
	// entities are all entities currently loaded.
	entities map[EntityID]*Entity
	// scrollPos is the current screen scrolling position.
	scrollPos Pos
	// scrollTarget is where we want to scroll to.
	scrollTarget Pos
	// scrollSpeed is the speed of scrolling to ScrollTarget, or 0 if not aiming for a target.
	scrollSpeed int
	// level is the current tilemap (universal covering with warpzones).
	level *level
}

func NewWorld() *World {
	// Load map.
	// Create player entity.
	// Load in the tile the player is standing on.
	return &World{}
}

func (w *World) Update() error {
	// Let all entities move/act. Fetch player position.
	// Update ScrollPos based on player position and scroll target.
	// Unmark all tiles and entities (just bump mark index).
	// Trace from player location to all directions.
	// Remember trace polygon.
	// Mark all tiles hit (incl. the tiles that stopped us).
	// Mark all entities hit.
	// Delete all unmarked entities.
	// Spawn all entities on existing tiles if not already spawned.
	// Mark all tiles on entities.
	// Delete all unmarked tiles.
	return nil
}

func (w *World) Draw(screen *ebiten.Image) {
	// Draw trace polygon to buffer.
	// Expand and blur buffer.
	// Draw all tiles currently marked for drawing to screen.
	// Multiply screen with buffer.
	// Invert buffer.
	// Multiply with previous screen, scroll pos delta applied.
	// Blur and darken buffer.
	// Add buffer to screen.
}

// LoadTile loads the next tile into the current world based on a currently
// known tile and its neighbor. Respects and applies warps.
func (w *World) loadTile(p tilePos, d tileDelta) tilePos {
	// TODO implement
	return tilePos{}
}

type TraceOptions struct {
	// If NoTiles is set, we ignore hits against tiles.
	NoTiles bool
	// If NoEntities is set, we ignore hits against entities.
	NoEntities bool
	// If LoadTiles is set, not yet known tiles will be loaded in by the trace operation.
	// Otherwise hitting a not-yet-loaded tile will end the trace.
	LoadTiles bool
}

// TraceResult returns the status of a trace operation.
type TraceResult struct {
	// Delta is the distance actually travelled until the trace stopped.
	Vector Delta
	// Path is the set of tiles touched, not including what stopped the trace.
	// For a line trace, any two neighboring tiles here are adjacent.
	path []tilePos
	// Entities is the set of entities touched, not including what stopped the trace.
	Entities []Entity
	// hitSolidTilePos is the position of the tile that stopped the trace, if any.
	hitSolidTilePos *tilePos
	// HitSolidTile is the tile that stopped the trace, if any.
	HitSolidTile *Tile
	// HitSolidEntity is the entity that stopped the trace, if any.
	HitSolidEntity Entity
	// HitFogOfWar is set if the trace ended by hitting an unloaded tile.
	HitFogOfWar bool
}

// TraceLine moves from x,y by dx,dy in pixel coordinates.
func (w *World) TraceLine(p Pos, d Delta, o TraceOptions) TraceResult {
	// TODO: Optimize?
	return w.TraceBox(p, Delta{}, d, o)
}

// TraceBox moves from x,y size sx,sy by dx,dy in pixel coordinates.
func (w *World) TraceBox(p Pos, s, d Delta, o TraceOptions) TraceResult {
	// TODO: Implement
	return TraceResult{}
}
