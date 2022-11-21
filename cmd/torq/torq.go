package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
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
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

var eventChannel = make(chan interface{}) //nolint:gochecknoglobals

var serviceChannel = make(chan commons.ServiceChannelMessage) //nolint:gochecknoglobals

type services struct {
	mu          sync.RWMutex
	runningList map[int]func()
	// bootLock entry guards against running restart code whilst it's already running
	bootLock map[int]*sync.Mutex
	bootTime map[int]time.Time
	// enforcedServiceStatus entry is a one time status enforcement for a service
	enforcedServiceStatus map[int]*commons.Status
	// noDelay entry is a one time no delay enforcement for a service
	noDelay map[int]bool
}

func (rs *services) AddSubscription(nodeId int, cancelFunc func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs)
	rs.runningList[nodeId] = cancelFunc
}

func (rs *services) RemoveSubscription(nodeId int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs)
	_, exists := rs.runningList[nodeId]
	if exists {
		delete(rs.runningList, nodeId)
	}
}

func (rs *services) Cancel(nodeId int, enforcedServiceStatus *commons.Status, noDelay bool) commons.Status {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs)
	rs.noDelay[nodeId] = noDelay
	_, exists := rs.runningList[nodeId]
	if exists {
		rs.enforcedServiceStatus[nodeId] = enforcedServiceStatus
		_, exists = rs.bootLock[nodeId]
		if exists && commons.MutexLocked(rs.bootLock[nodeId]) {
			return commons.Pending
		} else {
			rs.runningList[nodeId]()
			delete(rs.runningList, nodeId)
			return commons.Active
		}
	}
	return commons.Inactive
}

func (rs *services) GetEnforcedServiceStatusCheck(nodeId int) *commons.Status {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs)
	_, exists := rs.enforcedServiceStatus[nodeId]
	if exists {
		enforcedServiceStatus := rs.enforcedServiceStatus[nodeId]
		delete(rs.enforcedServiceStatus, nodeId)
		return enforcedServiceStatus
	}
	return nil
}

func (rs *services) IsNoDelay(nodeId int) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs)
	_, exists := rs.noDelay[nodeId]
	if exists {
		noDelay := rs.noDelay[nodeId]
		delete(rs.noDelay, nodeId)
		return noDelay
	}
	return false
}

func (rs *services) GetBootLock(nodeId int) *sync.Mutex {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	initServiceMaps(rs)
	lock := rs.bootLock[nodeId]
	if lock == nil {
		lock = &sync.Mutex{}
		rs.bootLock[nodeId] = lock
	}
	return lock
}

func (rs *services) Booted(nodeId int, bootLock *sync.Mutex) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	bootLock.Unlock()
	initServiceMaps(rs)
	rs.bootTime[nodeId] = time.Now().UTC()
}

func initServiceMaps(rs *services) {
	if rs.runningList == nil {
		rs.runningList = make(map[int]func())
	}
	if rs.bootLock == nil {
		rs.bootLock = make(map[int]*sync.Mutex)
	}
	if rs.bootTime == nil {
		rs.bootTime = make(map[int]time.Time)
	}
	if rs.enforcedServiceStatus == nil {
		rs.enforcedServiceStatus = make(map[int]*commons.Status)
	}
	if rs.noDelay == nil {
		rs.noDelay = make(map[int]bool)
	}
}

// IF YOU MAKE THESE SERVICES EXPORTED YOU CAN GET A LIST OF ACTIVE SUBSCRIPTIONS AND WHEN THEY WERE BOOTED
// NO CANCEL FUNCTION IN runningList MEANS NOT ACTIVE
// NICE TO ADD A CANCELLATION TIME???
var runningLndSubscriptions services //nolint:gochecknoglobals
var runningAmbossPings services      //nolint:gochecknoglobals
var runningVectorPings services      //nolint:gochecknoglobals

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

			fmt.Println("Checking for migrations..")
			// Check if the database needs to be migrated.
			err = database.MigrateUp(db)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			go commons.ManagedSettingsCache(commons.ManagedSettingsChannel, nil)
			err = settings.InitializeManagedSettingsCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain settings for ManagedSettings cache.")
			}

			go commons.ManagedNodeCache(commons.ManagedNodeChannel, nil)
			err = settings.InitializeManagedNodeCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain torq nodes for ManagedNode cache.")
			}

			go commons.ManagedChannelCache(commons.ManagedChannelChannel, nil)
			err = channels.InitializeManagedChannelCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain channels for ManagedChannel cache.")
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			broadcaster := broadcast.NewBroadcastServer(ctx, eventChannel)

			// if node specified on cmd flags then check if we already know about it
			if c.String("lnd.url") != "" && c.String("lnd.macaroon-path") != "" && c.String("lnd.tls-path") != "" {
				macaroonFile, err := os.ReadFile(c.String("lnd.macaroon-path"))
				if err != nil {
					log.Error().Err(err).Msg("Reading macaroon file from disk path from config")
					return errors.Wrap(err, "Reading macaroon file from disk path from config")
				}
				tlsFile, err := os.ReadFile(c.String("lnd.tls-path"))
				if err != nil {
					log.Error().Err(err).Msg("Reading tls file from disk path from config")
					return errors.Wrap(err, "Reading tls file from disk path from config")
				}
				grpcAddress := c.String("lnd.url")
				nodeId, err := settings.GetNodeIdByGRPC(db, grpcAddress)
				if err != nil {
					log.Error().Err(err).Msg("Checking if node specified in config exists")
					return errors.Wrap(err, "Checking if node specified in config exists")
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
						return errors.Wrap(err, "Updating node name")
					}
				} else {
					log.Info().Msg("Node specified in config is present, updating Macaroon and TLS files")
					if err = settings.SetNodeConnectionDetailsByConnectionDetails(db, nodeId, commons.Active, grpcAddress, tlsFile, macaroonFile); err != nil {
						log.Error().Err(err).Msg("Problem updating node files")
						return errors.Wrap(err, "Problem updating node files")
					}
				}
			}

			if !c.Bool("torq.no-sub") {
				// initialise package level var for keeping state of subsciptions
				runningLndSubscriptions = services{}
				runningVectorPings = services{}
				runningAmbossPings = services{}

				// go routine that responds to start command and starts all subscriptions
				go (func() {
					for {
						serviceCmd := <-serviceChannel
						var nodes []settings.ConnectionDetails
						var enforcedServiceStatus *commons.Status
						if serviceCmd.ServiceType == commons.LndSubscription {
							svcs := &runningLndSubscriptions
							if serviceCmd.ServiceCommand == commons.Boot {
								log.Info().Msgf("Verifying LND service requirement.")
								if serviceCmd.NodeId != 0 {
									enforcedServiceStatus = svcs.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
								}
								if serviceCmd.EnforcedServiceStatus != nil {
									enforcedServiceStatus = serviceCmd.EnforcedServiceStatus
								}
								if serviceCmd.NodeId == 0 {
									nodes, err = settings.GetActiveNodesConnectionDetails(db)
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
												if node.Status != commons.Active {
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
									if serviceCmd.NodeId == 0 || serviceCmd.NodeId == node.NodeId {
										bootLock := svcs.GetBootLock(node.NodeId)
										successful := bootLock.TryLock()
										if successful {
											go (func(node settings.ConnectionDetails, bootLock *sync.Mutex) {
												defer func() {
													if commons.MutexLocked(bootLock) {
														bootLock.Unlock()
													}
												}()

												ctx := context.Background()
												ctx, cancel := context.WithCancel(ctx)

												log.Info().Msgf("Subscribing to LND for node id: %v", node.NodeId)
												svcs.AddSubscription(node.NodeId, cancel)
												conn, err := lnd_connect.Connect(
													node.GRPCAddress,
													node.TLSFileBytes,
													node.MacaroonFileBytes,
												)
												if err != nil {
													log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
													svcs.RemoveSubscription(node.NodeId)
													log.Info().Msgf("LND Subscription will be restarted (when active) in 10 seconds for node id: %v", node.NodeId)
													time.Sleep(10 * time.Second)
													serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.LndSubscription, NodeId: node.NodeId}
													return
												}

												svcs.Booted(node.NodeId, bootLock)
												log.Info().Msgf("LND Subscription booted for node id: %v", node.NodeId)
												err = subscribe.Start(ctx, conn, db, node.NodeId, eventChannel)
												if err != nil {
													log.Error().Err(err).Send()
													// only log the error, don't return
												}
												log.Info().Msgf("LND Subscription stopped for node id: %v", node.NodeId)
												svcs.RemoveSubscription(node.NodeId)
												if svcs.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
													log.Info().Msgf("LND Subscription will be restarted (when active) for node id: %v", node.NodeId)
												} else {
													log.Info().Msgf("LND Subscription will be restarted (when active) in 10 seconds for node id: %v", node.NodeId)
													time.Sleep(10 * time.Second)
												}
												serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.LndSubscription, NodeId: node.NodeId}
											})(node, bootLock)
										} else {
											log.Error().Msgf("Requested Vector Ping Service start failed. A start is already running.")
										}
									}
								}
							}
							if serviceCmd.ServiceCommand == commons.Kill {
								serviceCmd.Out <- svcs.Cancel(serviceCmd.NodeId, serviceCmd.EnforcedServiceStatus, serviceCmd.NoDelay)
							}
						}
						if serviceCmd.ServiceType == commons.VectorSubscription {
							svcs := &runningVectorPings
							if serviceCmd.ServiceCommand == commons.Boot {
								log.Info().Msgf("Verifying Vector ping service requirement.")
								if serviceCmd.NodeId != 0 {
									enforcedServiceStatus = svcs.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
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
									bootLock := svcs.GetBootLock(node.NodeId)
									successful := bootLock.TryLock()
									if successful {
										go (func(node settings.ConnectionDetails, bootLock *sync.Mutex) {
											defer func() {
												if commons.MutexLocked(bootLock) {
													bootLock.Unlock()
												}
											}()

											ctx := context.Background()
											ctx, cancel := context.WithCancel(ctx)

											log.Info().Msgf("Generating Vector ping service for node id: %v", node.NodeId)
											svcs.AddSubscription(node.NodeId, cancel)
											conn, err := lnd_connect.Connect(
												node.GRPCAddress,
												node.TLSFileBytes,
												node.MacaroonFileBytes)
											if err != nil {
												log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
												svcs.RemoveSubscription(node.NodeId)
												return
											}

											svcs.Booted(node.NodeId, bootLock)
											log.Info().Msgf("Vector Ping Service booted for node id: %v", node.NodeId)
											err = vector_ping.Start(ctx, conn)
											if err != nil {
												log.Error().Err(err).Msgf("Vector ping ended for node id: %v", node.NodeId)
											}
											log.Info().Msgf("Vector Ping Service stopped for node id: %v", node.NodeId)
											svcs.RemoveSubscription(node.NodeId)
											if svcs.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
												log.Info().Msgf("Vector Ping Service will be restarted (when active) for node id: %v", node.NodeId)
											} else {
												log.Info().Msgf("Vector Ping Service will be restarted (when active) in 60 seconds for node id: %v", node.NodeId)
												time.Sleep(1 * time.Minute)
											}
											serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.VectorSubscription, NodeId: node.NodeId}
										})(node, bootLock)
									} else {
										log.Error().Msgf("Requested Vector Ping Service start failed. A start is already running.")
									}
								}
							}
							if serviceCmd.ServiceCommand == commons.Kill {
								serviceCmd.Out <- svcs.Cancel(serviceCmd.NodeId, serviceCmd.EnforcedServiceStatus, serviceCmd.NoDelay)
							}
						}
						if serviceCmd.ServiceType == commons.AmbossSubscription {
							svcs := &runningAmbossPings
							if serviceCmd.ServiceCommand == commons.Boot {
								log.Info().Msgf("Verifying Amboss ping service requirement.")
								if serviceCmd.NodeId != 0 {
									enforcedServiceStatus = svcs.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
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
									bootLock := svcs.GetBootLock(node.NodeId)
									successful := bootLock.TryLock()
									if successful {
										go (func(node settings.ConnectionDetails, bootLock *sync.Mutex) {
											defer func() {
												if commons.MutexLocked(bootLock) {
													bootLock.Unlock()
												}
											}()
											ctx := context.Background()
											ctx, cancel := context.WithCancel(ctx)

											log.Info().Msgf("Generating Amboss ping service for node id: %v", node.NodeId)
											svcs.AddSubscription(node.NodeId, cancel)
											conn, err := lnd_connect.Connect(
												node.GRPCAddress,
												node.TLSFileBytes,
												node.MacaroonFileBytes)
											if err != nil {
												log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
												svcs.RemoveSubscription(node.NodeId)
												return
											}

											svcs.Booted(node.NodeId, bootLock)
											log.Info().Msgf("Amboss Ping Service booted for node id: %v", node.NodeId)
											err = amboss_ping.Start(ctx, conn)
											if err != nil {
												log.Error().Err(err).Msgf("Amboss ping ended for node id: %v", node.NodeId)
											}
											log.Info().Msgf("Amboss Ping Service stopped for node id: %v", node.NodeId)
											svcs.RemoveSubscription(node.NodeId)
											if svcs.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
												log.Info().Msgf("Amboss Ping Service will be restarted (when active) for node id: %v", node.NodeId)
											} else {
												log.Info().Msgf("Amboss Ping Service will be restarted (when active) in 60 seconds for node id: %v", node.NodeId)
												time.Sleep(1 * time.Minute)
											}
											serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.AmbossSubscription, NodeId: node.NodeId}
										})(node, bootLock)
									} else {
										log.Error().Msgf("Requested Amboss Ping Service start failed. A start is already running.")
									}
								}
							}
							if serviceCmd.ServiceCommand == commons.Kill {
								serviceCmd.Out <- svcs.Cancel(serviceCmd.NodeId, serviceCmd.EnforcedServiceStatus, serviceCmd.NoDelay)
							}
						}
					}
				})()

				serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.LndSubscription}
				serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.VectorSubscription}
				serviceChannel <- commons.ServiceChannelMessage{ServiceCommand: commons.Boot, ServiceType: commons.AmbossSubscription}
			} else {
				go (func() {
					for {
						serviceCmd := <-serviceChannel
						log.Warn().Msgf("Ignoring Service call for node id: %v", serviceCmd.NodeId)
					}
				})()
			}

			if err = torqsrv.Start(c.Int("torq.port"), c.String("torq.password"), c.String("torq.cookie-path"),
				db, eventChannel, broadcaster, serviceChannel); err != nil {
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
				return err
			}

			defer func() {
				cerr := db.Close()
				if err == nil {
					err = cerr
				}
			}()

			err = database.MigrateUp(db)
			if err != nil {
				return err
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
		if _, err := os.Stat(context.String("config")); err == nil {
			return altsrc.NewTomlSourceFromFile(context.String("config"))
		}
		return altsrc.NewMapInputSource("", map[interface{}]interface{}{}), nil
	}
}
