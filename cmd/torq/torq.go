package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/cmd/torq/internal/amboss_ping"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/cmd/torq/internal/vector_ping"
	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

var eventChannelGlobal = make(chan interface{})                     //nolint:gochecknoglobals
var serviceChannelGlobal = make(chan commons.ServiceChannelMessage) //nolint:gochecknoglobals

func main() {

	app := cli.NewApp()
	app.Name = "torq"
	app.EnableBashCompletion = true
	app.Version = build.Version()

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
			Name:  "torq.debug",
			Value: false,
			Usage: "Enable debug logging",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.cookie-path",
			Usage: "Path to auth cookie file",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.password",
			Usage: "Password used to access the API and frontend.",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.port",
			Value: "8080",
			Usage: "Port to serve the HTTP API",
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:  "torq.no-sub",
			Value: false,
			Usage: "Start the server without subscribing to node data.",
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
			Usage: "port of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.host",
			Value: "localhost",
			Usage: "host of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.user",
			Value: "torq",
			Usage: "Name of the postgres user with access to the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.password",
			Value: "password",
			Usage: "Name of the postgres user with access to the database",
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
			if c.Bool("torq.debug") {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
				log.Debug().Msg("Debug logging enabled")
			}

			// Print startup message
			fmt.Printf("Starting Torq %s\n", build.Version())

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
			commons.RunningServices[commons.LndService] = &commons.Services{ServiceType: commons.LndService}
			commons.RunningServices[commons.VectorService] = &commons.Services{ServiceType: commons.VectorService}
			commons.RunningServices[commons.AmbossService] = &commons.Services{ServiceType: commons.AmbossService}
			commons.RunningServices[commons.TorqService] = &commons.Services{ServiceType: commons.TorqService}

			ctxGlobal, cancelGlobal := context.WithCancel(context.Background())
			defer cancelGlobal()

			broadcasterGlobal := broadcast.NewBroadcastServer(ctxGlobal, eventChannelGlobal)

			go commons.ManagedChannelGroupCache(commons.ManagedChannelGroupChannel, ctxGlobal)
			go commons.ManagedChannelStateCache(commons.ManagedChannelStateChannel, broadcasterGlobal, ctxGlobal)
			go commons.ManagedSettingsCache(commons.ManagedSettingsChannel, ctxGlobal)
			go commons.ManagedNodeCache(commons.ManagedNodeChannel, ctxGlobal)
			go commons.ManagedChannelCache(commons.ManagedChannelChannel, ctxGlobal)

			// This listens to events:
			// When Torq has status initializing it loads the caches and starts the LndServices
			// When Torq has status inactive a panic is created (i.e. migration failed)
			// When LndService has status active other services like Amboss and Vector are booted (they depend on LND)
			go func(db *sqlx.DB, serviceChannel chan commons.ServiceChannelMessage, broadcaster broadcast.BroadcastServer) {
				for {
					listener := broadcaster.Subscribe()
					for event := range listener {
						if serviceEvent, ok := event.(commons.ServiceEvent); ok {
							if serviceEvent.Type == commons.TorqService {
								switch serviceEvent.Status {
								case commons.Inactive:
									log.Info().Msg("Torq is dead.")
									panic("TorqService cannot be bootstrapped")
								case commons.Pending:
									log.Info().Msg("Torq is booting.")
								case commons.Initializing:
									log.Info().Msg("Torq is initialising.")
									err = settings.InitializeManagedSettingsCache(db)
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

									log.Info().Msg("Loading caches in memory.")
									err := corridors.RefreshCorridorCache(db)
									if err != nil {
										log.Error().Err(err).Msg("Torq cannot be initialized (Loading caches in memory).")
									}
									serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.LndService}
								}
							}
							if serviceEvent.Type == commons.LndService {
								if serviceEvent.Status == commons.Active && serviceEvent.SubscriptionStream == nil {
									log.Debug().Msgf("LndService booted checking for Vector activation for nodeId: %v", serviceEvent.NodeId)
									if commons.RunningServices[commons.VectorService].GetStatus(serviceEvent.NodeId) == commons.Inactive {
										serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.VectorService, NodeId: serviceEvent.NodeId}
									}
									log.Debug().Msgf("LndService booted checking for Amboss activation for nodeId: %v", serviceEvent.NodeId)
									if commons.RunningServices[commons.AmbossService].GetStatus(serviceEvent.NodeId) == commons.Inactive {
										serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.AmbossService, NodeId: serviceEvent.NodeId}
									}
								}
							}
						}
					}
				}
			}(db, serviceChannelGlobal, broadcasterGlobal)

			commons.RunningServices[commons.TorqService].AddSubscription(commons.TorqDummyNodeId, cancelGlobal, eventChannelGlobal)

			// This function initiates the database migration(s) and parses command line parameters
			// When done the TorqService is set to Initialising
			go func(db *sqlx.DB, c *cli.Context, eventChannel chan interface{}) {
				fmt.Println("Checking for migrations..")
				// Check if the database needs to be migrated.
				err = database.MigrateUp(db)
				if err != nil && !errors.Is(err, migrate.ErrNoChange) {
					log.Error().Err(err).Msg("Torq could not migrate the database.")
					commons.RunningServices[commons.TorqService].RemoveSubscription(commons.TorqDummyNodeId, eventChannel)
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
							_, err = settings.SetNodeConnectionDetails(db, nodeConnectionDetails)
							if err != nil {
								log.Error().Err(err).Msg("Failed to update the node name (cosmetics problem).")
							}
						} else {
							log.Info().Msg("Node specified in config is present, updating Macaroon and TLS files")
							if err = settings.SetNodeConnectionDetailsByConnectionDetails(db, nodeId, commons.Active, grpcAddress, tlsFile, macaroonFile); err != nil {
								log.Error().Err(err).Msg("Problem updating node files")
								commons.RunningServices[commons.TorqService].RemoveSubscription(commons.TorqDummyNodeId, eventChannel)
							}
						}
					}
					break
				}

				commons.RunningServices[commons.TorqService].Initialising(commons.TorqDummyNodeId, eventChannel)
			}(db, c, eventChannelGlobal)

			if !c.Bool("torq.no-sub") {
				// go routine that responds to commands to boot and kill services
				go (func(db *sqlx.DB, serviceChannel chan commons.ServiceChannelMessage, eventChannel chan interface{}, broadcaster broadcast.BroadcastServer) {
					for {
						serviceCmd := <-serviceChannel
						services := commons.RunningServices[serviceCmd.ServiceType]
						var nodes []settings.ConnectionDetails
						var enforcedServiceStatus *commons.Status
						if serviceCmd.ServiceType == commons.LndService {
							if serviceCmd.ServiceCommand == commons.Boot {
								log.Info().Msgf("Verifying LND service requirement.")
								if serviceCmd.NodeId != 0 {
									enforcedServiceStatus = services.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
								}
								if serviceCmd.EnforcedServiceStatus != nil {
									enforcedServiceStatus = serviceCmd.EnforcedServiceStatus
								}
								if serviceCmd.NodeId == 0 {
									commons.RunningServices[commons.TorqService].Booted(commons.TorqDummyNodeId, nil, eventChannel)
									nodes, err = settings.GetActiveNodesConnectionDetails(db)
									if err != nil {
										log.Error().Err(err).Msg("Getting connection details")
									}
								} else {
									if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Inactive {
										nodes = []settings.ConnectionDetails{}
									} else {
										node, err := settings.GetConnectionDetailsById(db, serviceCmd.NodeId)
										if err != nil {
											log.Error().Err(errors.Wrap(err, "Getting connection details")).Send()
											return
										}
										if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Active {
											nodes = []settings.ConnectionDetails{node}
										} else {
											if node.Status != commons.Active {
												nodes = []settings.ConnectionDetails{}
											} else {
												nodes = []settings.ConnectionDetails{node}
											}
										}
									}
								}
								for _, node := range nodes {
									if serviceCmd.NodeId == 0 || serviceCmd.NodeId == node.NodeId {
										bootLock := services.GetBootLock(node.NodeId)
										successful := bootLock.TryLock()
										if successful {
											go (func(node settings.ConnectionDetails, bootLock *sync.Mutex,
												services *commons.Services,
												serviceChannel chan commons.ServiceChannelMessage,
												eventChannel chan interface{}) {
												defer func() {
													if commons.MutexLocked(bootLock) {
														bootLock.Unlock()
													}
												}()

												ctx := context.Background()
												ctx, cancel := context.WithCancel(ctx)

												log.Info().Msgf("Subscribing to LND for node id: %v", node.NodeId)
												services.AddSubscription(node.NodeId, cancel, eventChannel)
												conn, err := lnd_connect.Connect(
													node.GRPCAddress,
													node.TLSFileBytes,
													node.MacaroonFileBytes,
												)
												if err != nil {
													log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
													services.RemoveSubscription(node.NodeId, eventChannel)
													log.Info().Msgf("LND Subscription will be restarted (when active) in 10 seconds for node id: %v", node.NodeId)
													time.Sleep(10 * time.Second)
													serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: serviceCmd.ServiceType, NodeId: node.NodeId}
													return
												}

												services.Booted(node.NodeId, bootLock, eventChannel)
												commons.RunningServices[commons.LndService].SetIncludeIncomplete(node.NodeId, node.HasNodeConnectionDetailCustomSettings(commons.ImportFailedPayments))
												log.Info().Msgf("LND Subscription booted for node id: %v", node.NodeId)
												err = subscribe.Start(ctx, conn, db, node.NodeId, broadcaster, eventChannel, serviceChannel)
												if err != nil {
													log.Error().Err(err).Send()
													// only log the error, don't return
												}
												log.Info().Msgf("LND Subscription stopped for node id: %v", node.NodeId)
												services.RemoveSubscription(node.NodeId, eventChannel)
												if services.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
													log.Info().Msgf("LND Subscription will be restarted (when active) for node id: %v", node.NodeId)
												} else {
													log.Info().Msgf("LND Subscription will be restarted (when active) in 10 seconds for node id: %v", node.NodeId)
													time.Sleep(10 * time.Second)
												}
												serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: serviceCmd.ServiceType, NodeId: node.NodeId}
											})(node, bootLock, services, serviceChannel, eventChannel)
										} else {
											log.Error().Msgf("Requested Vector Ping Service start failed. A start is already running.")
										}
									}
								}
							}
							if serviceCmd.ServiceCommand == commons.Kill {
								serviceCmd.Out <- services.Cancel(serviceCmd.NodeId, serviceCmd.EnforcedServiceStatus, serviceCmd.NoDelay, eventChannel)
							}
						}
						if serviceCmd.ServiceType == commons.VectorService {
							if serviceCmd.ServiceCommand == commons.Boot {
								log.Info().Msgf("Verifying Vector ping service requirement.")
								if serviceCmd.NodeId != 0 {
									enforcedServiceStatus = services.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
								}
								if serviceCmd.EnforcedServiceStatus != nil {
									enforcedServiceStatus = serviceCmd.EnforcedServiceStatus
								}
								if serviceCmd.NodeId == 0 {
									nodes, err = settings.GetVectorPingNodesConnectionDetails(db)
									if err != nil {
										log.Error().Err(err).Msg("Getting connection details")
									}
								} else {
									if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Inactive {
										nodes = []settings.ConnectionDetails{}
									} else {
										node, err := settings.GetConnectionDetailsById(db, serviceCmd.NodeId)
										if err == nil {
											if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Active {
												nodes = []settings.ConnectionDetails{node}
											} else {
												if node.Status != commons.Active || !node.HasPingSystem(commons.Vector) {
													nodes = []settings.ConnectionDetails{}
												} else {
													nodes = []settings.ConnectionDetails{node}
												}
											}
										} else {
											log.Error().Err(errors.Wrap(err, "Getting connection details")).Send()
											return
										}
									}
								}

								for _, node := range nodes {
									bootLock := services.GetBootLock(node.NodeId)
									successful := bootLock.TryLock()
									if successful {
										go (func(node settings.ConnectionDetails, bootLock *sync.Mutex,
											services *commons.Services,
											serviceChannel chan commons.ServiceChannelMessage,
											eventChannel chan interface{}) {
											defer func() {
												if commons.MutexLocked(bootLock) {
													bootLock.Unlock()
												}
											}()

											ctx := context.Background()
											ctx, cancel := context.WithCancel(ctx)

											log.Info().Msgf("Generating Vector ping service for node id: %v", node.NodeId)
											services.AddSubscription(node.NodeId, cancel, eventChannel)
											conn, err := lnd_connect.Connect(
												node.GRPCAddress,
												node.TLSFileBytes,
												node.MacaroonFileBytes)
											if err != nil {
												log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
												services.RemoveSubscription(node.NodeId, eventChannel)
												return
											}

											services.Booted(node.NodeId, bootLock, eventChannel)
											log.Info().Msgf("Vector Ping Service booted for node id: %v", node.NodeId)
											err = vector_ping.Start(ctx, conn)
											if err != nil {
												log.Error().Err(err).Msgf("Vector ping ended for node id: %v", node.NodeId)
											}
											log.Info().Msgf("Vector Ping Service stopped for node id: %v", node.NodeId)
											services.RemoveSubscription(node.NodeId, eventChannel)
											if services.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
												log.Info().Msgf("Vector Ping Service will be restarted (when active) for node id: %v", node.NodeId)
											} else {
												log.Info().Msgf("Vector Ping Service will be restarted (when active) in %v seconds for node id: %v", commons.SERVICES_ERROR_SLEEP_SECONDS, node.NodeId)
												time.Sleep(commons.SERVICES_ERROR_SLEEP_SECONDS * time.Second)
											}
											serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: serviceCmd.ServiceType, NodeId: node.NodeId}
										})(node, bootLock, services, serviceChannel, eventChannel)
									} else {
										log.Error().Msgf("Requested Vector Ping Service start failed. A start is already running.")
									}
								}
							}
							if serviceCmd.ServiceCommand == commons.Kill {
								serviceCmd.Out <- services.Cancel(serviceCmd.NodeId, serviceCmd.EnforcedServiceStatus, serviceCmd.NoDelay, eventChannel)
							}
						}
						if serviceCmd.ServiceType == commons.AmbossService {
							if serviceCmd.ServiceCommand == commons.Boot {
								log.Info().Msgf("Verifying Amboss ping service requirement.")
								if serviceCmd.NodeId != 0 {
									enforcedServiceStatus = services.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
								}
								if serviceCmd.EnforcedServiceStatus != nil {
									enforcedServiceStatus = serviceCmd.EnforcedServiceStatus
								}
								if serviceCmd.NodeId == 0 {
									nodes, err = settings.GetAmbossPingNodesConnectionDetails(db)
									if err != nil {
										log.Error().Err(err).Msg("Getting connection details")
									}
								} else {
									if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Inactive {
										nodes = []settings.ConnectionDetails{}
									} else {
										node, err := settings.GetConnectionDetailsById(db, serviceCmd.NodeId)
										if err == nil {
											if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.Active {
												nodes = []settings.ConnectionDetails{node}
											} else {
												if node.Status != commons.Active || !node.HasPingSystem(commons.Amboss) {
													nodes = []settings.ConnectionDetails{}
												} else {
													nodes = []settings.ConnectionDetails{node}
												}
											}
										} else {
											log.Error().Err(errors.Wrap(err, "Getting connection details")).Send()
											return
										}
									}
								}

								for _, node := range nodes {
									bootLock := services.GetBootLock(node.NodeId)
									successful := bootLock.TryLock()
									if successful {
										go (func(node settings.ConnectionDetails, bootLock *sync.Mutex,
											services *commons.Services,
											serviceChannel chan commons.ServiceChannelMessage,
											eventChannel chan interface{}) {
											defer func() {
												if commons.MutexLocked(bootLock) {
													bootLock.Unlock()
												}
											}()
											ctx := context.Background()
											ctx, cancel := context.WithCancel(ctx)

											log.Info().Msgf("Generating Amboss ping service for node id: %v", node.NodeId)
											services.AddSubscription(node.NodeId, cancel, eventChannel)
											conn, err := lnd_connect.Connect(
												node.GRPCAddress,
												node.TLSFileBytes,
												node.MacaroonFileBytes)
											if err != nil {
												log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
												services.RemoveSubscription(node.NodeId, eventChannel)
												return
											}

											services.Booted(node.NodeId, bootLock, eventChannel)
											log.Info().Msgf("Amboss Ping Service booted for node id: %v", node.NodeId)
											err = amboss_ping.Start(ctx, conn)
											if err != nil {
												log.Error().Err(err).Msgf("Amboss ping ended for node id: %v", node.NodeId)
											}
											log.Info().Msgf("Amboss Ping Service stopped for node id: %v", node.NodeId)
											services.RemoveSubscription(node.NodeId, eventChannel)
											if services.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
												log.Info().Msgf("Amboss Ping Service will be restarted (when active) for node id: %v", node.NodeId)
											} else {
												log.Info().Msgf("Amboss Ping Service will be restarted (when active) in %v seconds for node id: %v", commons.SERVICES_ERROR_SLEEP_SECONDS, node.NodeId)
												time.Sleep(commons.SERVICES_ERROR_SLEEP_SECONDS * time.Second)
											}
											serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: serviceCmd.ServiceType, NodeId: node.NodeId}
										})(node, bootLock, services, serviceChannel, eventChannel)
									} else {
										log.Error().Msgf("Requested Amboss Ping Service start failed. A start is already running.")
									}
								}
							}
							if serviceCmd.ServiceCommand == commons.Kill {
								serviceCmd.Out <- services.Cancel(serviceCmd.NodeId, serviceCmd.EnforcedServiceStatus, serviceCmd.NoDelay, eventChannel)
							}
						}
					}
				})(db, serviceChannelGlobal, eventChannelGlobal, broadcasterGlobal)
			} else {
				go (func(serviceChannel chan commons.ServiceChannelMessage) {
					for {
						serviceCmd := <-serviceChannel
						log.Warn().Msgf("Ignoring Service call for node id: %v", serviceCmd.NodeId)
					}
				})(serviceChannelGlobal)
			}

			if err = torqsrv.Start(c.Int("torq.port"), c.String("torq.password"), c.String("torq.cookie-path"),
				db, eventChannelGlobal, broadcasterGlobal, serviceChannelGlobal); err != nil {
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
