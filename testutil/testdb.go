package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/lncapital/torq/internal/channels"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/pkg/commons"
)

const superuserName = "postgres"
const testDbPort = 5433
const testDBPrefix = "torq_test_"
const TestPublicKey1 = "PublicKey1"
const TestPublicKey2 = "PublicKey2"
const TestFundingTransactionHash1 = "0101010101010101010101010101010101010101010101010101010101010101"
const TestChannelPoint1 = TestFundingTransactionHash1 + ":3"
const TestFundingTransactionHash2 = "0101010101010101010101010101010101010101010101010101010101010102"
const TestChannelPoint2 = TestFundingTransactionHash2 + ":3"
const TestFundingTransactionHash3 = "0101010101010101010101010101010101010101010101010101010101010103"
const TestChannelPoint3 = TestFundingTransactionHash3 + ":3"
const TestFundingTransactionHash4 = "0101010101010101010101010101010101010101010101010101010101010104"
const TestChannelPoint4 = TestFundingTransactionHash4 + ":3"
const TestFundingTransactionHash5_NOTINDB = "0101010101010101010101010101010101010101010101010101010101010105"
const TestChannelPoint5_NOTINDB = TestFundingTransactionHash5_NOTINDB + ":3"

func init() {
	// Set the seed for the random database name
	rand.Seed(time.Now().UnixNano())
}

// A Server represents a running PostgreSQL server.
type Server struct {
	baseURL string
	conn    *sql.DB
	dbNames []string
}

// InitTestDBConn creates a connection to the postgres user and creates the Server struct.
// This is used to create all other test databases and should be executed once at the top of a
// test file (in the Main function).
func InitTestDBConn() (*Server, error) {
	srv := &Server{
		baseURL: (&url.URL{
			Scheme: "postgres",
			Host:   fmt.Sprintf("localhost:%d", testDbPort),
			User:   url.UserPassword(superuserName, "password"),
			Path:   "/",
		}).String(),
	}

	var err error
	srv.conn, err = sql.Open("postgres", srv.baseURL+"?sslmode=disable")
	if err != nil {
		return nil, errors.Wrap(err, "SQL open connection")
	}

	//srv.conn.SetMaxOpenConns(1)

	return srv, nil
}

// Cleanup closes the connection to the connection to the postgres server used to create new test
// databases. This should only be used once for each test file.
func (srv *Server) Cleanup() error {

	killConnSql := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE
			-- don't kill my own connection!
			pid <> pg_backend_pid()
			-- don't kill the connections to other databases
			AND datname LIKE '` + testDBPrefix + `%';`

	// Kill all connections before deleting the test_databases
	_, err := srv.conn.Exec(killConnSql)
	if err != nil {
		return errors.Wrapf(err, "srv.conn.Cleanup(%s)", killConnSql)
	}

	// Drop (delete) all test databases
	for _, name := range srv.dbNames {
		_, err := srv.conn.Exec("DROP DATABASE " + name + ";")
		if err != nil {
			return errors.Wrapf(err, "srv.conn.Cleanup(\"DROP DATABASE %s;\"", name)
		}
	}

	if srv.conn != nil {
		err = srv.conn.Close()
		if err != nil {
			return errors.Wrap(err, "Closing server database connection")
		}
	}

	return nil
}

// dbUrl creates the db url based on the db name.
func (srv *Server) dbUrl(dbName string) string {
	return srv.baseURL + dbName + "?sslmode=disable"
}

// randomString is used to generate a unique database names.
func randString(n int) string {
	// rune used as source for random database names
	var letters = []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec
	}
	return string(b)
}

// createDatabase creates a new database on the server and returns its
// data source name.
func (srv *Server) createDatabase() (string, error) {

	// Create a new random name for the test database with prefix.
	dbName := testDBPrefix + randString(16)

	// Create a new test database
	_, err := srv.conn.Exec("CREATE DATABASE " + dbName + ";")
	if err != nil {
		return "", errors.Wrapf(err, "srv.conn.ExecContext(ctx, \"CREATE DATABASE %s;\"", dbName)
	}
	log.Debug().Msgf("Created database: %v", dbName)

	// Store all database names so that they can be easily dropped (deleted)
	srv.dbNames = append(srv.dbNames, dbName)
	return srv.dbUrl(dbName), nil
}

// NewTestDatabase opens a connection to a freshly created database on the server.
func (srv *Server) NewTestDatabase(migrate bool) (*sqlx.DB, context.CancelFunc, error) {

	// Create the new test database based on the main server connection
	dns, err := srv.createDatabase()
	if err != nil {
		return nil, nil, errors.Wrap(err, "srv.createDatabase(ctx)")
	}

	// Connect to the new test database
	db, err := sqlx.Open("postgres", dns)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "sqlx.Open(\"postgres\", %s)", dns)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	if migrate {
		// Migrate the new test database
		err = database.MigrateUp(db)
		if err != nil {
			cancel()
			return nil, nil, errors.Wrap(err, "Database Migrate Up")
		}

		var testNodeId1 int
		err = db.QueryRowx("INSERT INTO node (public_key, chain, network, created_on) VALUES ($1, $2, $3, $4) RETURNING node_id;",
			TestPublicKey1, commons.Bitcoin, commons.SigNet, time.Now().UTC()).Scan(&testNodeId1)
		if err != nil {
			cancel()
			return nil, nil, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", TestPublicKey1)
		}
		log.Debug().Msgf("Added test node with publicKey: %v nodeId: %v", TestPublicKey1, testNodeId1)

		var testNodeId2 int
		err = db.QueryRowx("INSERT INTO node (public_key, chain, network, created_on) VALUES ($1, $2, $3, $4) RETURNING node_id;",
			TestPublicKey2, commons.Bitcoin, commons.SigNet, time.Now().UTC()).Scan(&testNodeId2)
		if err != nil {
			cancel()
			return nil, nil, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", TestPublicKey2)
		}
		log.Debug().Msgf("Added test node with publicKey: %v nodeId: %v", TestPublicKey2, testNodeId2)

		_, err = db.Exec(`INSERT INTO node_connection_details
			(node_id, name, implementation, status_id, ping_system, created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6, $7);`,
			testNodeId1, "Node_1", commons.LND, commons.Active, 0, time.Now().UTC(), time.Now().UTC())
		if err != nil {
			cancel()
			return nil, nil, errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId1)
		}
		log.Debug().Msgf("Added test active node connection details with nodeId: %v", testNodeId1)

		_, err = db.Exec(`INSERT INTO node_connection_details
			(node_id, name, implementation, status_id, ping_system, created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6, $7);`,
			testNodeId2, "Node_2", commons.LND, commons.Active, 0, time.Now().UTC(), time.Now().UTC())
		if err != nil {
			cancel()
			return nil, nil, errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId2)
		}
		log.Debug().Msgf("Added test active node connection details with nodeId: %v", testNodeId2)

		go commons.ManagedChannelGroupCache(commons.ManagedChannelGroupChannel, ctx)
		go commons.ManagedSettingsCache(commons.ManagedSettingsChannel, ctx)
		go commons.ManagedNodeCache(commons.ManagedNodeChannel, ctx)
		go commons.ManagedChannelCache(commons.ManagedChannelChannel, ctx)

		err = settings.InitializeManagedSettingsCache(db)
		if err != nil {
			cancel()
			log.Fatal().Msgf("Problem initializing ManagedSettings cache: %v", err)
		}

		err = settings.InitializeManagedNodeCache(db)
		if err != nil {
			cancel()
			log.Fatal().Msgf("Problem initializing ManagedNode cache: %v", err)
		}
		log.Debug().Msgf("All Torq publicKeys: %v", commons.GetAllTorqPublicKeys(commons.Bitcoin, commons.SigNet))
		log.Debug().Msgf("All Torq nodeIds: %v", commons.GetAllTorqNodeIds(commons.Bitcoin, commons.SigNet))

		err = channels.InitializeManagedChannelCache(db)
		if err != nil {
			cancel()
			log.Fatal().Err(err).Msgf("Problem initializing ManagedChannel cache: %v", err)
		}
		log.Debug().Msgf("Channel publicKeys: %v", commons.GetChannelPublicKeys(commons.Bitcoin, commons.SigNet))
		log.Debug().Msgf("Channel nodeIds: %v", commons.GetChannelNodeIds(commons.Bitcoin, commons.SigNet))
		lndShortChannelId := uint64(1111)
		shortChannelId := channels.ConvertLNDShortChannelID(lndShortChannelId)
		testChannel1 := channels.Channel{
			ShortChannelID:         &shortChannelId,
			FirstNodeId:            commons.GetNodeIdByPublicKey(TestPublicKey1, commons.Bitcoin, commons.SigNet),
			SecondNodeId:           commons.GetNodeIdByPublicKey(TestPublicKey2, commons.Bitcoin, commons.SigNet),
			Capacity:               1_000_000,
			LNDShortChannelID:      &lndShortChannelId,
			FundingTransactionHash: TestFundingTransactionHash1,
			FundingOutputIndex:     3,
			Status:                 commons.Opening,
		}
		channelId, err := channels.AddChannelOrUpdateChannelStatus(db, testChannel1)
		if err != nil {
			cancel()
			log.Fatal().Err(err).Msgf("Problem adding channel %v", testChannel1)
		}
		log.Debug().Msgf("channel added with channelId: %v", channelId)

		lndShortChannelId = 2222
		shortChannelId = channels.ConvertLNDShortChannelID(lndShortChannelId)
		testChannel2 := channels.Channel{
			ShortChannelID:         &shortChannelId,
			FirstNodeId:            commons.GetNodeIdByPublicKey(TestPublicKey1, commons.Bitcoin, commons.SigNet),
			SecondNodeId:           commons.GetNodeIdByPublicKey(TestPublicKey2, commons.Bitcoin, commons.SigNet),
			Capacity:               1_000_000,
			LNDShortChannelID:      &lndShortChannelId,
			FundingTransactionHash: TestFundingTransactionHash2,
			FundingOutputIndex:     3,
			Status:                 commons.Opening,
		}
		channelId, err = channels.AddChannelOrUpdateChannelStatus(db, testChannel2)
		if err != nil {
			cancel()
			log.Fatal().Err(err).Msgf("Problem adding channel %v", testChannel2)
		}
		log.Debug().Msgf("channel added with channelId: %v", channelId)

		lndShortChannelId = 3333
		shortChannelId = channels.ConvertLNDShortChannelID(lndShortChannelId)
		testChannel3 := channels.Channel{
			ShortChannelID:         &shortChannelId,
			FirstNodeId:            commons.GetNodeIdByPublicKey(TestPublicKey1, commons.Bitcoin, commons.SigNet),
			SecondNodeId:           commons.GetNodeIdByPublicKey(TestPublicKey2, commons.Bitcoin, commons.SigNet),
			Capacity:               1_000_000,
			LNDShortChannelID:      &lndShortChannelId,
			FundingTransactionHash: TestFundingTransactionHash3,
			FundingOutputIndex:     3,
			Status:                 commons.Opening,
		}
		channelId, err = channels.AddChannelOrUpdateChannelStatus(db, testChannel3)
		if err != nil {
			cancel()
			log.Fatal().Err(err).Msgf("Problem adding channel %v", testChannel3)
		}
		log.Debug().Msgf("channel added with channelId: %v", channelId)

		lndShortChannelId = 4444
		shortChannelId = channels.ConvertLNDShortChannelID(lndShortChannelId)
		testChannel4 := channels.Channel{
			ShortChannelID:         &shortChannelId,
			FirstNodeId:            commons.GetNodeIdByPublicKey(TestPublicKey1, commons.Bitcoin, commons.SigNet),
			SecondNodeId:           commons.GetNodeIdByPublicKey(TestPublicKey2, commons.Bitcoin, commons.SigNet),
			Capacity:               1_000_000,
			LNDShortChannelID:      &lndShortChannelId,
			FundingTransactionHash: TestFundingTransactionHash4,
			FundingOutputIndex:     3,
			Status:                 commons.Opening,
		}
		channelId, err = channels.AddChannelOrUpdateChannelStatus(db, testChannel4)
		if err != nil {
			cancel()
			log.Fatal().Err(err).Msgf("Problem adding channel %v", testChannel4)
		}
		log.Debug().Msgf("channel added with channelId: %v", channelId)
		
		// initialise package level var for keeping state of subsciptions
		commons.RunningServices = make(map[commons.ServiceType]*commons.Services, 0)
		commons.RunningServices[commons.LndService] = &commons.Services{ServiceType: commons.LndService}
		commons.RunningServices[commons.VectorService] = &commons.Services{ServiceType: commons.VectorService}
		commons.RunningServices[commons.AmbossService] = &commons.Services{ServiceType: commons.AmbossService}
		commons.RunningServices[commons.TorqService] = &commons.Services{ServiceType: commons.TorqService}

	}

	return db, cancel, nil
}
