package main

import (
	game "github.com/jilio/tomb_mates"
	"github.com/gin-gonic/gin"
)

var world *game.World

func init() {
	world = &game.World{
		Replica: false,
		Units:   map[string]*game.Unit{},
	}
}

func main() {
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
