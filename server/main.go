package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jilio/tomb_mates/game"
)

func main() {
	world := &game.World{
		IsServer: true,
		Units:    game.Units{},
		Items:    game.Items{},
	}
	world.AddItem(game.ItemCoin, 65, 100)
	world.AddItem(game.ItemCoin, 75, 110)
	world.AddItem(game.ItemHealthPotion, 190, 45)

	hub := newHub()
	go hub.run()

	r := gin.New()
	r.GET("/ws", func(hub *Hub, world *game.World) gin.HandlerFunc {
		return gin.HandlerFunc(func(c *gin.Context) {
			serveWs(hub, world, c.Writer, c.Request)
		})
	}(hub, world))
	r.Run(":3000")
}
