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
	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/services_core"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/internal/vector"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/cln_connect"
	"github.com/lncapital/torq/pkg/lnd_connect"
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

		// CLN connection details
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "cln.url",
			Usage: "Host:Port of the CLN node",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "cln.certificate-path",
			Usage: "Path on disk to CLN client certificate file",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "cln.key-path",
			Usage: "Path on disk to CLN client key file",
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
		cache.CancelCoreService(services_core.RootService)
		cache.SetFailedCoreServiceState(services_core.RootService)
		return
	}

	for {
		// if node specified on cmd flags then check if we already know about it
		if c.String("lnd.url") != "" &&
			c.String("lnd.macaroon-path") != "" &&
			c.String("lnd.tls-path") != "" {

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
				log.Info().Msgf(
					"Node specified in config is not in DB, obtaining public key from GRPC: %v", grpcAddress)
				var nodeConnectionDetails settings.NodeConnectionDetails
				for {
					nodeConnectionDetails, err = settings.AddNodeToDB(db, core.LND, grpcAddress, tlsFile, macaroonFile)
					if err == nil && nodeConnectionDetails.NodeId != 0 {
						break
					} else {
						log.Error().Err(err).Msg("Adding node specified in config to database, " +
							"LND is probably booting (will retry in 10 seconds)")
						time.Sleep(10 * time.Second)
					}
				}
				nodeConnectionDetails.Name = "Auto configured node"
				nodeConnectionDetails.CustomSettings = core.NodeConnectionDetailCustomSettings(
					core.NodeConnectionDetailCustomSettingsMax - int(core.ImportFailedPayments))
				_, err = settings.SetNodeConnectionDetails(db, nodeConnectionDetails)
				if err != nil {
					log.Error().Err(err).Msg("Failed to update the node name (cosmetics problem).")
				}
			} else {
				log.Info().Msg("Node specified in config is present, updating Macaroon and TLS files")
				err = settings.SetNodeConnectionDetailsByConnectionDetails(
					db, nodeId, core.Active, core.LND, grpcAddress, tlsFile, macaroonFile)
				if err != nil {
					log.Error().Err(err).Msg("Problem updating node files")
					cache.CancelCoreService(services_core.RootService)
					cache.SetFailedCoreServiceState(services_core.RootService)
				}
			}
		}
		// if node specified on cmd flags then check if we already know about it
		if c.String("cln.url") != "" &&
			c.String("cln.certificate-path") != "" &&
			c.String("cln.key-path") != "" {

			certificate, err := os.ReadFile(c.String("cln.certificate-path"))
			if err != nil {
				log.Error().Err(err).Msg("Reading certificate file from disk path from config")
				log.Error().Err(err).Msg("CLN is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			key, err := os.ReadFile(c.String("cln.key-path"))
			if err != nil {
				log.Error().Err(err).Msg("Reading key file from disk path from config")
				log.Error().Err(err).Msg("CLN is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			grpcAddress := c.String("cln.url")
			nodeId, err := settings.GetNodeIdByGRPC(db, grpcAddress)
			if err != nil {
				log.Error().Err(err).Msg("Checking if node specified in config exists")
				log.Error().Err(err).Msg("CLN is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			if nodeId == 0 {
				log.Info().Msgf(
					"Node specified in config is not in DB, obtaining public key from GRPC: %v", grpcAddress)
				var nodeConnectionDetails settings.NodeConnectionDetails
				for {
					nodeConnectionDetails, err = settings.AddNodeToDB(db, core.CLN, grpcAddress, certificate, key)
					if err == nil && nodeConnectionDetails.NodeId != 0 {
						break
					} else {
						log.Error().Err(err).Msg("Adding node specified in config to database, " +
							"CLN is probably booting (will retry in 10 seconds)")
						time.Sleep(10 * time.Second)
					}
				}
				nodeConnectionDetails.Name = "Auto configured node"
				nodeConnectionDetails.CustomSettings = core.NodeConnectionDetailCustomSettings(
					core.NodeConnectionDetailCustomSettingsMax - int(core.ImportFailedPayments))
				_, err = settings.SetNodeConnectionDetails(db, nodeConnectionDetails)
				if err != nil {
					log.Error().Err(err).Msg("Failed to update the node name (cosmetics problem).")
				}
			} else {
				log.Info().Msg("Node specified in config is present, updating Certificate and Key files")
				err = settings.SetNodeConnectionDetailsByConnectionDetails(
					db, nodeId, core.Active, core.CLN, grpcAddress, certificate, key)
				if err != nil {
					log.Error().Err(err).Msg("Problem updating node files")
					cache.CancelCoreService(services_core.RootService)
					cache.SetFailedCoreServiceState(services_core.RootService)
				}
			}
		}
		break
	}

	cache.SetPendingCoreServiceState(services_core.RootService)
}

const hangingTimeoutInSeconds = 120
const failureTimeoutInSeconds = 60

func servicesMonitor(db *sqlx.DB) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C

		// Root service ended up in a failed state
		if cache.GetCoreFailedAttemptTime(services_core.RootService) != nil {
			log.Info().Msg("Torq is dead.")
			panic("RootService cannot be bootstrapped")
		}

		switch cache.GetCurrentCoreServiceState(services_core.RootService).Status {
		case services_core.Pending:
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
			cache.SetInitializingCoreServiceState(services_core.RootService)
			continue
		case services_core.Initializing:
			allGood := true
			for _, coreServiceType := range services_core.GetCoreServiceTypes() {
				if coreServiceType != services_core.RootService {
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
		case services_core.Active:
			for _, coreServiceType := range services_core.GetCoreServiceTypes() {
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
			channelEventStream := cache.GetCurrentNodeServiceState(services_core.LndServiceChannelEventStream, nodeId)
			for _, lndServiceType := range services_core.GetLndServiceTypes() {
				handleNodeServiceDelta(db, lndServiceType, nodeId, channelEventStream.Status == services_core.Active)
			}
		}

		for _, nodeId := range cache.GetClnNodeIds() {
			// check peers first only if that one works we start the others
			// because peers downloads our channels and routing policies from CLN
			channelEventStream := cache.GetCurrentNodeServiceState(services_core.ClnServicePeersService, nodeId)
			for _, clnServiceType := range services_core.GetClnServiceTypes() {
				handleNodeServiceDelta(db, clnServiceType, nodeId, channelEventStream.Status == services_core.Active)
			}
		}
	}
}

func processTorqInitialBoot(db *sqlx.DB) {
	if cache.GetCurrentCoreServiceState(services_core.RootService).Status != services_core.Initializing {
		return
	}
	for _, torqNode := range cache.GetActiveTorqNodeSettings() {
		var implementation core.Implementation
		var grpcAddress string
		var tls []byte
		var macaroon []byte
		var certificate []byte
		var key []byte
		var pingSystem core.PingSystem
		var customSettings core.NodeConnectionDetailCustomSettings
		err := db.QueryRow(`
					SELECT implementation, grpc_address,
					       tls_data, macaroon_data, certificate_data, key_data,
					       ping_system, custom_settings
					FROM node_connection_details
					WHERE node_id=$1`, torqNode.NodeId).Scan(&implementation, &grpcAddress,
			&tls, &macaroon, &certificate, &key,
			&pingSystem, &customSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Could not obtain desired state for nodeId: %v", torqNode.NodeId)
			continue
		}

		log.Info().Msgf("Torq is setting up the desired states for nodeId: %v.", torqNode.NodeId)

		switch implementation {
		case core.LND:
			for _, lndServiceType := range services_core.GetLndServiceTypes() {
				serviceStatus := services_core.Active
				switch lndServiceType {
				case services_core.LndServiceVectorService, services_core.LndServiceAmbossService:
					if pingSystem&(*lndServiceType.GetPingSystem()) == 0 {
						serviceStatus = services_core.Inactive
					}
				case services_core.LndServiceTransactionStream,
					services_core.LndServiceHtlcEventStream,
					services_core.LndServiceForwardsService,
					services_core.LndServiceInvoiceStream,
					services_core.LndServicePaymentsService:
					active := false
					for _, cs := range lndServiceType.GetNodeConnectionDetailCustomSettings() {
						if customSettings&cs != 0 {
							active = true
							break
						}
					}
					if !active {
						serviceStatus = services_core.Inactive
					}
				}
				cache.SetDesiredNodeServiceState(lndServiceType, torqNode.NodeId, serviceStatus)
			}
		case core.CLN:
			for _, clnServiceType := range services_core.GetClnServiceTypes() {
				serviceStatus := services_core.Active
				switch clnServiceType {
				case services_core.ClnServiceVectorService, services_core.ClnServiceAmbossService:
					if pingSystem&(*clnServiceType.GetPingSystem()) == 0 {
						serviceStatus = services_core.Inactive
					}
				}
				cache.SetDesiredNodeServiceState(clnServiceType, torqNode.NodeId, serviceStatus)
			}
		}
		cache.SetNodeConnectionDetails(torqNode.NodeId, cache.NodeConnectionDetails{
			Implementation:       implementation,
			GRPCAddress:          grpcAddress,
			TLSFileBytes:         tls,
			MacaroonFileBytes:    macaroon,
			CertificateFileBytes: certificate,
			KeyFileBytes:         key,
			CustomSettings:       customSettings,
		})
	}
	cache.SetActiveCoreServiceState(services_core.RootService)
}

func handleNodeServiceDelta(db *sqlx.DB,
	serviceType services_core.ServiceType,
	nodeId int,
	channelEventActive bool) {

	currentState := cache.GetCurrentNodeServiceState(serviceType, nodeId)
	desiredState := cache.GetDesiredNodeServiceState(serviceType, nodeId)
	if currentState.Status == desiredState.Status {
		return
	}
	switch currentState.Status {
	case services_core.Active:
		if desiredState.Status == services_core.Inactive || !channelEventActive {
			log.Info().Msgf("%v Inactivation for nodeId: %v.", serviceType.String(), nodeId)
			cache.CancelNodeService(serviceType, nodeId)
		}
	case services_core.Inactive:
		if channelEventActive ||
			serviceType == services_core.LndServiceChannelEventStream ||
			serviceType == services_core.ClnServicePeersService {

			bootService(db, serviceType, nodeId)
		}
	case services_core.Pending:
		if !channelEventActive &&
			serviceType != services_core.LndServiceChannelEventStream &&
			serviceType != services_core.ClnServicePeersService {

			log.Info().Msgf("%v Inactivation for nodeId: %v.", serviceType.String(), nodeId)
			cache.CancelNodeService(serviceType, nodeId)
			return
		}
		pendingTime := cache.GetNodeServiceTime(serviceType, nodeId, services_core.Pending)
		if pendingTime != nil && time.Since(*pendingTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on pending
			cache.CancelNodeService(serviceType, nodeId)
		}
	case services_core.Initializing:
		if !channelEventActive &&
			serviceType != services_core.LndServiceChannelEventStream &&
			serviceType != services_core.ClnServicePeersService {

			log.Info().Msgf("%v Inactivation for nodeId: %v.", serviceType.String(), nodeId)
			cache.CancelNodeService(serviceType, nodeId)
			return
		}
		initializationTime := cache.GetNodeServiceTime(serviceType, nodeId, services_core.Initializing)
		if initializationTime != nil && time.Since(*initializationTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on initialization
			cache.CancelNodeService(serviceType, nodeId)
		}
	}
}

func handleCoreServiceStateDelta(db *sqlx.DB, serviceType services_core.ServiceType) bool {
	currentState := cache.GetCurrentCoreServiceState(serviceType)
	desiredState := cache.GetDesiredCoreServiceState(serviceType)
	if currentState.Status == desiredState.Status {
		return true
	}
	switch currentState.Status {
	case services_core.Active:
		if desiredState.Status == services_core.Inactive {
			log.Info().Msgf("%v Inactivation.", serviceType.String())
			cache.CancelCoreService(serviceType)
		}
	case services_core.Inactive:
		bootService(db, serviceType, 0)
	case services_core.Pending:
		pendingTime := cache.GetCoreServiceTime(serviceType, services_core.Pending)
		if pendingTime != nil && time.Since(*pendingTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on pending
			cache.CancelCoreService(serviceType)
		}
	case services_core.Initializing:
		initializationTime := cache.GetCoreServiceTime(serviceType, services_core.Initializing)
		if initializationTime != nil && time.Since(*initializationTime).Seconds() > hangingTimeoutInSeconds {
			// hanging idle on initialization
			cache.CancelCoreService(serviceType)
		}
	}
	return false
}

func bootService(db *sqlx.DB, serviceType services_core.ServiceType, nodeId int) {
	var failedAttemptTime *time.Time
	if nodeId == 0 {
		failedAttemptTime = cache.GetCoreFailedAttemptTime(serviceType)
	} else {
		failedAttemptTime = cache.GetNodeFailedAttemptTime(serviceType, nodeId)
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
		cache.InitNodeServiceState(serviceType, nodeId, cancel)
	}

	if !isBootable(serviceType, nodeId) {
		return
	}

	var conn *grpc.ClientConn
	var err error
	implementation := serviceType.GetImplementation()
	if implementation != nil {
		switch *implementation {
		case core.LND:
			nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
			conn, err = lnd_connect.Connect(
				nodeConnectionDetails.GRPCAddress,
				nodeConnectionDetails.TLSFileBytes,
				nodeConnectionDetails.MacaroonFileBytes)
			if err != nil {
				log.Error().Err(err).Msgf("%v failed to connect for node id: %v", serviceType.String(), nodeId)
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
		case core.CLN:
			nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
			conn, err = cln_connect.Connect(
				nodeConnectionDetails.GRPCAddress,
				nodeConnectionDetails.CertificateFileBytes,
				nodeConnectionDetails.KeyFileBytes)
			if err != nil {
				log.Error().Err(err).Msgf("%v failed to connect for node id: %v", serviceType.String(), nodeId)
				cache.SetFailedNodeServiceState(serviceType, nodeId)
				return
			}
		}
	}

	log.Info().Msgf("%v service booted for nodeId: %v", serviceType.String(), nodeId)
	switch serviceType {
	// NOT NODE ID SPECIFIC
	case services_core.AutomationChannelBalanceEventTriggerService:
		go services.StartChannelBalanceEventService(ctx, db)
	case services_core.AutomationChannelEventTriggerService:
		go services.StartChannelEventService(ctx, db)
	case services_core.AutomationIntervalTriggerService:
		go services.StartIntervalService(ctx, db)
	case services_core.AutomationScheduledTriggerService:
		go services.StartScheduledService(ctx, db)
	case services_core.MaintenanceService:
		go services.StartMaintenanceService(ctx, db)
	case services_core.CronService:
		go services.StartCronService(ctx, db)
	case services_core.NotifierService:
		go notifications.StartNotifier(ctx, db)
	case services_core.SlackService:
		go notifications.StartSlackListener(ctx, db)
	case services_core.TelegramHighService:
		go notifications.StartTelegramListeners(ctx, db, true)
	case services_core.TelegramLowService:
		go notifications.StartTelegramListeners(ctx, db, false)
	// LND NODE SPECIFIC
	case services_core.LndServiceVectorService:
		go vector_ping.Start(ctx, conn, core.LND, nodeId)
	case services_core.LndServiceAmbossService:
		go amboss_ping.Start(ctx, conn, core.LND, nodeId)
	case services_core.LndServiceRebalanceService:
		go services.StartRebalanceService(ctx, conn, db, nodeId)
	case services_core.LndServiceChannelEventStream:
		go subscribe.StartChannelEventStream(ctx, conn, db, nodeId)
	case services_core.LndServiceGraphEventStream:
		go subscribe.StartGraphEventStream(ctx, conn, db, nodeId)
	case services_core.LndServiceTransactionStream:
		go subscribe.StartTransactionStream(ctx, conn, db, nodeId)
	case services_core.LndServiceHtlcEventStream:
		go subscribe.StartHtlcEvents(ctx, conn, db, nodeId)
	case services_core.LndServiceForwardsService:
		go subscribe.StartForwardsService(ctx, conn, db, nodeId)
	case services_core.LndServiceInvoiceStream:
		go subscribe.StartInvoiceStream(ctx, conn, db, nodeId)
	case services_core.LndServicePaymentsService:
		go subscribe.StartPaymentsService(ctx, conn, db, nodeId)
	case services_core.LndServicePeerEventStream:
		go subscribe.StartPeerEvents(ctx, conn, db, nodeId)
	case services_core.LndServiceInFlightPaymentsService:
		go subscribe.StartInFlightPaymentsService(ctx, conn, db, nodeId)
	case services_core.LndServiceChannelBalanceCacheService:
		go subscribe.StartChannelBalanceCacheMaintenance(ctx, conn, db, nodeId)
	// CLN NODE SPECIFIC
	case services_core.ClnServiceVectorService:
		go vector_ping.Start(ctx, conn, core.CLN, nodeId)
	case services_core.ClnServiceAmbossService:
		go amboss_ping.Start(ctx, conn, core.CLN, nodeId)
	case services_core.ClnServicePeersService:
		go subscribe.StartPeersService(ctx, conn, db, nodeId)
	case services_core.ClnServiceChannelsService:
		go subscribe.StartChannelsService(ctx, conn, db, nodeId)
	case services_core.ClnServiceFundsService:
		go subscribe.StartFundsService(ctx, conn, db, nodeId)
	case services_core.ClnServiceNodesService:
		go subscribe.StartNodesService(ctx, conn, db, nodeId)
	}
}

func isBootable(serviceType services_core.ServiceType, nodeId int) bool {
	switch serviceType {
	case services_core.LndServiceVectorService, services_core.LndServiceAmbossService, services_core.LndServiceRebalanceService,
		services_core.LndServiceChannelEventStream,
		services_core.LndServiceGraphEventStream,
		services_core.LndServiceTransactionStream,
		services_core.LndServiceHtlcEventStream,
		services_core.LndServiceForwardsService,
		services_core.LndServiceInvoiceStream,
		services_core.LndServicePaymentsService,
		services_core.LndServicePeerEventStream,
		services_core.LndServiceInFlightPaymentsService,
		services_core.LndServiceChannelBalanceCacheService:
		nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
		if nodeConnectionDetails.Implementation == core.LND &&
			(nodeConnectionDetails.GRPCAddress == "" ||
				nodeConnectionDetails.MacaroonFileBytes == nil ||
				len(nodeConnectionDetails.MacaroonFileBytes) == 0 ||
				nodeConnectionDetails.TLSFileBytes == nil ||
				len(nodeConnectionDetails.TLSFileBytes) == 0) {
			log.Error().Msgf("%v failed to get connection details for node id: %v", serviceType.String(), nodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return false
		}
	case services_core.ClnServiceVectorService, services_core.ClnServiceAmbossService,
		services_core.ClnServicePeersService,
		services_core.ClnServiceChannelsService,
		services_core.ClnServiceFundsService,
		services_core.ClnServiceNodesService:
		nodeConnectionDetails := cache.GetNodeConnectionDetails(nodeId)
		if nodeConnectionDetails.Implementation == core.CLN &&
			(nodeConnectionDetails.GRPCAddress == "" ||
				nodeConnectionDetails.CertificateFileBytes == nil ||
				len(nodeConnectionDetails.CertificateFileBytes) == 0 ||
				nodeConnectionDetails.KeyFileBytes == nil ||
				len(nodeConnectionDetails.KeyFileBytes) == 0) {
			log.Error().Msgf("%v failed to get connection details for node id: %v", serviceType.String(), nodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeId)
			return false
		}
	case services_core.TelegramHighService:
		if cache.GetSettings().GetTelegramCredential(true) == "" {
			cache.SetInactiveCoreServiceState(serviceType)
			cache.SetDesiredCoreServiceState(serviceType, services_core.Inactive)
			log.Info().Msgf("%v service deactivated since there are no credentials", serviceType.String())
			return false
		}
	case services_core.TelegramLowService:
		if cache.GetSettings().GetTelegramCredential(false) == "" {
			cache.SetInactiveCoreServiceState(serviceType)
			cache.SetDesiredCoreServiceState(serviceType, services_core.Inactive)
			log.Info().Msgf("%v service deactivated since there are no credentials", serviceType.String())
			return false
		}
	case services_core.SlackService:
		oauth, botToken := cache.GetSettings().GetSlackCredential()
		if oauth == "" || botToken == "" {
			cache.SetInactiveCoreServiceState(serviceType)
			cache.SetDesiredCoreServiceState(serviceType, services_core.Inactive)
			log.Info().Msgf("%v service deactivated since there are no credentials", serviceType.String())
			return false
		}
	case services_core.NotifierService:
		oauth, botToken := cache.GetSettings().GetSlackCredential()
		if (oauth == "" || botToken == "") &&
			cache.GetSettings().GetTelegramCredential(true) == "" &&
			cache.GetSettings().GetTelegramCredential(false) == "" {
			cache.SetInactiveCoreServiceState(serviceType)
			cache.SetDesiredCoreServiceState(serviceType, services_core.Inactive)
			log.Info().Msgf("%v Service deactivated since there are no credentials", serviceType.String())
			return false
		}
	}
	return true
}
