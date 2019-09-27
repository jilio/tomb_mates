package tinyrpg

import (
	"log"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
)

// World represents game state
type World struct {
	Replica bool
	Units   map[string]*Unit
	MyID    string
}

func (world *World) AddPlayer() string {
	id := uuid.NewV4().String()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	unit := &Unit{
		Id: id,
		X:  rnd.Float64()*300 + 10,
		Y:  rnd.Float64()*220 + 10,
	}
	world.Units[id] = unit

	return id
}

func (world *World) HandleEvent(event *Event) {
	log.Println(event.GetType())
	log.Println(event.GetData())

	switch event.GetType() {
	case Event_type_connect:
		data := event.GetConnect()
		world.Units[data.Unit.Id] = data.Unit

	case Event_type_init:
		data := event.GetInit()
		if world.Replica {
			world.MyID = data.PlayerId
			world.Units = data.Units
		}

	default:
		log.Println("UNKNOWN EVENT: ", event)
	}
}
