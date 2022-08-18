package main

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd_connect"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"golang.org/x/sync/errgroup"
	"os"
	"time"
)

var startchan = make(chan struct{})
var stopchan = make(chan struct{})
var stoppedchan = make(chan struct{})

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

		// Torq connection details
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
	}

	start := &cli.Command{
		Name:  "start",
		Usage: "Start the main daemon",
		Action: func(c *cli.Context) error {

			// Print startup message
			fmt.Printf("Starting Torq v%s\n", build.Version())

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

			if !c.Bool("torq.no-sub") {

				fmt.Println("Connecting to lightning node")
				// Connect to the node

				ctx := context.Background()
				ctx, cancel := context.WithCancel(ctx)

				// Subscribe to data from the node
				//   TODO: Attempt to restart subscriptions if they fail.
				go (func() error {
					for {
						select {
						case <-startchan:
							var nodes []settings.ConnectionDetails
							for {
								nodes, err = settings.GetConnectionDetails(db)
								if err != nil && err.Error() != "Missing node details" {
									return errors.Wrap(err, "Getting connection details")
								}
								if err != nil && err.Error() == "Missing node details" {
									log.Error().Msg("Missing node details. " +
										"Go to settings and enter IP, port, macasoon and tls certificate. " +
										"Retrying after a short delay")
									time.Sleep(5 * time.Second)
									log.Info().Msg("Retrying...")
									continue
								}
								break
							}

							for _, node := range nodes {
								go (func(node settings.ConnectionDetails) {
									conn, err := lnd_connect.Connect(
										node.GRPCAddress,
										node.TLSFileBytes,
										node.MacaroonFileBytes)
									if err != nil {
										log.Error().Msgf("Failed to connect to lnd: %v", err)
										stoppedchan <- struct{}{}
										return
									}

									log.Info().Msgf("Subscribing to LND for node id: %v", node.LocalNodeId)
									err = subscribe.Start(ctx, conn, db, node.LocalNodeId)
									if err != nil {
										log.Error().Err(err).Send()
									}
									log.Info().Msgf("LND Subscription stopped for node id: %v", node.LocalNodeId)
									stoppedchan <- struct{}{}
								})(node)
							}
						}
					}
				})()
				// starts LND subscription when Torq starts
				startchan <- struct{}{}

				go (func() {
					for {
						select {
						case <-stopchan:
							cancel()
							ctx, cancel = context.WithCancel(context.Background())
						}
					}
				})()

			}

			torqsrv.Start(c.Int("torq.port"), c.String("torq.password"), db, RestartLNDSubscription)

			return nil
		},
	}

	subscribe := &cli.Command{
		Name:  "subscribe",
		Usage: "Start the subscribe daemon, listening for data from LND",
		Action: func(c *cli.Context) error {

			// Print startup message
			fmt.Printf("Starting Torq v%s\n", build.Version())

			fmt.Println("Connecting to the Torq database")
			db, err := database.PgConnect(c.String("db.name"), c.String("db.user"),
				c.String("db.password"), c.String("db.host"), c.String("db.port"))
			if err != nil {
				return errors.Wrap(err, "subscribe cmd")
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

			fmt.Println("Connecting to lightning node")
			// Connect to the node
			connectionDetails, err := settings.GetConnectionDetails(db)
			if err != nil {
				return fmt.Errorf("failed to get node connection details: %v", err)
			}

			conn, err := lnd_connect.Connect(
				connectionDetails[0].GRPCAddress,
				connectionDetails[0].TLSFileBytes,
				connectionDetails[0].MacaroonFileBytes)
			if err != nil {
				return fmt.Errorf("failed to connect to lnd: %v", err)
			}

			ctx := context.Background()
			errs, ctx := errgroup.WithContext(ctx)

			// Subscribe to data from the node
			//   TODO: Attempt to restart subscriptions if they fail.
			errs.Go(func() error {
				err = subscribe.Start(ctx, conn, db, 1)
				if err != nil {
					return err
				}
				return nil
			})

			return errs.Wait()
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
		subscribe,
		migrateUp,
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

}

func RestartLNDSubscription() {
	fmt.Println("Stopping")
	stopchan <- struct{}{}
	<-stoppedchan
	fmt.Println("Stopped")
	fmt.Println("Starting again")
	startchan <- struct{}{}
}

func loadFlags() func(context *cli.Context) (altsrc.InputSourceContext, error) {
	return func(context *cli.Context) (altsrc.InputSourceContext, error) {
		if _, err := os.Stat(context.String("config")); err == nil {
			return altsrc.NewTomlSourceFromFile(context.String("config"))
		}
		return altsrc.NewMapInputSource("", map[interface{}]interface{}{}), nil
	}
}
