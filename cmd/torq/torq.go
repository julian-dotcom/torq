package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/lnd"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "torq"
	app.EnableBashCompletion = true
	app.Version = build.Version()

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error finding home directory of user: %v", err)
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
			Name:  "torq.host",
			Value: "localhost",
			Usage: "Host address for your regular grpc",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.port",
			Value: "8080",
			Usage: "Port for your regular grpc",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.cert",
			Value: "./cert.pem",
			Usage: "Path to your cert.pem file used by the GRPC server (torq)",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.key",
			Value: "./key.pem",
			Usage: "Path to your key.pem file used by the GRPC server",
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

		// LND node connection details
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "lnd.node_address",
			Aliases: []string{"na"},
			Value:   "localhost:10009",
			Usage:   "Where to reach the lnd. Default: localhost:10009",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.tls",
			Usage: "Path to your tls.cert file (LND node).",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.macaroon",
			Usage: "Path to your admin.macaroon file. (LND node)",
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
				return fmt.Errorf("(cmd/lnc streamHtlcCommand) error connecting to db: %v", err)
			}

			defer func() {
				cerr := db.Close()
				if err == nil {
					err = cerr
				}
			}()

			fmt.Println("Checking for migrations..")
			// Check if the database needs to be migrated.
			err = database.MigrateUp(db.DB)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			if !c.Bool("torq.no-sub") {

				fmt.Println("Connecting to lightning node")
				// Connect to the node
				connectionDetails, err := settings.GetConnectionDetails(db)
				if err != nil {
					return fmt.Errorf("failed to get node connection details: %v", err)
				}

				conn, err := lnd.Connect(
					connectionDetails.GRPCAddress,
					connectionDetails.TLSFileBytes,
					connectionDetails.MacaroonFileBytes)
				if err != nil {
					return fmt.Errorf("failed to connect to lnd: %v", err)
				}

				ctx := context.Background()
				errs, ctx := errgroup.WithContext(ctx)

				// Subscribe to data from the node
				//   TODO: Attempt to restart subscriptions if they fail.
				errs.Go(func() error {
					err = subscribe.Start(ctx, conn, db)
					if err != nil {
						return err
					}
					return nil
				})
			}

			torqsrv.Start(c.Int("torq.port"), c.String("torq.password"), db)

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
				return fmt.Errorf("(cmd/lnc streamHtlcCommand) error connecting to db: %v", err)
			}

			defer func() {
				cerr := db.Close()
				if err == nil {
					err = cerr
				}
			}()

			fmt.Println("Checking for migrations..")
			// Check if the database needs to be migrated.
			err = database.MigrateUp(db.DB)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			fmt.Println("Connecting to lightning node")
			// Connect to the node
			connectionDetails, err := settings.GetConnectionDetails(db)
			if err != nil {
				return fmt.Errorf("failed to get node connection details: %v", err)
			}

			conn, err := lnd.Connect(
				connectionDetails.GRPCAddress,
				connectionDetails.TLSFileBytes,
				connectionDetails.MacaroonFileBytes)
			if err != nil {
				return fmt.Errorf("failed to connect to lnd: %v", err)
			}

			ctx := context.Background()
			errs, ctx := errgroup.WithContext(ctx)

			// Subscribe to data from the node
			//   TODO: Attempt to restart subscriptions if they fail.
			errs.Go(func() error {
				err = subscribe.Start(ctx, conn, db)
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

			err = database.MigrateUp(db.DB)
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
		log.Fatal(err)
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
