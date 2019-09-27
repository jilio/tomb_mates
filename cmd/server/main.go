package main

import (
	"github.com/gin-gonic/gin"
	game "github.com/jilio/tomb_mates"
)

var world *game.World

func init() {
	world = &game.World{
		Replica: false,
		Units:   map[string]*game.Unit{},
	}
}

func main() {
	go world.Evolve()

	hub := newHub()
	go hub.run()

	r := gin.New()
	r.GET("/ws", ginWsServe(hub, world))
	r.Run(":3000")
}

func ginWsServe(hub *Hub, world *game.World) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		serveWs(hub, world, c.Writer, c.Request)
	})
}
