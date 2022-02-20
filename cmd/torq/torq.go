package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/migrations"
	"github.com/lncapital/torq/pkg/database"
	"github.com/lncapital/torq/pkg/lndutil"
	"github.com/lncapital/torq/torqrpc"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
)

func loadFlags() func(context *cli.Context) (altsrc.InputSourceContext, error) {

	return func(context *cli.Context) (altsrc.InputSourceContext, error) {
		return altsrc.NewTomlSourceFromFile(context.String("config"))
	}

}

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
			Name:  "torq.host",
			Value: "localhost",
			Usage: "Host address for your regular grpc",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.port",
			Value: "50050",
			Usage: "Port for your regular grpc",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.web_port",
			Value: "50051",
			Usage: "Port for your web grpc",
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
			err = migrations.MigrateUp(db.DB)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			fmt.Println("Connecting to lightning node")
			// Connect to the node
			conn, err := lndutil.Connect(
				c.String("lnd.node_address"),
				c.String("lnd.tls"),
				c.String("lnd.macaroon"))

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

			srv, err := torqsrv.NewServer(c.String("torq.host"), c.String("torq.port"),
				c.String("torq.web_port"), c.String("torq.cert"), c.String("torq.key"), db)

			// Starts the grpc server
			errs.Go(func() error {
				err := srv.StartGrpc()
				if err != nil {
					return err
				}
				return nil
			})

			// Starts the grpc-web proxy server
			errs.Go(func() error {
				err := srv.StartWeb()
				if err != nil {
					return err
				}
				return nil
			})

			return errs.Wait()
		},
	}

	startGrpc := &cli.Command{
		Name:  "start_grpc",
		Usage: "",
		Action: func(c *cli.Context) error {
			fmt.Println("Starting Torq gRPC server only")

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
			err = migrations.MigrateUp(db.DB)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			if err != nil {
				return fmt.Errorf("failed to connect to lnd: %v", err)
			}

			ctx := context.Background()
			errs, ctx := errgroup.WithContext(ctx)

			srv, err := torqsrv.NewServer(c.String("torq.host"), c.String("torq.port"),
				c.String("torq.web_port"), c.String("torq.cert"), c.String("torq.key"), db)

			// Starts the grpc server
			errs.Go(func() error {
				err := srv.StartGrpc()
				if err != nil {
					return err
				}
				return nil
			})

			// Starts the grpc-web proxy server
			errs.Go(func() error {
				err := srv.StartWeb()
				if err != nil {
					return err
				}
				return nil
			})

			return errs.Wait()
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
			err = migrations.MigrateUp(db.DB)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}

			fmt.Println("Connecting to lightning node")
			// Connect to the node
			conn, err := lndutil.Connect(
				c.String("lnd.node_address"),
				c.String("lnd.tls"),
				c.String("lnd.macaroon"))

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

	// TODO: Remove. Only used for manually testing grpc calls
	callGrpc := &cli.Command{
		Name:  "call",
		Usage: "",
		Action: func(c *cli.Context) error {

			var conn *grpc.ClientConn

			creds, err := credentials.NewClientTLSFromFile(c.String("torq.cert"), "")
			if err != nil {
				return fmt.Errorf("failed to load certificates: %v", err)
			}

			conn, err = grpc.Dial(fmt.Sprintf("%s:%s", c.String("torq.host"),
				c.String("torq.port")), grpc.WithTransportCredentials(creds))
			if err != nil {
				log.Fatalf("did not connect: %s", err)
			}
			defer conn.Close()

			client := torqrpc.NewTorqrpcClient(conn)
			ctx := context.Background()
			response, err := client.GetAggrigatedForwards(ctx, &torqrpc.AggregatedForwardsRequest{
				//FromTs: time.Date(2021, 02, 01, 0, 0, 0, 0, time.UTC).Unix(),
				//ToTs:   time.Date(2022, 02, 11, 0, 0, 0, 0, time.UTC).Unix(),
				FromTs: 0,
				ToTs:   0,
				Ids:    &torqrpc.AggregatedForwardsRequest_ChannelIds{},
				//Ids: &torqrpc.AggregatedForwardsRequest_PeerIds{
				//	PeerIds: &torqrpc.PeerIDs{
				//		PubKeys: []string{
				//			"03b323009613777edc4e8024c6726758396bc026da99545971164a6baad31dfc69",
				//			"039efdf7a40f848d9fe30f1ab97b297c06f96fde1992ece0eca7639711ecac5a93",
				//			"03e8b9a977fa3ae7acce74c25986c7240a921222e349729737df832a1b5ceb49df",
				//			"030bec0fb4a5dd5c2dffa5f008210a9b7e0689ee4bf7ae15a7e52491c65fbf4ca5",
				//			"021c97a90a411ff2b10dc2a8e32de2f29d2fa49d41bfbb52bd416e460db0747d0d",
				//		},
				//	},
				//},
			})
			if err != nil {
				return err
			}

			log.Printf("Response from server: %s", response)

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

			err = migrations.MigrateUp(db.DB)
			if err != nil {
				return err
			}

			return nil
		},
	}

	migrateDown := &cli.Command{
		Name:  "migrate_down",
		Usage: "Migrates the database down one step",
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

			err = migrations.MigrateDown(db.DB)
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
		startGrpc,
		subscribe,
		callGrpc,
		migrateUp,
		migrateDown,
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
