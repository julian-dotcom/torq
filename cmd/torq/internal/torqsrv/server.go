package torqsrv

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/channels/tags"
	"strconv"
)

func Start(port int, db *sqlx.DB) {
	r := gin.Default()
	registerRoutes(r, db)
	applyCors(r)
	fmt.Println("Listening on port " + strconv.Itoa(port))
	r.Run(":" + strconv.Itoa(port))
}

func applyCors(r *gin.Engine) {
	corsConfig := cors.DefaultConfig()
	//hot reload CORS
	corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))
}

func registerRoutes(r *gin.Engine, db *sqlx.DB) {
	registerStaticRoutes(r)
	api := r.Group("/api")
	{
		channelRoutes := api.Group("/channels")
		{
			channels.RegisterChannelRoutes(channelRoutes, db)
			channelTagRoutes := channelRoutes.Group(":channelDbId/tags")
			{
				tags.RegisterTagRoutes(channelTagRoutes, db)
			}
		}

		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}
}

func registerStaticRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/build/index.html")
	})
	r.StaticFile("/", "./web/build/index.html")
	r.StaticFile("/favicon.ico", "./web/build/favicon.ico")
	r.StaticFile("/favicon-16x16.png", "./web/build/favicon-16x16.png")
	r.StaticFile("/favicon-32x32.png", "./web/build/favicon-32x32.png")
	r.StaticFile("/mstile-150x150.png", "./web/build/mstile-150x150.png")
	r.StaticFile("/safari-pinned-tab.svg", "./web/build/safari-pinned-tab.svg")
	r.StaticFile("/android-chrome-192x192.png", "./web/build/android-chrome-192x192.png")
	r.StaticFile("/android-chrome-512x512.png", "./web/build/android-chrome-512x512.png")
	r.StaticFile("/apple-touch-icon.png", "./web/build/apple-touch-icon.png")
	r.StaticFile("/browserconfig.xml", "./web/build/browserconfig.xml")
	r.StaticFile("/manifest.json", "./web/build/manifest.json")
	r.StaticFile("/robots.txt", "./web/build/robots.txt")
	r.Static("/static", "./web/build/static")
}
