package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/divVerent/aaaaaa/internal/centerprint"
	"github.com/divVerent/aaaaaa/internal/engine"
	m "github.com/divVerent/aaaaaa/internal/math"
)

// Checkpoint remembers that it was hit and allows spawning from there again. Also displays a text.
type Checkpoint struct {
	World  *engine.World
	Entity *engine.Entity

	RequiredOrientation m.Orientation
	PlayerProperty      string
	Name                string
	Text                string
}

func (c *Checkpoint) Spawn(w *engine.World, s *engine.Spawnable, e *engine.Entity) error {
	c.World = w
	c.Entity = e
	var err error
	c.RequiredOrientation, err = m.ParseOrientation(s.Properties["required_orientation"])
	if err != nil {
		return fmt.Errorf("could not parse required orientation: %v", err)
	}
	// Field contains orientation OF THE PLAYER to make it easier in the map editor. So we need to invert.
	c.RequiredOrientation = c.RequiredOrientation.Inverse()
	c.Name = s.Properties["name"]
	c.PlayerProperty = "checkpoint_seen." + c.Name
	c.Text = s.Properties["text"]
	return nil
}

func (c *Checkpoint) Despawn() {}

func (c *Checkpoint) Update() {
	// The "down" direction must match. That way we allow x-flipping and still matching the CP.
	flippedStr := "Identity"
	if c.Entity.Orientation != c.RequiredOrientation {
		if c.Entity.Orientation != m.FlipX().Concat(c.RequiredOrientation) {
			return
		}
		flippedStr = "FlipX"
	}
	if (c.World.Player.Rect.Delta(c.Entity.Rect) != m.Delta{}) {
		return
	}
	player := c.World.Player.Impl.(*Player)
	if player.PersistentState["last_checkpoint"] == c.Name && player.PersistentState[c.PlayerProperty] == flippedStr {
		return
	}
	player.PersistentState[c.PlayerProperty] = flippedStr
	player.PersistentState["last_checkpoint"] = c.Name
	centerprint.New(c.Text, centerprint.Important, centerprint.Middle, centerprint.BigFont, color.NRGBA{R: 255, G: 255, B: 255, A: 255}).SetFadeOut(true)
}

func (c *Checkpoint) Touch(other *engine.Entity) {}

func (c *Checkpoint) DrawOverlay(screen *ebiten.Image, scrollDelta m.Delta) {}

func init() {
	engine.RegisterEntityType(&Checkpoint{})
}
