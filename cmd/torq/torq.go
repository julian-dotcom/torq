package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/cmd/torq/internal/amboss_ping"
	"github.com/lncapital/torq/cmd/torq/internal/notifications"
	"github.com/lncapital/torq/cmd/torq/internal/services"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/cmd/torq/internal/vector_ping"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/cache"
	"github.com/lncapital/torq/pkg/core"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/lncapital/torq/pkg/vector"
)

var debuglevels = map[string]zerolog.Level{ //nolint:gochecknoglobals
	"panic": zerolog.PanicLevel,
	"fatal": zerolog.FatalLevel,
	"error": zerolog.ErrorLevel,
	"warn":  zerolog.WarnLevel,
	"info":  zerolog.InfoLevel,
	"debug": zerolog.DebugLevel,
	"trace": zerolog.TraceLevel,
}

func main() {

	app := cli.NewApp()
	app.Name = "torq"
	app.EnableBashCompletion = true
	app.Version = build.ExtendedVersion()

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Msgf("error finding home directory of user: %v", err)
	}

	cmdFlags := []cli.Flag{

		// All these flags can be set though a common config file.
		&cli.StringFlag{
			Name:    "config",
			Value:   homedir + "/.torq/torq.conf",
			Aliases: []string{"c"},
			Usage:   "Path to config file",
		},

		// Torq details
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.pprof.path",
			Value: "localhost:6060",
			Usage: "Set pprof path",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.vector.url",
			Value: vector.VectorUrl,
			Usage: "Enable test mode",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.debuglevel",
			Value: "info",
			Usage: "Specify different debuglevels (panic|fatal|error|warn|info|debug|trace)",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.cookie-path",
			Usage: "Path to auth cookie file",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.password",
			Usage: "Password used to access the API and frontend",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.port",
			Value: "8080",
			Usage: "Port to serve the HTTP API",
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:  "torq.no-sub",
			Value: false,
			Usage: "Start the server without subscribing to node data",
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:  "torq.auto-login",
			Value: false,
			Usage: "Allows logging in without a password",
		}),

		// Torq database
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.name",
			Value: "torq",
			Usage: "Name of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.port",
			Value: "5432",
			Usage: "Port of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.host",
			Value: "localhost",
			Usage: "Host of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.user",
			Value: "torq",
			Usage: "Name of the postgres user with access to the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.password",
			Value: "password",
			Usage: "Password used to access the database",
		}),

		// LND connection details
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.url",
			Usage: "Host:Port of the LND node",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.macaroon-path",
			Usage: "Path on disk to LND Macaroon",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.tls-path",
			Usage: "Path on disk to LND TLS file",
		}),
	}

	start := &cli.Command{
		Name:  "start",
		Usage: "Start the main daemon",
		Action: func(c *cli.Context) error {

			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if debuglevel, ok := debuglevels[strings.ToLower(c.String("torq.debuglevel"))]; ok {
				zerolog.SetGlobalLevel(debuglevel)
				log.Debug().Msgf("DebugLevel: %v enabled", debuglevel)
			}

			// Print startup message
			fmt.Printf("Starting Torq %s\n", build.ExtendedVersion())

			fmt.Println("Connecting to the Torq database")
			db, err := database.PgConnect(c.String("db.name"), c.String("db.user"),
				c.String("db.password"), c.String("db.host"), c.String("db.port"))
			if err != nil {
				return errors.Wrap(err, "start cmd")
			}

			defer func() {
				cerr := db.Close()
				if err == nil {
					err = cerr
				}
			}()

			ctxGlobal, cancelGlobal := context.WithCancel(context.Background())
			defer cancelGlobal()

			go cache.ChannelStatesCacheHandler(cache.ChannelStatesCacheChannel, ctxGlobal)
			go cache.SettingsCacheHandle(cache.SettingsCacheChannel, ctxGlobal)
			go cache.NodesCacheHandler(cache.NodesCacheChannel, ctxGlobal)
			go cache.NodeAliasesCacheHandler(cache.NodeAliasesCacheChannel, ctxGlobal)
			go cache.ChannelsCacheHandler(cache.ChannelsCacheChannel, ctxGlobal)
			go cache.TaggedCacheHandler(cache.TaggedCacheChannel, ctxGlobal)
			go cache.TriggersCacheHandler(cache.TriggersCacheChannel, ctxGlobal)
			go tags.TagsCacheHandler(tags.TagsCacheChannel, ctxGlobal)
			go workflows.RebalanceCacheHandler(workflows.RebalancesCacheChannel, ctxGlobal)
			go cache.ServiceCacheHandler(cache.ServicesCacheChannel, ctxGlobal)

			cache.SetVectorUrlBase(c.String("torq.vector.url"))

			cache.InitStates(c.Bool("torq.no-sub"))

			_, cancelRoot := context.WithCancel(ctxGlobal)
			// RootService is equivalent to PID 1 in a unix system
			// Lifecycle:
			// * Inactive (initial state)
			// * Pending (post database migration)
			// * Initializing (post cache initialization)
			// * Active (post desired state initialization from the database)
			// * Inactive again: Torq will panic (catastrophic failure i.e. database migration failed)
			cache.InitRootService(cancelRoot)

			// This function initiates the database migration(s) and parses command line parameters
			// When done the RootService is set to Initialising
			go migrateAndProcessArguments(db, c)

			go servicesMonitor(db)

			if c.String("torq.pprof.path") != "" {
				go pprofStartup(c)
			}

			if err = torqsrv.Start(c.Int("torq.port"), c.String("torq.password"),
				c.String("torq.cookie-path"),
				db, c.Bool("torq.auto-login")); err != nil {
				return errors.Wrap(err, "Starting torq webserver")
			}

			return nil
		},
	}

	migrateUp := &cli.Command{
		Name:  "migrate_up",
		Usage: "Migrates the database to the latest version",
		Action: func(c *cli.Context) error {
			db, err := database.PgConnect(c.String("db.name"), c.String("db.user"),
				c.String("db.password"), c.String("db.host"), c.String("db.port"))
			if err != nil {
				return errors.Wrap(err, "Database connect")
			}

			defer func() {
				cerr := db.Close()
				if err == nil {
					err = cerr
				}
			}()

			err = database.MigrateUp(db)
			if err != nil {
				return errors.Wrap(err, "Migrating database up")
			}

			return nil
		},
	}

	app.Flags = cmdFlags

	app.Before = altsrc.InitInputSourceWithContext(cmdFlags, loadFlags())

	app.Commands = cli.Commands{
		start,
		migrateUp,
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func pprofStartup(c *cli.Context) {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
	runtime.SetCPUProfileRate(1)
	err := http.ListenAndServe(c.String("torq.pprof.path"), nil) //nolint:gosec
	if err != nil {
		log.Error().Err(err).Msg("Torq could not start pprof")
	}
}

func loadFlags() func(context *cli.Context) (altsrc.InputSourceContext, error) {
	return func(context *cli.Context) (altsrc.InputSourceContext, error) {
		if _, err := os.Stat(context.String("config")); err != nil {
			return altsrc.NewMapInputSource("", map[interface{}]interface{}{}), nil
		}
		tomlSource, err := altsrc.NewTomlSourceFromFile(context.String("config"))
		if err != nil {
			return nil, errors.Wrap(err, "Creating new toml config from file")
		}
		return tomlSource, nil
	}
}

func migrateAndProcessArguments(db *sqlx.DB, c *cli.Context) {
	fmt.Println("Checking for migrations..")
	// Check if the database needs to be migrated.
	err := database.MigrateUp(db)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error().Err(err).Msg("Torq could not migrate the database.")
		cache.CancelCoreService(core.RootService)
		cache.SetFailedCoreServiceState(core.RootService)
		return
	}

	for {
		// if node specified on cmd flags then check if we already know about it
		if c.String("lnd.url") != "" && c.String("lnd.macaroon-path") != "" && c.String("lnd.tls-path") != "" {
			macaroonFile, err := os.ReadFile(c.String("lnd.macaroon-path"))
			if err != nil {
				log.Error().Err(err).Msg("Reading macaroon file from disk path from config")
				log.Error().Err(err).Msg("LND is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			tlsFile, err := os.ReadFile(c.String("lnd.tls-path"))
			if err != nil {
				log.Error().Err(err).Msg("Reading tls file from disk path from config")
				log.Error().Err(err).Msg("LND is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			grpcAddress := c.String("lnd.url")
			nodeId, err := settings.GetNodeIdByGRPC(db, grpcAddress)
			if err != nil {
				log.Error().Err(err).Msg("Checking if node specified in config exists")
				log.Error().Err(err).Msg("LND is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			if nodeId == 0 {
				log.Info().Msgf("Node specified in config is not in DB, obtaining public key from GRPC: %v", grpcAddress)
				var nodeConnectionDetails settings.NodeConnectionDetails
				for {
					nodeConnectionDetails, err = settings.AddNodeToDB(db, core.LND, grpcAddress, tlsFile, macaroonFile)
					if err == nil && nodeConnectionDetails.NodeId != 0 {
						break
					} else {
						log.Error().Err(err).Msg("Adding node specified in config to database, LND is probably booting (will retry in 10 seconds)")
						time.Sleep(10 * time.Second)
					}
				}
				nodeConnectionDetails.Name = "Auto configured node"
				nodeConnectionDetails.CustomSettings = core.NodeConnectionDetailCustomSettings(core.NodeConnectionDetailCustomSettingsMax - int(core.ImportFailedPayments))
				_, err = settings.SetNodeConnectionDetails(db, nodeConnectionDetails)
				if err != nil {
					log.Error().Err(err).Msg("Failed to update the node name (cosmetics problem).")
				}
			} else {
				log.Info().Msg("Node specified in config is present, updating Macaroon and TLS files")
				if err = settings.SetNodeConnectionDetailsByConnectionDetails(db, nodeId, core.Active, grpcAddress, tlsFile, macaroonFile); err != nil {
					log.Error().Err(err).Msg("Problem updating node files")
					cache.CancelCoreService(core.RootService)
					cache.SetFailedCoreServiceState(core.RootService)
				}
			}
		}
		break
	}

	cache.SetPendingCoreServiceState(core.RootService)
}

const hangingTimeoutInSeconds = 120
const failureTimeoutInSeconds = 60

func servicesMonitor(db *sqlx.DB) {
	ticker := clock.New().Tick(1 * time.Second)
	for {
		<-ticker

		// Root service ended up in a failed state
		if cache.GetCoreFailedAttemptTime(core.RootService) != nil {
			log.Info().Msg("Torq is dead.")
			panic("RootService cannot be bootstrapped")
		}

		switch cache.GetCurrentCoreServiceState(core.RootService).Status {
		case core.ServicePending:
			log.Info().Msg("Torq is setting up caches.")

			err := settings.InitializeSettingsCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain settings for SettingsCache cache.")
			}

			err = settings.InitializeNodesCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain torq nodes for NodeCache cache.")
			}

			err = settings.InitializeChannelsCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain channels for ChannelCache cache.")
			}

			settings.InitializeNodeAliasesCache(db)

			err = settings.InitializeTaggedCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain tags for TaggedCache cache.")
			}

			err = tags.InitializeTagsCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain tags for TagCache cache.")
			}

			log.Info().Msg("Loading caches in memory.")
			err = corridors.RefreshCorridorCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Torq cannot be initialized (Loading caches in memory).")
			}
			cache.SetInitializingCoreServiceState(core.RootService)
			continue
		case core.ServiceInitializing:
			allGood := true
			for _, coreServiceType := range core.GetCoreServiceTypes() {
				if coreServiceType != core.RootService {
					success := handleCoreServiceStateDelta(db, coreServiceType)
					if !success {
						allGood = false
					}
				}
			}
			if !allGood {
				log.Info().Msg("Torq is initializing.")
				continue
			}
			log.Info().Msg("Torq initialization is done.")
		case core.ServiceActive:
			for _, coreServiceType := range core.GetCoreServiceTypes() {
				handleCoreServiceStateDelta(db, coreServiceType)
			}
		default:
			// We are waiting for the root service to become active
			continue
		}

		// This function actually perform an action (and only once) the first time the RootService becomes active.
		processTorqInitialBoot(db)

		// We end up here when the main Torq service AND all non node specific services have the desired states
		for _, nodeId := range cache.GetLndNodeIds() {
			// check channel events first only if that one works we start the others
			// because channel events downloads our channels and routing policies from LND
			channelEventStream := cache.GetCurrentLndServiceState(core.LndServiceChannelEventStream, nodeId)
			for _, lndServiceType := range core.GetLndServiceTypes() {
				handleLndServiceDelta(db, lndServiceType, nodeId, channelEventStream.Status == core.ServiceActive)
			}
		}
	}
}

func processTorqInitialBoot(db *sqlx.DB) {
	if cache.GetCurrentCoreServiceState(core.RootService).Status != core.ServiceInitializing {
		return
	}
	for _, torqNode := range cache.GetActiveTorqNodeSettings() {
		var grpcAddress string
		var tls []byte
		var macaroon []byte
		var pingSystem core.PingSystem
		var customSettings core.NodeConnectionDetailCustomSettings
		err := db.QueryRow(`
					SELECT grpc_address, tls_data, macaroon_data, ping_system, custom_settings
					FROM node_connection_details
					WHERE node_id=$1`, torqNode.NodeId).Scan(&grpcAddress, &tls, &macaroon, &pingSystem, &customSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Could not obtain desired state for nodeId: %v", torqNode.NodeId)
			continue
		}

		log.Info().Msgf("Torq is setting up the desired states for nodeId: %v.", torqNode.NodeId)

		for _, lndServiceType := range core.GetLndServiceTypes() {
			serviceStatus := core.ServiceActive
			switch lndServiceType {
			case core.VectorService, core.AmbossService:
				if pingSystem&(*lndServiceType.GetPingSystem()) == 0 {
					serviceStatus = core.ServiceInactive
				}
			case core.LndServiceTransactionStream,
				core.LndServiceHtlcEventStream,
				core.LndServiceForwardStream,
				core.LndServiceInvoiceStream,
				core.LndServicePaymentStream:
				active := false
				for _, cs := range lndServiceType.GetNodeConnectionDetailCustomSettings() {
					if customSettings&cs != 0 {
						active = true
						break
					}
				}
				if !active {
					serviceStatus = core.ServiceInactive
				}
			}
			cache.SetDesiredLndServiceState(lndServiceType, torqNode.NodeId, serviceStatus)
			cache.SetLndNodeConnectionDetails(torqNode.NodeId, cache.LndNodeConnectionDetails{
				GRPCAddress:       grpcAddress,
				TLSFileBytes:      tls,
				MacaroonFileBytes: macaroon,
				CustomSettings:    customSettings,
			})
		}
	}
	cache.SetActiveCoreServiceState(core.RootService)
}

func handleLndServiceDelta(db *sqlx.DB, serviceType core.ServiceType, nodeId int, channelEventActive bool) {
	currentState := cache.GetCurrentLndServiceState(serviceType, nodeId)
	desiredState := cache.GetDesiredLndServiceState(serviceType, nodeId)
	if currentState.Status == desiredState.Status {
		return
	}
	switch currentState.Status {
	case core.ServiceActive:
		if desiredState.Status == core.ServiceInactive || !channelEventActive {
			log.Info().Msgf("%v Inactivation for nodeId: %v.", serviceType.String(), nodeId)
			cache.CancelLndService(serviceType, nodeId)
		}
	case core.ServiceInactive:
		if channelEventActive || serviceType == core.LndServiceChannelEventStream {
			bootService(db, serviceType, nodeId)
		}
	case core.ServicePending:
		if !channelEventActive && serviceType != core.LndServiceChannelEventStream {
			log.Info().Msgf("%v Inactivation for nodeId: %v.", serviceType.String(), nodeId)
			cache.CancelLndService(serviceType, nodeId)
			return
		}
		pendingTime := cache.GetLndServiceTime(serviceType, nodeId, core.ServicePending)
		if pendingTime != nil && time.Since(*pendingTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on pending
			cache.CancelLndService(serviceType, nodeId)
		}
	case core.ServiceInitializing:
		if !channelEventActive && serviceType != core.LndServiceChannelEventStream {
			log.Info().Msgf("%v Inactivation for nodeId: %v.", serviceType.String(), nodeId)
			cache.CancelLndService(serviceType, nodeId)
			return
		}
		initializationTime := cache.GetLndServiceTime(serviceType, nodeId, core.ServiceInitializing)
		if initializationTime != nil && time.Since(*initializationTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on initialization
			cache.CancelLndService(serviceType, nodeId)
		}
	}
}

func handleCoreServiceStateDelta(db *sqlx.DB, serviceType core.ServiceType) bool {
	currentState := cache.GetCurrentCoreServiceState(serviceType)
	desiredState := cache.GetDesiredCoreServiceState(serviceType)
	if currentState.Status == desiredState.Status {
		return true
	}
	switch currentState.Status {
	case core.ServiceActive:
		if desiredState.Status == core.ServiceInactive {
			log.Info().Msgf("%v Inactivation.", serviceType.String())
			cache.CancelCoreService(serviceType)
		}
	case core.ServiceInactive:
		bootService(db, serviceType, 0)
	case core.ServicePending:
		pendingTime := cache.GetCoreServiceTime(serviceType, core.ServicePending)
		if pendingTime != nil && time.Since(*pendingTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on pending
			cache.CancelCoreService(serviceType)
		}
	case core.ServiceInitializing:
		initializationTime := cache.GetCoreServiceTime(serviceType, core.ServiceInitializing)
		if initializationTime != nil && time.Since(*initializationTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on initialization
			cache.CancelCoreService(serviceType)
		}
	}
	return false
}

func bootService(db *sqlx.DB, serviceType core.ServiceType, nodeId int) {
	var failedAttemptTime *time.Time
	if nodeId == 0 {
		failedAttemptTime = cache.GetCoreFailedAttemptTime(serviceType)
	} else {
		failedAttemptTime = cache.GetLndFailedAttemptTime(serviceType, nodeId)
	}
	if failedAttemptTime != nil && time.Since(*failedAttemptTime).Seconds() < failureTimeoutInSeconds {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	if nodeId == 0 {
		log.Info().Msgf("%v boot attempt.", serviceType.String())
		cache.InitCoreServiceState(serviceType, cancel)
	} else {
		log.Info().Msgf("%v boot attempt for nodeId: %v.", serviceType.String(), nodeId)
		cache.InitLndServiceState(serviceType, nodeId, cancel)
	}

	if !isBootable(serviceType, nodeId) {
		return
	}

	var conn *grpc.ClientConn
	var err error
	if serviceType.IsLndService() {
		nodeConnectionDetails := cache.GetLndNodeConnectionDetails(nodeId)
		conn, err = lnd_connect.Connect(
			nodeConnectionDetails.GRPCAddress,
			nodeConnectionDetails.TLSFileBytes,
			nodeConnectionDetails.MacaroonFileBytes)
		if err != nil {
			log.Error().Err(err).Msgf("%v failed to connect to lnd for node id: %v", serviceType.String(), nodeId)
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return
		}
	case commons.TelegramHighService:
		if commons.GetSettings().GetTelegramCredential(true) == "" {
			commons.SetInactiveTorqServiceState(serviceType)
			commons.SetDesiredTorqServiceState(serviceType, commons.ServiceInactive)
			log.Info().Msgf("%v Service deactivated since there are no credentials", serviceType.String())
			return
		}
	case commons.TelegramLowService:
		if commons.GetSettings().GetTelegramCredential(false) == "" {
			commons.SetInactiveTorqServiceState(serviceType)
			commons.SetDesiredTorqServiceState(serviceType, commons.ServiceInactive)
			log.Info().Msgf("%v Service deactivated since there are no credentials", serviceType.String())
			return
		}
	case commons.SlackService:
		oauth, botToken := commons.GetSettings().GetSlackCredential()
		if oauth == "" || botToken == "" {
			commons.SetInactiveTorqServiceState(serviceType)
			commons.SetDesiredTorqServiceState(serviceType, commons.ServiceInactive)
			log.Info().Msgf("%v Service deactivated since there are no credentials", serviceType.String())
			return
		}
	case commons.NotifierService:
		oauth, botToken := commons.GetSettings().GetSlackCredential()
		if (oauth == "" || botToken == "") &&
			commons.GetSettings().GetTelegramCredential(true) == "" &&
			commons.GetSettings().GetTelegramCredential(false) == "" {
			commons.SetInactiveTorqServiceState(serviceType)
			commons.SetDesiredTorqServiceState(serviceType, commons.ServiceInactive)
			log.Info().Msgf("%v Service deactivated since there are no credentials", serviceType.String())
			return
		}
	}

	log.Info().Msgf("%v Service booted for nodeId: %v", serviceType.String(), nodeId)
	switch serviceType {
	// NOT NODE ID SPECIFIC
	case core.AutomationChannelBalanceEventTriggerService:
		go services.StartChannelBalanceEventService(ctx, db)
	case core.AutomationChannelEventTriggerService:
		go services.StartChannelEventService(ctx, db)
	case core.AutomationIntervalTriggerService:
		go services.StartIntervalService(ctx, db)
	case core.AutomationScheduledTriggerService:
		go services.StartScheduledService(ctx, db)
	case core.MaintenanceService:
		go services.StartMaintenanceService(ctx, db)
	case core.CronService:
		go services.StartCronService(ctx, db)
	case commons.NotifierService:
		go notifications.StartNotifier(ctx, db)
	case commons.SlackService:
		go notifications.StartSlackListener(ctx, db)
	case commons.TelegramHighService:
		go notifications.StartTelegramListeners(ctx, db, true)
	case commons.TelegramLowService:
		go notifications.StartTelegramListeners(ctx, db, false)
	// NODE SPECIFIC
	case core.VectorService:
		go vector_ping.Start(ctx, conn, nodeId)
	case core.AmbossService:
		go amboss_ping.Start(ctx, conn, nodeId)
	case core.RebalanceService:
		go services.StartRebalanceService(ctx, conn, db, nodeId)
	case core.LndServiceChannelEventStream:
		go subscribe.StartChannelEventStream(ctx, conn, db, nodeId)
	case core.LndServiceGraphEventStream:
		go subscribe.StartGraphEventStream(ctx, conn, db, nodeId)
	case core.LndServiceTransactionStream:
		go subscribe.StartTransactionStream(ctx, conn, db, nodeId)
	case core.LndServiceHtlcEventStream:
		go subscribe.StartHtlcEvents(ctx, conn, db, nodeId)
	case core.LndServiceForwardStream:
		go subscribe.StartForwardStream(ctx, conn, db, nodeId)
	case core.LndServiceInvoiceStream:
		go subscribe.StartInvoiceStream(ctx, conn, db, nodeId)
	case core.LndServicePaymentStream:
		go subscribe.StartPaymentStream(ctx, conn, db, nodeId)
	case core.LndServicePeerEventStream:
		go subscribe.StartPeerEvents(ctx, conn, db, nodeId)
	case core.LndServiceInFlightPaymentStream:
		go subscribe.StartInFlightPaymentStream(ctx, conn, db, nodeId)
	case core.LndServiceChannelBalanceCacheStream:
		go subscribe.StartChannelBalanceCacheMaintenance(ctx, conn, db, nodeId)
	}
}

func isBootable(serviceType core.ServiceType, nodeId int) bool {
	switch serviceType {
	case core.VectorService, core.AmbossService, core.RebalanceService,
		core.LndServiceChannelEventStream,
		core.LndServiceGraphEventStream,
		core.LndServiceTransactionStream,
		core.LndServiceHtlcEventStream,
		core.LndServiceForwardStream,
		core.LndServiceInvoiceStream,
		core.LndServicePaymentStream,
		core.LndServicePeerEventStream,
		core.LndServiceInFlightPaymentStream,
		core.LndServiceChannelBalanceCacheStream:
		nodeConnectionDetails := cache.GetLndNodeConnectionDetails(nodeId)
		if nodeConnectionDetails.GRPCAddress == "" ||
			nodeConnectionDetails.MacaroonFileBytes == nil ||
			len(nodeConnectionDetails.MacaroonFileBytes) == 0 ||
			nodeConnectionDetails.TLSFileBytes == nil ||
			len(nodeConnectionDetails.TLSFileBytes) == 0 {
			log.Error().Msgf("%v failed to get connection details for node id: %v", serviceType.String(), nodeId)
			cache.SetFailedLndServiceState(serviceType, nodeId)
			return false
		}
	}
	return true
}
