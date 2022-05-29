package torqsrv

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lncapital/torq/internal/auth"
	"github.com/lncapital/torq/internal/channel_history"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/flow"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/views"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"log"
	"strconv"
)

func Start(port int, apiPswd string, db *sqlx.DB, restartLNDSub func()) {
	r := gin.Default()
	applyCors(r)
	registerRoutes(r, db, apiPswd, restartLNDSub)

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

// loginKeyGetter is used to force the Login rate
// limiter to limit all requests regardless of IP etc.
func loginKeyGetter(c *gin.Context) string {
	return "login_limiter"
}

// NewLoginRateLimitMiddleware is used to limit login attempts
func NewLoginRateLimitMiddleware() gin.HandlerFunc {
	// Define a limit rate to 10 requests per minute.
	rate, err := limiter.NewRateFromFormatted("10-M")
	if err != nil {
		log.Fatal(err)
	}
	store := memory.NewStore()
	return mgin.NewMiddleware(limiter.New(store, rate), mgin.WithKeyGetter(loginKeyGetter))
}

func apiPasswordMiddleware(apiPswd string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("apiPswd", apiPswd)
	}
}

func registerRoutes(r *gin.Engine, db *sqlx.DB, apiPwd string, restartLNDSub func()) {
	registerStaticRoutes(r)

	// TODO: Generate this secret!
	var Secret = []byte("secret")
	r.Use(sessions.Sessions("torq_session", sessions.NewCookieStore(Secret)))

	api := r.Group("/api")

	api.POST("/logout", auth.Logout)

	// Limit login attempts to 10 per minute.
	rl := NewLoginRateLimitMiddleware()
	api.POST("/login", rl, auth.Login(apiPwd))

	unauthorisedSettingRoutes := api.Group("settings")
	{
		settings.RegisterUnauthorisedRoutes(unauthorisedSettingRoutes, db)
	}

	api.Use(auth.AuthRequired)
	{

		tableViewRoutes := api.Group("/table-views")
		{
			views.RegisterTableViewRoutes(tableViewRoutes, db)
		}

		//channelTagRoutes := api.Group("/tags")
		//{
		//	tags.RegisterTagRoutes(channelTagRoutes, db)
		//}

		channelRoutes := api.Group("/channels")
		{
			channels.RegisterChannelRoutes(channelRoutes, db)
			channel_history.RegisterChannelHistoryRoutes(channelRoutes, db)

		}

		flowRoutes := api.Group("/flow")
		{
			flow.RegisterFlowRoutes(flowRoutes, db)
		}

		settingRoutes := api.Group("settings")
		{
			settings.RegisterSettingRoutes(settingRoutes, db, restartLNDSub)
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
