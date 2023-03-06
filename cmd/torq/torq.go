package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"os"
	"runtime"
	"strings"
	"sync"
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
	"github.com/lncapital/torq/cmd/torq/internal/services"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/cmd/torq/internal/vector_ping"
	"github.com/lncapital/torq/internal/automation"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

var serviceChannelGlobal = make(chan commons.ServiceChannelMessage)     //nolint:gochecknoglobals
var lightningRequestChannelGlobal = make(chan interface{})              //nolint:gochecknoglobals
var rebalanceRequestChannelGlobal = make(chan commons.RebalanceRequest) //nolint:gochecknoglobals

var serviceEventChannelGlobal = make(chan commons.ServiceEvent)               //nolint:gochecknoglobals
var htlcEventChannelGlobal = make(chan commons.HtlcEvent)                     //nolint:gochecknoglobals
var forwardEventChannelGlobal = make(chan commons.ForwardEvent)               //nolint:gochecknoglobals
var channelBalanceEventChannelGlobal = make(chan commons.ChannelBalanceEvent) //nolint:gochecknoglobals
var channelEventChannelGlobal = make(chan commons.ChannelEvent)               //nolint:gochecknoglobals
var nodeGraphEventChannelGlobal = make(chan commons.NodeGraphEvent)           //nolint:gochecknoglobals
var channelGraphEventChannelGlobal = make(chan commons.ChannelGraphEvent)     //nolint:gochecknoglobals
var invoiceEventChannelGlobal = make(chan commons.InvoiceEvent)               //nolint:gochecknoglobals
var paymentEventChannelGlobal = make(chan commons.PaymentEvent)               //nolint:gochecknoglobals
var transactionEventChannelGlobal = make(chan commons.TransactionEvent)       //nolint:gochecknoglobals
var peerEventChannelGlobal = make(chan commons.PeerEvent)                     //nolint:gochecknoglobals
var blockEventChannelGlobal = make(chan commons.BlockEvent)                   //nolint:gochecknoglobals
var webSocketResponseChannelGlobal = make(chan interface{})                   //nolint:gochecknoglobals

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
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:  "torq.pprof.active",
			Value: false,
			Usage: "Enable pprof",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.pprof.path",
			Value: "localhost:6060",
			Usage: "Set pprof path",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.vector.url",
			Value: commons.VECTOR_URL,
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

			// initialise package level var for keeping state of subsciptions
			commons.RunningServices = make(map[commons.ServiceType]*commons.Services, 0)
			for _, serviceType := range commons.GetServiceTypes() {
				commons.RunningServices[serviceType] = &commons.Services{ServiceType: serviceType}
			}

			ctxGlobal, cancelGlobal := context.WithCancel(context.Background())
			defer cancelGlobal()

			broadcasterGlobal := broadcast.NewBroadcastServer(ctxGlobal,
				serviceEventChannelGlobal, htlcEventChannelGlobal, forwardEventChannelGlobal,
				channelBalanceEventChannelGlobal, channelEventChannelGlobal, nodeGraphEventChannelGlobal, channelGraphEventChannelGlobal,
				invoiceEventChannelGlobal, paymentEventChannelGlobal, transactionEventChannelGlobal, peerEventChannelGlobal, blockEventChannelGlobal,
				webSocketResponseChannelGlobal, lightningRequestChannelGlobal, rebalanceRequestChannelGlobal)

			go commons.ManagedChannelGroupCache(commons.ManagedChannelGroupChannel, ctxGlobal)
			go commons.ManagedChannelStateCache(commons.ManagedChannelStateChannel, ctxGlobal, channelBalanceEventChannelGlobal)
			go commons.ManagedSettingsCache(commons.ManagedSettingsChannel, ctxGlobal)
			go commons.ManagedNodeCache(commons.ManagedNodeChannel, ctxGlobal)
			go commons.ManagedNodeAliasCache(commons.ManagedNodeAliasChannel, ctxGlobal)
			go commons.ManagedChannelCache(commons.ManagedChannelChannel, ctxGlobal)
			go commons.ManagedTaggedCache(commons.ManagedTaggedChannel, ctxGlobal)
			go commons.ManagedTriggerCache(commons.ManagedTriggerChannel, ctxGlobal)
			go tags.ManagedTagCache(tags.ManagedTagChannel, ctxGlobal)
			go automation.ManagedRebalanceCache(automation.ManagedRebalanceChannel, ctxGlobal)

			// This listens to events:
			// When Torq has status initializing it loads the caches and starts the LndServices
			// When Torq has status inactive a panic is created (i.e. migration failed)
			// When LndService has status active other services like Amboss and Vector are booted (they depend on LND)
			go processServiceEvents(db, serviceChannelGlobal, broadcasterGlobal)

			previousStatus := commons.RunningServices[commons.TorqService].AddSubscription(commons.TorqDummyNodeId, cancelGlobal)
			commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannelGlobal, previousStatus, commons.Pending, commons.TorqService, nil)

			// This function initiates the database migration(s) and parses command line parameters
			// When done the TorqService is set to Initialising
			go migrateAndProcessArguments(db, c, serviceEventChannelGlobal)

			// go routine that responds to commands to boot and kill services
			if !c.Bool("torq.no-sub") {
				go serviceChannelRoutine(db, c,
					serviceChannelGlobal,
					lightningRequestChannelGlobal,
					rebalanceRequestChannelGlobal,
					serviceEventChannelGlobal, broadcasterGlobal)
			} else {
				go serviceChannelDummyRoutine(serviceChannelGlobal)
			}

			if c.Bool("torq.pprof.active") {
				go pprofStartup(c)
			}

			if err = torqsrv.Start(c.Int("torq.port"), c.String("torq.password"), c.String("torq.cookie-path"),
				db, webSocketResponseChannelGlobal, broadcasterGlobal, lightningRequestChannelGlobal, rebalanceRequestChannelGlobal,
				serviceChannelGlobal, c.Bool("torq.auto-login")); err != nil {
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

func migrateAndProcessArguments(db *sqlx.DB, c *cli.Context, serviceEventChannel chan commons.ServiceEvent) {
	fmt.Println("Checking for migrations..")
	// Check if the database needs to be migrated.
	err := database.MigrateUp(db)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error().Err(err).Msg("Torq could not migrate the database.")
		previousStatus := commons.RunningServices[commons.TorqService].RemoveSubscription(commons.TorqDummyNodeId)
		commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannel, previousStatus, commons.Inactive, commons.TorqService, nil)
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
					previousStatus := commons.RunningServices[commons.TorqService].RemoveSubscription(commons.TorqDummyNodeId)
					commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannel, previousStatus, commons.Inactive, commons.TorqService, nil)
				}
			}
		}
		break
	}

	previousStatus := commons.RunningServices[commons.TorqService].Initialising(commons.TorqDummyNodeId)
	commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannel, previousStatus, commons.Initializing, commons.TorqService, nil)
}

func serviceChannelDummyRoutine(serviceChannel chan commons.ServiceChannelMessage) {
	for {
		serviceCmd := <-serviceChannel
		log.Warn().Msgf("Ignoring Service call for node id: %v", serviceCmd.NodeId)
	}
}

func serviceChannelRoutine(db *sqlx.DB, c *cli.Context, serviceChannel chan commons.ServiceChannelMessage,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest,
	serviceEventChannel chan commons.ServiceEvent, broadcaster broadcast.BroadcastServer) {

	for {
		serviceCmd := <-serviceChannel
		runningServices := commons.RunningServices[serviceCmd.ServiceType]
		var nodes []settings.ConnectionDetails
		var serviceNode settings.ConnectionDetails
		var err error
		var enforcedServiceStatus *commons.Status
		var name string
		switch serviceCmd.ServiceType {
		case commons.LndService:
			name = "LND"
		case commons.VectorService:
			name = "Vector Ping"
		case commons.AmbossService:
			name = "Amboss Ping"
		case commons.AutomationService:
			name = "Automation"
		case commons.LightningCommunicationService:
			name = "LightningCommunication"
		case commons.RebalanceService:
			name = "RebalanceService"
		case commons.MaintenanceService:
			name = "MaintenanceService"
		case commons.CronService:
			name = "CronService"
		}
		if name != "" {
			if serviceCmd.ServiceCommand == commons.Kill {
				previousStatus, cancelStatus := runningServices.Cancel(serviceCmd.NodeId, serviceCmd.EnforcedServiceStatus, serviceCmd.NoDelay)
				if cancelStatus == commons.Active {
					log.Info().Msgf("%v Service: Killed.", name)
					commons.SendServiceEvent(serviceCmd.NodeId, serviceEventChannel, previousStatus, commons.Inactive, serviceCmd.ServiceType, nil)
				}
				serviceCmd.Out <- cancelStatus
			}
			if serviceCmd.ServiceCommand == commons.Boot {
				if serviceCmd.NodeId != 0 {
					enforcedServiceStatus = runningServices.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
				}
				if serviceCmd.EnforcedServiceStatus != nil {
					enforcedServiceStatus = serviceCmd.EnforcedServiceStatus
				}
				if serviceCmd.NodeId != 0 {
					if serviceCmd.NodeId == commons.TorqDummyNodeId {
						nodes = []settings.ConnectionDetails{{
							NodeId: commons.TorqDummyNodeId,
							Status: commons.Active,
						}}
					} else {
						if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Inactive {
							nodes = []settings.ConnectionDetails{}
						} else {
							serviceNode, err = settings.GetConnectionDetailsById(db, serviceCmd.NodeId)
							if err != nil {
								log.Error().Err(errors.Wrapf(err, "%v Service Boot: Getting connection details", name)).Send()
								return
							}
							if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Active {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								if serviceNode.Status != commons.Active {
									nodes = []settings.ConnectionDetails{}
								}
							}
						}
					}
				}
				log.Info().Msgf("%v Service: Verifying requirement.", name)
				if nodes == nil {
					if serviceCmd.NodeId == 0 {
						if serviceCmd.ServiceType == commons.LndService {
							previousStatus := commons.RunningServices[commons.TorqService].Booted(commons.TorqDummyNodeId, nil)
							commons.SendServiceEvent(commons.TorqDummyNodeId, serviceEventChannel, previousStatus, commons.Active, commons.TorqService, nil)
						}
						switch serviceCmd.ServiceType {
						case commons.VectorService:
							nodes, err = settings.GetVectorPingNodesConnectionDetails(db)
						case commons.AmbossService:
							nodes, err = settings.GetAmbossPingNodesConnectionDetails(db)
						default:
							nodes, err = settings.GetActiveNodesConnectionDetails(db)
						}
						if err != nil {
							log.Error().Err(err).Msgf("%v Service: Getting connection details", name)
						}
					} else {
						switch serviceCmd.ServiceType {
						case commons.VectorService:
							if serviceNode.Status == commons.Active && serviceNode.HasPingSystem(commons.Vector) {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								nodes = []settings.ConnectionDetails{}
							}
						case commons.AmbossService:
							if serviceNode.Status == commons.Active && serviceNode.HasPingSystem(commons.Amboss) {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								nodes = []settings.ConnectionDetails{}
							}
						default:
							if serviceNode.Status == commons.Active {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								nodes = []settings.ConnectionDetails{}
							}
						}
					}
				}
				for _, node := range nodes {
					bootLock := runningServices.GetBootLock(node.NodeId)
					successful := bootLock.TryLock()
					if successful {
						switch serviceCmd.ServiceType {
						case commons.LndService:
							go processLndBoot(db, c, node, bootLock, runningServices, serviceCmd, serviceChannel,
								lightningRequestChannel, broadcaster, serviceEventChannel)
						default:
							go processServiceBoot(name, db, c, node, bootLock, runningServices, serviceCmd, serviceChannel,
								lightningRequestChannel, rebalanceRequestChannel, broadcaster, serviceEventChannel)
						}
					} else {
						log.Error().Msgf("%v Service: Requested start failed. A start is already running.", name)
					}
				}
			}
		}
	}
}

func processServiceEvents(db *sqlx.DB, serviceChannel chan commons.ServiceChannelMessage, broadcaster broadcast.BroadcastServer) {
	listener := broadcaster.SubscribeServiceEvent()
	for serviceEvent := range listener {
		if serviceEvent.Type == commons.TorqService {
			switch serviceEvent.Status {
			case commons.Inactive:
				log.Info().Msg("Torq is dead.")
				panic("TorqService cannot be bootstrapped")
			case commons.Pending:
				log.Info().Msg("Torq is booting.")
			case commons.Initializing:
				log.Info().Msg("Torq is initialising.")

				err := settings.InitializeManagedSettingsCache(db)
				if err != nil {
					log.Error().Err(err).Msg("Failed to obtain settings for ManagedSettings cache.")
				}

				err = settings.InitializeManagedNodeCache(db)
				if err != nil {
					log.Error().Err(err).Msg("Failed to obtain torq nodes for ManagedNode cache.")
				}

				err = channels.InitializeManagedChannelCache(db)
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
				serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.LndService}
			case commons.Active:
				serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.MaintenanceService, NodeId: commons.TorqDummyNodeId}
				serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.AutomationService, NodeId: commons.TorqDummyNodeId}
				serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.CronService, NodeId: commons.TorqDummyNodeId}
			}
		}
		if serviceEvent.Type == commons.LndService {
			if serviceEvent.Status == commons.Active && serviceEvent.SubscriptionStream == nil {
				log.Debug().Msgf("LndService booted for nodeId: %v", serviceEvent.NodeId)
				log.Debug().Msgf("Starting LightningCommunication Service for nodeId: %v", serviceEvent.NodeId)
				if commons.RunningServices[commons.LightningCommunicationService].GetStatus(serviceEvent.NodeId) == commons.Inactive {
					serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.LightningCommunicationService, NodeId: serviceEvent.NodeId}
				}
				log.Debug().Msgf("Starting Rebalance Service for nodeId: %v", serviceEvent.NodeId)
				if commons.RunningServices[commons.RebalanceService].GetStatus(serviceEvent.NodeId) == commons.Inactive {
					serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.RebalanceService, NodeId: serviceEvent.NodeId}
				}
				log.Debug().Msgf("Checking for Vector activation for nodeId: %v", serviceEvent.NodeId)
				if commons.RunningServices[commons.VectorService].GetStatus(serviceEvent.NodeId) == commons.Inactive {
					serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.VectorService, NodeId: serviceEvent.NodeId}
				}
				log.Debug().Msgf("Checking for Amboss activation for nodeId: %v", serviceEvent.NodeId)
				if commons.RunningServices[commons.AmbossService].GetStatus(serviceEvent.NodeId) == commons.Inactive {
					serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.AmbossService, NodeId: serviceEvent.NodeId}
				}
			}
		}
	}
}

func processLndBoot(db *sqlx.DB, c *cli.Context, node settings.ConnectionDetails, bootLock *sync.Mutex,
	runningServices *commons.Services, serviceCmd commons.ServiceChannelMessage, serviceChannel chan commons.ServiceChannelMessage,
	lightningRequestChannel chan interface{},
	broadcaster broadcast.BroadcastServer, serviceEventChannel chan commons.ServiceEvent) {

	defer func() {
		if commons.MutexLocked(bootLock) {
			bootLock.Unlock()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	log.Info().Msgf("Subscribing to LND for node id: %v", node.NodeId)
	previousStatus := runningServices.AddSubscription(node.NodeId, cancel)
	commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Pending, serviceCmd.ServiceType, nil)
	conn, err := lnd_connect.Connect(
		node.GRPCAddress,
		node.TLSFileBytes,
		node.MacaroonFileBytes,
	)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
		previousStatus = runningServices.RemoveSubscription(node.NodeId)
		commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Inactive, serviceCmd.ServiceType, nil)
		log.Info().Msgf("LND Subscription will be restarted (when active) in 10 seconds for node id: %v", node.NodeId)
		time.Sleep(10 * time.Second)
		serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: serviceCmd.ServiceType, NodeId: node.NodeId}
		return
	}

	previousStatus = runningServices.Booted(node.NodeId, bootLock)
	commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Active, serviceCmd.ServiceType, nil)
	commons.RunningServices[commons.LndService].SetNodeConnectionDetailCustomSettings(node.NodeId, node.CustomSettings)
	log.Info().Msgf("LND Subscription booted for node id: %v", node.NodeId)
	err = subscribe.Start(ctx, conn, db, c.String("torq.vector.url"), node.NodeId, broadcaster,
		serviceEventChannelGlobal, htlcEventChannelGlobal, forwardEventChannelGlobal,
		channelEventChannelGlobal, nodeGraphEventChannelGlobal, channelGraphEventChannelGlobal,
		invoiceEventChannelGlobal, paymentEventChannelGlobal, transactionEventChannelGlobal, peerEventChannelGlobal, blockEventChannelGlobal,
		lightningRequestChannel)
	if err != nil {
		log.Error().Err(err).Send()
		// only log the error, don't return
	}
	log.Info().Msgf("LND Subscription stopped for node id: %v", node.NodeId)
	previousStatus = runningServices.RemoveSubscription(node.NodeId)
	commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Inactive, serviceCmd.ServiceType, nil)
	if runningServices.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
		log.Info().Msgf("LND Subscription will be restarted (when active) for node id: %v", node.NodeId)
	} else {
		log.Info().Msgf("LND Subscription will be restarted (when active) in 10 seconds for node id: %v", node.NodeId)
		time.Sleep(10 * time.Second)
	}
	serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: serviceCmd.ServiceType, NodeId: node.NodeId}
}

func processServiceBoot(name string, db *sqlx.DB, c *cli.Context, node settings.ConnectionDetails, bootLock *sync.Mutex,
	runningServices *commons.Services, serviceCmd commons.ServiceChannelMessage, serviceChannel chan commons.ServiceChannelMessage,
	lightningRequestChannel chan interface{},
	rebalanceRequestChannel chan commons.RebalanceRequest,
	broadcaster broadcast.BroadcastServer, serviceEventChannel chan commons.ServiceEvent) {

	defer func() {
		if commons.MutexLocked(bootLock) {
			bootLock.Unlock()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	log.Info().Msgf("Generating %v Service for node id: %v", name, node.NodeId)
	previousStatus := runningServices.AddSubscription(node.NodeId, cancel)
	commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Pending, serviceCmd.ServiceType, nil)

	var conn *grpc.ClientConn
	var err error
	switch serviceCmd.ServiceType {
	case commons.VectorService:
		fallthrough
	case commons.AmbossService:
		fallthrough
	case commons.LightningCommunicationService:
		fallthrough
	case commons.RebalanceService:
		conn, err = lnd_connect.Connect(
			node.GRPCAddress,
			node.TLSFileBytes,
			node.MacaroonFileBytes)
		if err != nil {
			log.Error().Err(err).Msgf("%v Service Failed to connect to lnd for node id: %v", name, node.NodeId)
			previousStatus = runningServices.RemoveSubscription(node.NodeId)
			commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Inactive, serviceCmd.ServiceType, nil)
			return
		}
	}

	previousStatus = runningServices.Booted(node.NodeId, bootLock)
	commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Active, serviceCmd.ServiceType, nil)
	log.Info().Msgf("%v Service booted for node id: %v", name, node.NodeId)
	switch serviceCmd.ServiceType {
	case commons.VectorService:
		err = vector_ping.Start(ctx, conn, c.String("torq.vector.url"), node.NodeId)
	case commons.AmbossService:
		err = amboss_ping.Start(ctx, conn, node.NodeId)
	case commons.LightningCommunicationService:
		err = services.StartLightningCommunicationService(ctx, conn, db, node.NodeId, broadcaster)
	case commons.RebalanceService:
		err = services.StartRebalanceService(ctx, conn, db, node.NodeId, broadcaster)
	case commons.AutomationService:
		err = services.Start(ctx, db, lightningRequestChannel, rebalanceRequestChannel, broadcaster)
	case commons.MaintenanceService:
		err = services.StartMaintenanceService(ctx, db, c.String("torq.vector.url"), lightningRequestChannel)
	case commons.CronService:
		err = services.StartCronService(ctx, db)
	}
	if err != nil {
		log.Error().Err(err).Msgf("%v Service ended for node id: %v", name, node.NodeId)
	}
	log.Info().Msgf("%v Service stopped for node id: %v", name, node.NodeId)
	previousStatus = runningServices.RemoveSubscription(node.NodeId)
	commons.SendServiceEvent(node.NodeId, serviceEventChannel, previousStatus, commons.Inactive, serviceCmd.ServiceType, nil)
	if runningServices.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
		log.Info().Msgf("%v Service will be restarted (when active) for node id: %v", name, node.NodeId)
	} else {
		log.Info().Msgf("%v Service will be restarted (when active) in %v seconds for node id: %v", name, commons.SERVICES_ERROR_SLEEP_SECONDS, node.NodeId)
		time.Sleep(commons.SERVICES_ERROR_SLEEP_SECONDS * time.Second)
	}
	serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: serviceCmd.ServiceType, NodeId: node.NodeId}
}
