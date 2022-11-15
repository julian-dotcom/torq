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

var serviceChannel = make(chan serviceChannelMessage) //nolint:gochecknoglobals

type serviceType int

const (
	lndSubscription = serviceType(iota)
	vector
	amboss
)

type serviceCommand int

const (
	boot = serviceCommand(iota)
	kill
)

type serviceChannelMessage = struct {
	pingService serviceType
	pingCommand serviceCommand
	nodeId      int
}

type services struct {
	mu          sync.RWMutex
	runningList map[int]func()
}

func (rs *services) AddSubscription(localNodeId int, cancelFunc func()) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.runningList == nil {
		rs.runningList = make(map[int]func())
	}

	rs.runningList[localNodeId] = cancelFunc
}

func (rs *services) RemoveSubscription(localNodeId int) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.runningList == nil {
		rs.runningList = make(map[int]func())
	}
	delete(rs.runningList, localNodeId)
}

func (rs *services) Contains(localNodeId int) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	_, exists := rs.runningList[localNodeId]
	return exists
}

func (rs *services) GetCancelFunction(localNodeId int) func() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	cancelFunction := rs.runningList[localNodeId]
	return cancelFunction
}

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
					log.Debug().Msg("Node specified in config is not in DB, adding it")
					nodeConnectionDetails, err := settings.AddNodeToDB(db, commons.LND, grpcAddress, tlsFile, macaroonFile)
					if err != nil {
						log.Error().Err(err).Msg("Adding node specified in config to database")
						return errors.Wrap(err, "Adding node specified in config to database")
					}
					nodeConnectionDetails.Name = "Auto configured node"
					_, err = settings.SetNodeConnectionDetails(db, nodeConnectionDetails)
					if err != nil {
						return errors.Wrap(err, "Updating node name")
					}
				} else {
					log.Debug().Msg("Node specified in config is present, updating Macaroon and TLS files")
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
						if serviceCmd.pingService == lndSubscription {
							if serviceCmd.pingCommand == boot {
								nodes, err := settings.GetActiveNodesConnectionDetails(db)
								if err != nil {
									log.Error().Err(errors.Wrap(err, "Getting connection details")).Send()
									return
								}
								for _, node := range nodes {
									if serviceCmd.nodeId == 0 || serviceCmd.nodeId == node.NodeId {
										go (func(node settings.ConnectionDetails) {

											ctx := context.Background()
											ctx, cancel := context.WithCancel(ctx)

											log.Info().Msgf("Subscribing to LND for node id: %v", node.NodeId)
											runningLndSubscriptions.AddSubscription(node.NodeId, cancel)
											conn, err := lnd_connect.Connect(
												node.GRPCAddress,
												node.TLSFileBytes,
												node.MacaroonFileBytes,
											)
											if err != nil {
												log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
												runningLndSubscriptions.RemoveSubscription(node.NodeId)
												return
											}

											err = subscribe.Start(ctx, conn, db, node.NodeId, eventChannel)
											if err != nil {
												log.Error().Err(err).Send()
												// only log the error, don't return
											}
											log.Info().Msgf("LND Subscription stopped for node id: %v", node.NodeId)
											runningLndSubscriptions.RemoveSubscription(node.NodeId)
											log.Info().Msgf("LND Subscription will be restarted (when active) in 60 seconds for node id: %v", node.NodeId)
											time.Sleep(1 * time.Minute)
											serviceChannel <- serviceChannelMessage{pingCommand: boot, pingService: lndSubscription, nodeId: node.NodeId}
										})(node)
									}
								}
							}
							if serviceCmd.pingCommand == kill {
								runningLndSubscriptions.GetCancelFunction(serviceCmd.nodeId)()
							}
						}
						if serviceCmd.pingService == vector {
							if serviceCmd.pingCommand == boot {
								log.Info().Msgf("Verifying Vector ping service requirement.")

								var nodes []settings.ConnectionDetails
								if serviceCmd.nodeId == 0 {
									nodes, err = settings.GetVectorPingNodesConnectionDetails(db)
									if err != nil {
										log.Error().Err(err).Msg("Getting connection details")
									}
								} else {
									node, err := settings.GetConnectionDetailsById(db, serviceCmd.nodeId)
									if err != nil {
										log.Error().Err(err).Msg("Getting connection details")
										return
									}
									if node.Status != commons.Active || !node.HasPingSystem(commons.Vector) {
										nodes = []settings.ConnectionDetails{}
									} else {
										nodes = []settings.ConnectionDetails{node}
									}
								}

								for _, node := range nodes {
									go (func(node settings.ConnectionDetails) {

										ctx := context.Background()
										ctx, cancel := context.WithCancel(ctx)

										log.Info().Msgf("Generating Vector ping service for node id: %v", node.NodeId)
										runningVectorPings.AddSubscription(node.NodeId, cancel)
										conn, err := lnd_connect.Connect(
											node.GRPCAddress,
											node.TLSFileBytes,
											node.MacaroonFileBytes)
										if err != nil {
											log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
											runningVectorPings.RemoveSubscription(node.NodeId)
											return
										}

										err = vector_ping.Start(ctx, conn)
										if err != nil {
											log.Error().Err(err).Msgf("Vector ping ended for node id: %v", node.NodeId)
										}
										log.Info().Msgf("Vector Ping Service stopped for node id: %v", node.NodeId)
										runningVectorPings.RemoveSubscription(node.NodeId)
										log.Info().Msgf("Vector Ping Service will be restarted (when active) in 60 seconds for node id: %v", node.NodeId)
										time.Sleep(1 * time.Minute)
										serviceChannel <- serviceChannelMessage{pingCommand: boot, pingService: vector, nodeId: node.NodeId}
									})(node)
								}
							}
							if serviceCmd.pingCommand == kill {
								runningVectorPings.GetCancelFunction(serviceCmd.nodeId)()
							}
						}
						if serviceCmd.pingService == amboss {
							if serviceCmd.pingCommand == boot {
								log.Info().Msgf("Verifying Amboss ping service requirement.")
								var nodes []settings.ConnectionDetails
								if serviceCmd.nodeId == 0 {
									nodes, err = settings.GetAmbossPingNodesConnectionDetails(db)
									if err != nil {
										log.Error().Err(err).Msg("Getting connection details")
									}
								} else {
									node, err := settings.GetConnectionDetailsById(db, serviceCmd.nodeId)
									if err != nil {
										log.Error().Err(err).Msg("Getting connection details")
										return
									}
									if node.Status != commons.Active || !node.HasPingSystem(commons.Amboss) {
										nodes = []settings.ConnectionDetails{}
									} else {
										nodes = []settings.ConnectionDetails{node}
									}
								}

								for _, node := range nodes {
									go (func(node settings.ConnectionDetails) {

										ctx := context.Background()
										ctx, cancel := context.WithCancel(ctx)

										log.Info().Msgf("Generating Amboss ping service for node id: %v", node.NodeId)
										runningAmbossPings.AddSubscription(node.NodeId, cancel)
										conn, err := lnd_connect.Connect(
											node.GRPCAddress,
											node.TLSFileBytes,
											node.MacaroonFileBytes)
										if err != nil {
											log.Error().Err(err).Msgf("Failed to connect to lnd for node id: %v", node.NodeId)
											runningAmbossPings.RemoveSubscription(node.NodeId)
											return
										}

										err = amboss_ping.Start(ctx, conn)
										if err != nil {
											log.Error().Err(err).Msgf("Amboss ping ended for node id: %v", node.NodeId)
										}
										log.Info().Msgf("Amboss Ping Service stopped for node id: %v", node.NodeId)
										runningAmbossPings.RemoveSubscription(node.NodeId)
										log.Info().Msgf("Amboss Ping Service will be restarted (when active) in 60 seconds for node id: %v", node.NodeId)
										time.Sleep(1 * time.Minute)
										serviceChannel <- serviceChannelMessage{pingCommand: boot, pingService: amboss, nodeId: node.NodeId}
									})(node)
								}
							}
							if serviceCmd.pingCommand == kill {
								runningAmbossPings.GetCancelFunction(serviceCmd.nodeId)()
							}
						}
					}
				})()

				serviceChannel <- serviceChannelMessage{pingCommand: boot, pingService: lndSubscription}
				serviceChannel <- serviceChannelMessage{pingCommand: boot, pingService: vector}
				serviceChannel <- serviceChannelMessage{pingCommand: boot, pingService: amboss}

			}

			if err = torqsrv.Start(c.Int("torq.port"), c.String("torq.password"), c.String("torq.cookie-path"),
				db, eventChannel, broadcaster, RestartLNDSubscription); err != nil {
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

// guards against running restart code whilst it's already running
var restartLock sync.RWMutex //nolint:gochecknoglobals

func RestartLNDSubscription() error {
	locked := restartLock.TryLock()
	if !locked {
		return errors.New("Already restarting")
	}
	defer restartLock.Unlock()

	log.Info().Msg("Stopping subscriptions")
	for nodeId := range runningLndSubscriptions.runningList {
		serviceChannel <- serviceChannelMessage{pingCommand: kill, pingService: lndSubscription, nodeId: nodeId}
	}

	for {
		if len(runningLndSubscriptions.runningList) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Info().Msg("All subscriptions stopped")
	log.Info().Msg("Restarting subscriptions")
	serviceChannel <- serviceChannelMessage{pingCommand: boot, pingService: lndSubscription}
	return nil
}

func loadFlags() func(context *cli.Context) (altsrc.InputSourceContext, error) {
	return func(context *cli.Context) (altsrc.InputSourceContext, error) {
		if _, err := os.Stat(context.String("config")); err == nil {
			return altsrc.NewTomlSourceFromFile(context.String("config"))
		}
		return altsrc.NewMapInputSource("", map[interface{}]interface{}{}), nil
	}
}
