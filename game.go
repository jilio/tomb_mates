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
		Id:     id,
		X:      rnd.Float64()*300 + 10,
		Y:      rnd.Float64()*220 + 10,
		Frame:  int32(rnd.Intn(4)),
		Skin:   []string{"big_demon", "big_zombie"}[rnd.Intn(2)],
		Action: "idle",
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

	case Event_type_exit:
		data := event.GetExit()
		delete(world.Units, data.PlayerId)

	default:
		log.Println("UNKNOWN EVENT: ", event)
	}
}
