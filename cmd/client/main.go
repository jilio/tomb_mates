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
	e.Run(update, config.width, config.height, config.scale, config.title)
}

func update(screen *e.Image) error {
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

	return nil
}
