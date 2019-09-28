package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	e "github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
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
	Side   game.Direction
	Config image.Config
}

type Camera struct {
	X       float64
	Y       float64
	Padding float64
}

var config *Config
var world *game.World
var camera *Camera
var frames map[string]game.Frames
var frame int
var lastKey e.Key
var prevKey e.Key
var levelImage *e.Image

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

	levelImage, err = prepareLevelImage()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	go world.Evolve()

	host := getEnv("HOST", "localhost")
	c, _, _ := websocket.DefaultDialer.Dial("ws://"+host+":3000/ws", nil)
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

			if event.Type == game.Event_type_connect {
				me := world.Units[world.MyID]
				camera = &Camera{
					X:       me.X,
					Y:       me.Y,
					Padding: 30,
				}
			}
		}
	}(c)

	e.SetRunnableInBackground(true)
	e.Run(update(c), config.width, config.height, config.scale, config.title)
}

func update(c *websocket.Conn) func(screen *e.Image) error {
	return func(screen *e.Image) error {
		handleKeyboard(c)

		if e.IsDrawingSkipped() {
			return nil
		}

		handleCamera(screen)

		frame++

		sprites := []Sprite{}
		for _, unit := range world.Units {
			sprites = append(sprites, Sprite{
				Frames: frames[unit.Skin+"_"+unit.Action].Frames,
				Frame:  int(unit.Frame),
				X:      unit.X,
				Y:      unit.Y,
				Side:   unit.Side,
				Config: frames[unit.Skin+"_"+unit.Action].Config,
			})
		}
		sort.Slice(sprites, func(i, j int) bool {
			depth1 := sprites[i].Y + float64(sprites[i].Config.Height)
			depth2 := sprites[j].Y + float64(sprites[j].Config.Height)
			return depth1 < depth2
		})

		for _, sprite := range sprites {
			op := &e.DrawImageOptions{}

			if sprite.Side == game.Direction_left {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(float64(sprite.Config.Width), 0)
			}

			op.GeoM.Translate(sprite.X-camera.X, sprite.Y-camera.Y)

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

		ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", e.CurrentTPS()))

		return nil
	}
}

func prepareLevelImage() (*e.Image, error) {
	tileSize := 16
	level := game.LoadLevel()
	width := len(level[0])
	height := len(level)
	levelImage, _ := e.NewImage(width*tileSize, height*tileSize, e.FilterDefault)

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			op := &e.DrawImageOptions{}
			op.GeoM.Translate(float64(i*tileSize), float64(j*tileSize))

			img, err := e.NewImageFromImage(frames[level[j][i]].Frames[0], e.FilterDefault)
			if err != nil {
				log.Println(err)
				return levelImage, err
			}
			err = levelImage.DrawImage(img, op)
			if err != nil {
				log.Println(err)
				return levelImage, err
			}
		}
	}

	return levelImage, nil
}

func handleCamera(screen *e.Image) {
	if camera == nil {
		return
	}

	player := world.Units[world.MyID]
	frame := frames[player.Skin+"_"+player.Action]
	camera.X = player.X - float64(config.width-frame.Config.Width)/2
	camera.Y = player.Y - float64(config.height-frame.Config.Height)/2

	op := &e.DrawImageOptions{}
	op.GeoM.Translate(-camera.X, -camera.Y)
	screen.DrawImage(levelImage, op)
}

func handleKeyboard(c *websocket.Conn) {
	event := &game.Event{}

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
		if lastKey != e.KeyA {
			lastKey = e.KeyA
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
		if lastKey != e.KeyD {
			lastKey = e.KeyD
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
		if lastKey != e.KeyW {
			lastKey = e.KeyW
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
		if lastKey != e.KeyS {
			lastKey = e.KeyS
		}
	}

	unit := world.Units[world.MyID]

	if event.Type == game.Event_type_move {
		if prevKey != lastKey {
			message, err := proto.Marshal(event)
			if err != nil {
				log.Println(err)
				return
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
				return
			}
			c.WriteMessage(websocket.BinaryMessage, message)
			lastKey = -1
		}
	}

	prevKey = lastKey
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
