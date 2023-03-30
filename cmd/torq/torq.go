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
	"github.com/lncapital/torq/cmd/torq/internal/services"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/cmd/torq/internal/vector_ping"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/commons"
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
			Value: "localhost:6060",
			Usage: "Set pprof path",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.vector.url",
			Value: commons.VectorUrl,
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

			go commons.ManagedChannelGroupCache(commons.ManagedChannelGroupChannel, ctxGlobal)
			go commons.ManagedChannelStateCache(commons.ManagedChannelStateChannel, ctxGlobal)
			go commons.ManagedSettingsCache(commons.ManagedSettingsChannel, ctxGlobal)
			go commons.ManagedNodeCache(commons.ManagedNodeChannel, ctxGlobal)
			go commons.ManagedNodeAliasCache(commons.ManagedNodeAliasChannel, ctxGlobal)
			go commons.ManagedChannelCache(commons.ManagedChannelChannel, ctxGlobal)
			go commons.ManagedTaggedCache(commons.ManagedTaggedChannel, ctxGlobal)
			go commons.ManagedTriggerCache(commons.ManagedTriggerChannel, ctxGlobal)
			go tags.ManagedTagCache(tags.ManagedTagChannel, ctxGlobal)
			go workflows.ManagedRebalanceCache(workflows.ManagedRebalanceChannel, ctxGlobal)
			go commons.ManagedServiceCache(commons.ManagedServiceChannel, ctxGlobal)

			commons.SetVectorUrlBase(c.String("torq.vector.url"))

			commons.InitStates(c.Bool("torq.no-sub"))

			_, cancelTorq := context.WithCancel(ctxGlobal)
			// TorqService is equivalent to PID 1 in a unix system
			// Lifecycle:
			// * Inactive (initial state)
			// * Pending (post database migration)
			// * Initializing (post cache initialization)
			// * Active (post desired state initialization from the database)
			// * Inactive again: Torq will panic (catastrophic failure i.e. database migration failed)
			commons.InitTorqService(cancelTorq)

			// This function initiates the database migration(s) and parses command line parameters
			// When done the TorqService is set to Initialising
			go migrateAndProcessArguments(db, c)

			go manageServices(db)

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
		commons.SetInactiveTorqServiceState(commons.TorqService)
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
					nodeConnectionDetails, err = settings.AddNodeToDB(db, commons.LND, grpcAddress, tlsFile, macaroonFile)
					if err == nil && nodeConnectionDetails.NodeId != 0 {
						break
					} else {
						log.Error().Err(err).Msg("Adding node specified in config to database, LND is probably booting (will retry in 10 seconds)")
						time.Sleep(10 * time.Second)
					}
				}
				nodeConnectionDetails.Name = "Auto configured node"
				nodeConnectionDetails.CustomSettings = commons.NodeConnectionDetailCustomSettings(commons.NodeConnectionDetailCustomSettingsMax - int(commons.ImportFailedPayments))
				_, err = settings.SetNodeConnectionDetails(db, nodeConnectionDetails)
				if err != nil {
					log.Error().Err(err).Msg("Failed to update the node name (cosmetics problem).")
				}
			} else {
				log.Info().Msg("Node specified in config is present, updating Macaroon and TLS files")
				if err = settings.SetNodeConnectionDetailsByConnectionDetails(db, nodeId, commons.Active, grpcAddress, tlsFile, macaroonFile); err != nil {
					log.Error().Err(err).Msg("Problem updating node files")
					commons.SetInactiveTorqServiceState(commons.TorqService)
				}
			}
		}
		break
	}

	commons.SetPendingTorqServiceState(commons.TorqService)
}

func manageServices(db *sqlx.DB) {
	ticker := clock.New().Tick(1 * time.Second)
	for {
		<-ticker

		// Torq was booting but is now inactive so bootstrapping failed
		if commons.GetTorqFailedAttemptTime(commons.TorqService) != nil {
			log.Info().Msg("Torq is dead.")
			panic("TorqService cannot be bootstrapped")
		}

		switch commons.GetCurrentTorqServiceState(commons.TorqService).Status {
		case commons.ServiceInitializing:
			allGood := true
			for _, torqServiceType := range commons.GetTorqServiceTypes() {
				if torqServiceType != commons.TorqService {
					success := processTorqService(db, torqServiceType)
					if !success {
						allGood = false
					}
				}
			}
			if !allGood {
				log.Info().Msg("Torq initialization is done.")
				continue
			}
			log.Info().Msg("Torq is initializing.")
		case commons.ServicePending:
			log.Info().Msg("Torq is setting up caches.")

			err := settings.InitializeManagedSettingsCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain settings for ManagedSettings cache.")
			}

			err = settings.InitializeManagedNodeCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain torq nodes for ManagedNode cache.")
			}

			err = settings.InitializeManagedChannelCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain channels for ManagedChannel cache.")
			}

			settings.InitializeManagedNodeAliasCache(db)

			err = settings.InitializeManagedTaggedCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain tags for ManagedTagged cache.")
			}

			err = tags.InitializeManagedTagCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain tags for ManagedTag cache.")
			}

			log.Info().Msg("Loading caches in memory.")
			err = corridors.RefreshCorridorCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Torq cannot be initialized (Loading caches in memory).")
			}
			commons.SetInitializingTorqServiceState(commons.TorqService)
			continue
		case commons.ServiceActive:
			// All is good
		default:
			// We are waiting for the Torq service to become active
			continue
		}

		// This function actually perform an action (and only once) the first time the TorqService becomes active.
		proccessTorqInitialBoot(db)

		// We end up here when the main Torq service AND all non node specific services have the desired states
		for _, nodeId := range commons.GetLndNodeIds() {
			// check channel events first only if that one works we start the others
			// because channel events downloads our channels and routing policies from LND
			channelEventStream := commons.GetCurrentLndServiceState(commons.LndServiceChannelEventStream, nodeId)
			switch channelEventStream.Status {
			case commons.ServiceActive:
				for _, lndServiceType := range commons.GetLndServiceTypes() {
					processLndService(db, lndServiceType, nodeId)
				}
			case commons.ServiceInactive:
				processLndService(db, commons.LndServiceChannelEventStream, nodeId)
			}
		}
	}
}

func proccessTorqInitialBoot(db *sqlx.DB) {
	if commons.GetCurrentTorqServiceState(commons.TorqService).Status != commons.ServiceInitializing {
		return
	}
	for _, torqNode := range commons.GetActiveTorqNodeSettings() {
		var grpcAddress string
		var tls []byte
		var macaroon []byte
		var pingSystem commons.PingSystem
		var customSettings commons.NodeConnectionDetailCustomSettings
		err := db.QueryRow(`
					SELECT grpc_address, tls_data, macaroon_data, ping_system, custom_settings
					FROM node_connection_details
					WHERE node_id=$1`, torqNode.NodeId).Scan(&grpcAddress, &tls, &macaroon, &pingSystem, &customSettings)
		if err != nil {
			log.Error().Err(err).Msgf("Could not obtain desired state for nodeId: %v", torqNode.NodeId)
			continue
		}

		log.Info().Msgf("Torq is setting up the desired states for nodeId: %v.", torqNode.NodeId)

		for _, lndServiceType := range commons.GetLndServiceTypes() {
			serviceStatus := commons.ServiceActive
			switch lndServiceType {
			case commons.VectorService:
				if pingSystem&commons.Vector == 0 {
					serviceStatus = commons.ServiceInactive
				}
			case commons.AmbossService:
				if pingSystem&commons.Amboss == 0 {
					serviceStatus = commons.ServiceInactive
				}
			case commons.LndServiceTransactionStream:
				if customSettings&commons.ImportTransactions == 0 {
					serviceStatus = commons.ServiceInactive
				}
			case commons.LndServiceHtlcEventStream:
				if customSettings&commons.ImportHtlcEvents == 0 {
					serviceStatus = commons.ServiceInactive
				}
			case commons.LndServiceForwardStream:
				if customSettings&commons.ImportForwards == 0 {
					serviceStatus = commons.ServiceInactive
				}
			case commons.LndServiceInvoiceStream:
				if customSettings&commons.ImportInvoices == 0 {
					serviceStatus = commons.ServiceInactive
				}
			case commons.LndServicePaymentStream:
				if customSettings&commons.ImportPayments == 0 {
					serviceStatus = commons.ServiceInactive
				}
			case commons.LndServicePeerEventStream:
				if customSettings&commons.ImportPeerEvents == 0 {
					serviceStatus = commons.ServiceInactive
				}
			}
			commons.SetDesiredLndServiceState(lndServiceType, torqNode.NodeId, serviceStatus)
			commons.SetLndNodeConnectionDetails(torqNode.NodeId, commons.LndNodeConnectionDetails{
				GRPCAddress:       grpcAddress,
				TLSFileBytes:      tls,
				MacaroonFileBytes: macaroon,
				CustomSettings:    customSettings,
			})
		}
	}
	commons.SetActiveTorqServiceState(commons.TorqService)
}

func processLndService(db *sqlx.DB, serviceType commons.ServiceType, nodeId int) {
	currentState := commons.GetCurrentLndServiceState(serviceType, nodeId)
	desiredState := commons.GetDesiredLndServiceState(serviceType, nodeId)
	if currentState.Status == desiredState.Status {
		return
	}
	switch desiredState.Status {
	case commons.ServiceActive:
		if currentState.Status == commons.ServiceInactive {
			processServiceBoot(db, serviceType, nodeId)
		}
	case commons.ServiceInactive:
		log.Info().Msgf("Inactivating %v for nodeId: %v.", serviceType.String(), nodeId)
		commons.SetInactiveLndServiceState(serviceType, nodeId)
	}
}

func processTorqService(db *sqlx.DB, serviceType commons.ServiceType) bool {
	currentState := commons.GetCurrentTorqServiceState(serviceType)
	desiredState := commons.GetDesiredTorqServiceState(serviceType)
	if currentState.Status == desiredState.Status {
		return true
	}
	switch desiredState.Status {
	case commons.ServiceActive:
		if currentState.Status == commons.ServiceInactive {
			processServiceBoot(db, serviceType, 0)
		}
	case commons.ServiceInactive:
		commons.SetInactiveTorqServiceState(serviceType)
	}
	return false
}

func processServiceBoot(db *sqlx.DB, serviceType commons.ServiceType, nodeId int) {
	failedAttemptTime := commons.GetLndFailedAttemptTime(serviceType, nodeId)
	if failedAttemptTime != nil && time.Since(*failedAttemptTime).Seconds() < 60 {
		return
	}

	log.Info().Msgf("Activating %v for nodeId: %v.", serviceType.String(), nodeId)

	ctx, cancel := context.WithCancel(context.Background())

	if nodeId == 0 {
		commons.InitTorqServiceState(serviceType, cancel)
	} else {
		commons.InitLndServiceState(serviceType, nodeId, cancel)
	}

	var conn *grpc.ClientConn
	var err error
	switch serviceType {
	case commons.VectorService, commons.AmbossService, commons.RebalanceService,
		commons.LndServiceChannelEventStream,
		commons.LndServiceGraphEventStream,
		commons.LndServiceTransactionStream,
		commons.LndServiceHtlcEventStream,
		commons.LndServiceForwardStream,
		commons.LndServiceInvoiceStream,
		commons.LndServicePaymentStream,
		commons.LndServicePeerEventStream,
		commons.LndServiceInFlightPaymentStream,
		commons.LndServiceChannelBalanceCacheStream:
		nodeConnectionDetails := commons.GetLndNodeConnectionDetails(nodeId)
		conn, err = lnd_connect.Connect(
			nodeConnectionDetails.GRPCAddress,
			nodeConnectionDetails.TLSFileBytes,
			nodeConnectionDetails.MacaroonFileBytes)
		if err != nil {
			log.Error().Err(err).Msgf("%v Failed to connect to lnd for node id: %v", serviceType.String(), nodeId)
			if nodeId == 0 {
				commons.SetFailedTorqServiceState(serviceType)
			} else {
				commons.SetFailedLndServiceState(serviceType, nodeId)
			}
			return
		}
	}

	log.Info().Msgf("%v Service booted for node id: %v", serviceType.String(), nodeId)
	switch serviceType {
	// NOT NODE ID SPECIFIC
	case commons.AutomationService:
		go services.Start(ctx, db)
	case commons.MaintenanceService:
		go services.StartMaintenanceService(ctx, db)
	case commons.CronService:
		go services.StartCronService(ctx, db)
	// NODE SPECIFIC
	case commons.VectorService:
		go vector_ping.Start(ctx, conn, nodeId)
	case commons.AmbossService:
		go amboss_ping.Start(ctx, conn, nodeId)
	case commons.RebalanceService:
		go services.StartRebalanceService(ctx, conn, db, nodeId)
	case commons.LndServiceChannelEventStream:
		go subscribe.StartChannelEventStream(ctx, conn, db, nodeId)
	case commons.LndServiceGraphEventStream:
		go subscribe.StartGraphEventStream(ctx, conn, db, nodeId)
	case commons.LndServiceTransactionStream:
		go subscribe.StartTransactionStream(ctx, conn, db, nodeId)
	case commons.LndServiceHtlcEventStream:
		go subscribe.StartHtlcEvents(ctx, conn, db, nodeId)
	case commons.LndServiceForwardStream:
		go subscribe.StartForwardStream(ctx, conn, db, nodeId)
	case commons.LndServiceInvoiceStream:
		go subscribe.StartInvoiceStream(ctx, conn, db, nodeId)
	case commons.LndServicePaymentStream:
		go subscribe.StartPaymentStream(ctx, conn, db, nodeId)
	case commons.LndServicePeerEventStream:
		go subscribe.StartPeerEvents(ctx, conn, db, nodeId)
	case commons.LndServiceInFlightPaymentStream:
		go subscribe.StartInFlightPaymentStream(ctx, conn, db, nodeId)
	case commons.LndServiceChannelBalanceCacheStream:
		go subscribe.StartChannelBalanceCacheMaintenance(ctx, conn, db, nodeId)
	}
}
