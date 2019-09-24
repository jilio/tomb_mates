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

var world game.World
var frame int
var img *e.Image

func init() {
	world = game.World{
		IsServer: false,
		Units:    game.Units{},
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

		unitList := []*game.Unit{}
		for _, unit := range world.Units {
			unitList = append(unitList, unit)
		}
		sort.Slice(unitList, func(i, j int) bool {
			return unitList[i].Y < unitList[j].Y
		})

		for _, unit := range unitList {
			op := &e.DrawImageOptions{}
			if unit.HorizontalDirection == game.DirectionLeft {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(16, 0)
			}
			op.GeoM.Translate(unit.X, unit.Y)

			spriteIndex := (frame/7 + unit.Frame) % 4
			img, _, _ = ebitenutil.NewImageFromFile(
				"sprites/"+unit.SpriteName+"_"+unit.Action+"_anim_f"+strconv.Itoa(spriteIndex)+".png",
				e.FilterDefault,
			)
			screen.DrawImage(img, op)
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
