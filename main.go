package main

import (
	"log"
	"sort"
	"strconv"

	"github.com/gorilla/websocket"
	e "github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/jilio/tomb_mates/game"
)

type Sprite []*e.Image
type Drawable struct {
	Sprite              Sprite
	X, Y                float64
	Frame               int
	HorizontalDirection int
}

var world game.World
var frame int
var img *e.Image
var unitSprites map[string]map[string]Sprite
var itemSprites map[string]Sprite

func init() {
	world = game.World{
		IsServer: false,
		Units:    game.Units{},
	}

	unitSprites = map[string]map[string]Sprite{}
	for _, skin := range game.GetUnits() {
		unitSprites[skin] = map[string]Sprite{}
		for _, action := range []string{game.ActionIdle, game.ActionRun} {
			sprite := make(Sprite, 4)
			for i := 0; i < 4; i++ {
				path := "sprites/" + skin + "_" + action + "_anim_f" + strconv.Itoa(i) + ".png"
				img, _, _ = ebitenutil.NewImageFromFile(path, e.FilterDefault)
				sprite[i] = img
			}
			unitSprites[skin][action] = sprite
		}
	}

	itemSprites = map[string]Sprite{}
	for _, item := range game.GetItems() {
		if item.Frames == 1 {
			img, _, _ = ebitenutil.NewImageFromFile(item.Prefix+".png", e.FilterDefault)
			itemSprites[item.Entity] = []*e.Image{img}
		}
		if item.Frames > 1 {
			sprite := make(Sprite, item.Frames)
			for i := 0; i < item.Frames; i++ {
				path := item.Prefix + strconv.Itoa(i) + ".png"
				img, _, _ = ebitenutil.NewImageFromFile(path, e.FilterDefault)
				sprite[i] = img
			}
			itemSprites[item.Entity] = sprite
		}
	}
}

func update(c *websocket.Conn) func(screen *e.Image) error {
	return func(screen *e.Image) error {
		frame++

		img, _, _ = ebitenutil.NewImageFromFile(
			"sprites/background.png",
			e.FilterDefault,
		)
		screen.DrawImage(img, nil)

		drawableList := []*Drawable{}
		for _, unit := range world.Units {
			drawableList = append(drawableList, &Drawable{
				unitSprites[unit.SpriteName][unit.Action],
				unit.X,
				unit.Y,
				unit.Frame,
				unit.HorizontalDirection,
			})
		}
		for _, item := range world.Items {
			drawableList = append(drawableList, &Drawable{
				itemSprites[item.Entity],
				item.X,
				item.Y,
				0,
				game.DirectionRight,
			})
		}
		sort.Slice(drawableList, func(i, j int) bool {
			return drawableList[i].Y < drawableList[j].Y // todo: depth instead of Y
		})

		for _, drawable := range drawableList {
			op := &e.DrawImageOptions{}
			if drawable.HorizontalDirection == game.DirectionLeft {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(16, 0) // todo: half width instead of 16
			}
			op.GeoM.Translate(drawable.X, drawable.Y)

			spriteIndex := (frame/7 + drawable.Frame) % len(drawable.Sprite)
			screen.DrawImage(drawable.Sprite[spriteIndex], op)
		}

		if e.IsKeyPressed(e.KeyD) || e.IsKeyPressed(e.KeyRight) {
			c.WriteJSON(game.Event{
				Type: game.EventTypeMove,
				Data: game.EventMove{
					UnitID:    world.MyID,
					Direction: game.DirectionRight,
				},
			})
			return nil
		}

		if e.IsKeyPressed(e.KeyA) || e.IsKeyPressed(e.KeyLeft) {
			c.WriteJSON(game.Event{
				Type: game.EventTypeMove,
				Data: game.EventMove{
					UnitID:    world.MyID,
					Direction: game.DirectionLeft,
				},
			})
			return nil
		}

		if e.IsKeyPressed(e.KeyW) || e.IsKeyPressed(e.KeyUp) {
			c.WriteJSON(game.Event{
				Type: game.EventTypeMove,
				Data: game.EventMove{
					UnitID:    world.MyID,
					Direction: game.DirectionUp,
				},
			})
			return nil
		}

		if e.IsKeyPressed(e.KeyS) || e.IsKeyPressed(e.KeyDown) {
			c.WriteJSON(game.Event{
				Type: game.EventTypeMove,
				Data: game.EventMove{
					UnitID:    world.MyID,
					Direction: game.DirectionDown,
				},
			})
			return nil
		}

		if world.Units[world.MyID].Action == game.ActionRun {
			c.WriteJSON(game.Event{
				Type: game.EventTypeIdle,
				Data: game.EventMove{
					UnitID: world.MyID,
				},
			})
		}

		return nil
	}
}

func main() {
	c, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:3000/ws", nil)
	go func(c *websocket.Conn) {
		defer c.Close()

		for {
			var event game.Event
			c.ReadJSON(&event)
			world.HandleEvent(&event)
			log.Println(event)
		}
	}(c)

	e.SetRunnableInBackground(true)
	e.Run(update(c), 320, 240, 2, "Tomb Mates")
}
