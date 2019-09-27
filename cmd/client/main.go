package main

import (
	"image"
	"log"

	game "github.com/jilio/tomb_mates"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	e "github.com/hajimehoshi/ebiten"
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

	eimg, err := e.NewImageFromImage(frames["big_demon_run"][frame/7%4], e.FilterDefault)
	if err != nil {
		log.Println(err)
		return err
	}

	err = screen.DrawImage(eimg, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
