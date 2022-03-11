package torqsrv

import (
	"github.com/gin-gonic/gin"
)

func Start(port int) {
	r := gin.Default()
	registerRoutes(r)
	r.Run()
}

func registerRoutes(r *gin.Engine) {
	registerStaticRoutes(r)
	api := r.Group("/api")
	{
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
	r.StaticFile("/logo192.png", "./web/build/logo192.png")
	r.StaticFile("/logo512.png", "./web/build/logo512.png")
	r.StaticFile("/manifest.json", "./web/build/manifest.json")
	r.StaticFile("/robots.txt", "./web/build/robots.txt")
	r.StaticFile("/asset-manifest.json", "./web/build/asset-manifest.json")
	r.Static("/static", "./web/build/static")

}
