package torqsrv

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/cockroachdb/errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/lncapital/torq/internal/auth"
	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/categories"
	"github.com/lncapital/torq/internal/channel_history"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/flow"
	"github.com/lncapital/torq/internal/forwards"
	"github.com/lncapital/torq/internal/invoices"
	"github.com/lncapital/torq/internal/lightning"
	"github.com/lncapital/torq/internal/messages"
	"github.com/lncapital/torq/internal/nodes"
	"github.com/lncapital/torq/internal/on_chain_tx"
	"github.com/lncapital/torq/internal/payments"
	"github.com/lncapital/torq/internal/peers"
	"github.com/lncapital/torq/internal/services"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/internal/views"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/web"
)

func Start(port int, apiPswd string, cookiePath string, db *sqlx.DB, autoLogin bool) error {
	r := gin.Default()

	if err := auth.RefreshCookieFile(cookiePath); err != nil {
		return errors.Wrap(err, "Refreshing cookie file")
	}

	err := auth.CreateSession(r, apiPswd)
	if err != nil {
		return errors.Wrap(err, "Creating Gin Session")
	}

	registerRoutes(r, db, apiPswd, cookiePath, autoLogin)

	fmt.Println("Listening on port " + strconv.Itoa(port))

	if err := r.Run(":" + strconv.Itoa(port)); err != nil {
		return errors.Wrap(err, "Running gin webserver")
	}
	return nil
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
		log.Fatal().Err(err).Send()
	}
	store := memory.NewStore()
	return mgin.NewMiddleware(limiter.New(store, rate), mgin.WithKeyGetter(loginKeyGetter))
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}

func registerRoutes(r *gin.Engine, db *sqlx.DB, apiPwd string, cookiePath string, autoLogin bool) {
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	applyCors(r)
	// Websocket
	ws := r.Group("/ws")
	ws.Use(auth.AuthRequired(autoLogin))
	ws.GET("", func(c *gin.Context) {
		err := WebsocketHandler(c, db)
		log.Debug().Msgf("WebsocketHandler: %v", err)
	})

	api := r.Group("/api")

	api.POST("/logout", auth.Logout)

	// Limit login attempts to 10 per minute.
	rl := NewLoginRateLimitMiddleware()
	api.POST("/login", rl, auth.Login(apiPwd))
	api.POST("/cookie-login", rl, auth.CookieLogin(cookiePath))
	api.GET("auto-login-setting", rl, auth.AutoLoginSetting(autoLogin))

	unauthorisedSettingRoutes := api.Group("settings")
	{
		settings.RegisterUnauthenticatedRoutes(unauthorisedSettingRoutes, db)
	}

	unauthorisedServicesRoutes := api.Group("services")
	{
		services.RegisterUnauthenticatedRoutes(unauthorisedServicesRoutes, db)
	}

	api.Use(auth.AuthRequired(autoLogin)).Use(auth.TorqRequired)
	{

		tableViewRoutes := api.Group("/table-views")
		{
			views.RegisterTableViewRoutes(tableViewRoutes, db)
		}

		categoryRoutes := api.Group("/categories")
		{
			categories.RegisterCategoryRoutes(categoryRoutes, db)
		}

		tagRoutes := api.Group("/tags")
		{
			tags.RegisterTagRoutes(tagRoutes, db)
		}

		corridorRoutes := api.Group("/corridors")
		{
			corridors.RegisterCorridorRoutes(corridorRoutes, db)
		}

		paymentRoutes := api.Group("/payments")
		{
			payments.RegisterPaymentsRoutes(paymentRoutes, db)
		}

		invoiceRoutes := api.Group("/invoices")
		{
			invoices.RegisterInvoicesRoutes(invoiceRoutes, db)
		}

		onChainTx := api.Group("/on-chain-tx")
		{
			on_chain_tx.RegisterOnChainTxsRoutes(onChainTx, db)
		}

		peerRoutes := api.Group("/peers")
		{
			peers.RegisterPeerRoutes(peerRoutes, db)
		}

		nodeRoutes := api.Group("/nodes")
		{
			nodes.RegisterNodeRoutes(nodeRoutes, db)
		}

		channelRoutes := api.Group("/channels")
		{
			channel_history.RegisterChannelHistoryRoutes(channelRoutes, db)
			channels.RegisterChannelRoutes(channelRoutes, db)
		}

		forwardRoutes := api.Group("/forwards")
		{
			forwards.RegisterForwardsRoutes(forwardRoutes, db)
		}

		flowRoutes := api.Group("/flow")
		{
			flow.RegisterFlowRoutes(flowRoutes, db)
		}

		lightningRoutes := api.Group("/lightning")
		{
			lightning.RegisterLightningRoutes(lightningRoutes, db)
		}

		workflowRoutes := api.Group("/workflows")
		{
			workflows.RegisterWorkflowRoutes(workflowRoutes, db)
		}

		automationRoutes := api.Group("/automation")
		{
			automation.RegisterAutomationRoutes(automationRoutes, db)
		}

		messageRoutes := api.Group("messages")
		{
			messages.RegisterMessagesRoutes(messageRoutes)
		}

		settingRoutes := api.Group("settings")
		{
			settings.RegisterSettingRoutes(settingRoutes, db)
		}

		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	web.AddRoutes(r)

	registerStaticRoutes(r)
}

func registerStaticRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		knownAssetList := []string{
			"/favicon.ico",
			"/favicon-16x16.png",
			"/favicon-32x32.png",
			"/mstile-150x150.png",
			"/safari-pinned-tab.svg",
			"/android-chrome-192x192.png",
			"/android-chrome-512x512.png",
			"/apple-touch-icon.png",
			"/browserconfig.xml",
			"/manifest.json",
			"/robots.txt"}

		for _, knownAsset := range knownAssetList {
			if strings.HasSuffix(path, knownAsset) {
				c.File("./web/build" + knownAsset)
				return
			}
		}

		// probably a file, this might not be bulletproof
		if strings.Contains(path, "/static/") && strings.Contains(path, ".") &&
			(strings.Contains(path, "css") || strings.Contains(path, "js") || strings.Contains(path, "media")) {
			parts := strings.Split(path, "/")
			c.File("./web/build/static/" + parts[len(parts)-2] + "/" + parts[len(parts)-1])
			return
		}

		// locales json files
		if strings.Contains(path, "/locales/") && strings.Contains(path, ".json") {
			parts := strings.Split(path, "/")
			c.File("./web/build/locales/" + parts[len(parts)-1])
			return
		}

		c.File("./web/build/index.html")
	})
}
