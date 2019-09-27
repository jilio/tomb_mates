package main

import (
	"image"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	e "github.com/hajimehoshi/ebiten"
	game "github.com/jilio/tomb_mates"
)

type Config struct {
	title  string
	width  int
	height int
	scale  float64
}

type Sprite struct {
	Frames []image.Image
	Frame  int
	X      float64
	Y      float64
}

var config *Config
var world *game.World
var frames map[string][]image.Image
var frame int

func init() {
	config = &Config{
		title:  "Another Hero",
		width:  320,
		height: 240,
		scale:  2,
	}

	world = &game.World{
		Replica: true,
		Units:   map[string]*game.Unit{},
	}

	var err error
	frames, err = game.LoadResources()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	go world.Evolve()

	c, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:3000/ws", nil)
	go func(c *websocket.Conn) {
		defer c.Close()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Fatal(err)
			}

			event := &game.Event{}
			err = proto.Unmarshal(message, event)
			if err != nil {
				log.Fatal(err)
			}

			world.HandleEvent(event)
		}
	}(c)

	e.SetRunnableInBackground(true)
	e.Run(update(c), config.width, config.height, config.scale, config.title)
}

func update(c *websocket.Conn) func(screen *e.Image) error {
	return func(screen *e.Image) error {
		frame++

		sprites := []Sprite{}
		for _, unit := range world.Units {
			sprites = append(sprites, Sprite{
				Frames: frames[unit.Skin+"_"+unit.Action],
				Frame:  int(unit.Frame),
				X:      unit.X,
				Y:      unit.Y,
			})
		}

		for _, sprite := range sprites {
			op := &e.DrawImageOptions{}
			op.GeoM.Translate(sprite.X, sprite.Y)

			img, err := e.NewImageFromImage(sprite.Frames[(frame/7+sprite.Frame)%4], e.FilterDefault)
			if err != nil {
				log.Println(err)
				return err
			}

			err = screen.DrawImage(img, op)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		err := handleMove(c)
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	}
}

func handleMove(c *websocket.Conn) error {
	event := &game.Event{}
	unit := world.Units[world.MyID]

	if e.IsKeyPressed(e.KeyA) || e.IsKeyPressed(e.KeyLeft) {
		event = &game.Event{
			Type: game.Event_type_move,
			Data: &game.Event_Move{
				&game.EventMove{
					PlayerId:  world.MyID,
					Direction: game.Direction_left,
				},
			},
		}
	}

	if e.IsKeyPressed(e.KeyD) || e.IsKeyPressed(e.KeyRight) {
		event = &game.Event{
			Type: game.Event_type_move,
			Data: &game.Event_Move{
				&game.EventMove{
					PlayerId:  world.MyID,
					Direction: game.Direction_right,
				},
			},
		}
	}

	if e.IsKeyPressed(e.KeyW) || e.IsKeyPressed(e.KeyUp) {
		event = &game.Event{
			Type: game.Event_type_move,
			Data: &game.Event_Move{
				&game.EventMove{
					PlayerId:  world.MyID,
					Direction: game.Direction_up,
				},
			},
		}
	}

	if e.IsKeyPressed(e.KeyS) || e.IsKeyPressed(e.KeyDown) {
		event = &game.Event{
			Type: game.Event_type_move,
			Data: &game.Event_Move{
				&game.EventMove{
					PlayerId:  world.MyID,
					Direction: game.Direction_down,
				},
			},
		}
	}

	if event.Type == game.Event_type_move {
		if unit.Action != game.UnitActionMove {
			message, err := proto.Marshal(event)
			if err != nil {
				log.Println(err)
				return err
			}
			c.WriteMessage(websocket.BinaryMessage, message)
		}
	} else {
		if unit.Action != game.UnitActionIdle {
			event = &game.Event{
				Type: game.Event_type_idle,
				Data: &game.Event_Idle{
					&game.EventIdle{PlayerId: world.MyID},
				},
			}
			message, err := proto.Marshal(event)
			if err != nil {
				log.Println(err)
				return err
			}
			c.WriteMessage(websocket.BinaryMessage, message)
		}
	}

	return nil
}
