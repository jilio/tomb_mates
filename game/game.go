package game

import (
	"encoding/json"
	"math/rand"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Unit struct {
	ID                  string  `json:"id"`
	X                   float64 `json:"x"`
	Y                   float64 `json:"y"`
	SpriteName          string  `json:"sprite_name"`
	Action              string  `json:"action"`
	Frame               int     `json:"frame"`
	HorizontalDirection int     `json:"direction"`
}

type Units map[string]*Unit

type World struct {
	MyID     string `json:"-"`
	IsServer bool   `json:"-"`
	Units    `json:"units"`
}

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type EventConnect struct {
	Unit
}

type EventMove struct {
	UnitID    string `json:"unit_id"`
	Direction int    `json:"direction"`
}

type EventIdle struct {
	UnitID string `json:"unit_id"`
}

type EventInit struct {
	PlayerID string `json:"player_id"`
	Units    Units  `json:"units"`
}

type EventExit struct {
	UnitID string `json:"unit_id"`
}

const EventTypeConnect = "connect"
const EventTypeMove = "move"
const EventTypeIdle = "idle"
const EventTypeInit = "init"
const EventTypeExit = "exit"

const ActionRun = "run"
const ActionIdle = "idle"

const DirectionUp = 0
const DirectionDown = 1
const DirectionLeft = 2
const DirectionRight = 3

func (world *World) HandleEvent(event *Event) {
	switch event.Type {
	case EventTypeConnect:
		str, _ := json.Marshal(event.Data)
		var ev EventConnect
		json.Unmarshal(str, &ev)

		world.Units[ev.ID] = &ev.Unit

	case EventTypeInit:
		str, _ := json.Marshal(event.Data)
		var ev EventInit
		json.Unmarshal(str, &ev)

		if !world.IsServer {
			world.MyID = ev.PlayerID
			world.Units = ev.Units
		}

	case EventTypeMove:
		str, _ := json.Marshal(event.Data)
		var ev EventMove
		json.Unmarshal(str, &ev)

		unit := world.Units[ev.UnitID]
		unit.Action = ActionRun

		switch ev.Direction {
		case DirectionUp:
			unit.Y--
		case DirectionDown:
			unit.Y++
		case DirectionLeft:
			unit.X--
			unit.HorizontalDirection = ev.Direction
		case DirectionRight:
			unit.X++
			unit.HorizontalDirection = ev.Direction
		}

	case EventTypeIdle:
		str, _ := json.Marshal(event.Data)
		var ev EventIdle
		json.Unmarshal(str, &ev)

		unit := world.Units[ev.UnitID]
		unit.Action = ActionIdle

	case EventTypeExit:
		str, _ := json.Marshal(event.Data)
		var ev EventExit
		json.Unmarshal(str, &ev)

		delete(world.Units, ev.UnitID)
	}
}

func (world *World) AddPlayer() *Unit {
	skins := []string{
		"elf_f", "elf_m", "knight_f", "knight_m",
		"lizard_f", "lizard_m", "wizzard_f", "wizzard_m",
	}
	id := uuid.NewV4().String()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	unit := &Unit{
		ID:         id,
		Action:     ActionIdle,
		X:          rnd.Float64() * 320,
		Y:          rnd.Float64() * 240,
		Frame:      rnd.Intn(4),
		SpriteName: skins[rnd.Intn(len(skins))],
	}
	world.Units[id] = unit

	return unit
}
