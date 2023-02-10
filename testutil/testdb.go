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

	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/pkg/broadcast"
	"github.com/lncapital/torq/pkg/commons"
)

const superuserName = "postgres"
const testDbPort = 5433
const testDBPrefix = "torq_test_"
const TestPublicKey1 = "PublicKey1"
const TestPublicKey2 = "PublicKey2"
const TestFundingOutputIndex = 3
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
func (srv *Server) NewTestDatabase(migrate bool) (*sqlx.DB, context.CancelFunc, broadcast.BroadcastServer, error) {

	// Create the new test database based on the main server connection
	dns, err := srv.createDatabase()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "srv.createDatabase(ctx)")
	}

	// Connect to the new test database
	db, err := sqlx.Open("postgres", dns)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "sqlx.Open(\"postgres\", %s)", dns)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Migrate the new test database
	err = database.MigrateUp(db)
	if err != nil {
		cancel()
		return nil, nil, nil, errors.Wrap(err, "Database Migrate Up")
	}

	testNodeId1, err := addNode(db, TestPublicKey1, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", TestPublicKey1)
	}

	testNodeId2, err := addNode(db, TestPublicKey2, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", TestPublicKey2)
	}

	err = addNodeConnectionDetails(db, testNodeId1, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId1)
	}

	err = addNodeConnectionDetails(db, testNodeId2, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId2)
	}

	lndShortChannelId := uint64(1111)
	shortChannelId := commons.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash1, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	lndShortChannelId = 2222
	shortChannelId = commons.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash2, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	lndShortChannelId = 3333
	shortChannelId = commons.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash3, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	lndShortChannelId = 4444
	shortChannelId = commons.ConvertLNDShortChannelID(lndShortChannelId)
	err = AddChannel(db, shortChannelId, lndShortChannelId, TestFundingTransactionHash4, TestFundingOutputIndex, testNodeId1, testNodeId2, cancel)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}

	//var serviceChannelGlobal = make(chan commons.ServiceChannelMessage)
	//var lightningRequestChannelGlobal = make(chan interface{})
	//var rebalanceRequestChannelGlobal = make(chan commons.RebalanceRequest)

	var serviceEventChannelGlobal = make(chan commons.ServiceEvent)
	var htlcEventChannelGlobal = make(chan commons.HtlcEvent)
	var forwardEventChannelGlobal = make(chan commons.ForwardEvent)
	var channelBalanceEventChannelGlobal = make(chan commons.ChannelBalanceEvent)
	var channelEventChannelGlobal = make(chan commons.ChannelEvent)
	var nodeGraphEventChannelGlobal = make(chan commons.NodeGraphEvent)
	var channelGraphEventChannelGlobal = make(chan commons.ChannelGraphEvent)
	var invoiceEventChannelGlobal = make(chan commons.InvoiceEvent)
	var paymentEventChannelGlobal = make(chan commons.PaymentEvent)
	var transactionEventChannelGlobal = make(chan commons.TransactionEvent)
	var peerEventChannelGlobal = make(chan commons.PeerEvent)
	var blockEventChannelGlobal = make(chan commons.BlockEvent)
	var webSocketResponseChannelGlobal = make(chan interface{})
	var lightningRequestChannelGlobal = make(chan interface{})
	var rebalanceRequestChannelGlobal = make(chan commons.RebalanceRequest)

	broadcaster := broadcast.NewBroadcastServer(ctx,
		serviceEventChannelGlobal, htlcEventChannelGlobal, forwardEventChannelGlobal,
		channelBalanceEventChannelGlobal, channelEventChannelGlobal, nodeGraphEventChannelGlobal, channelGraphEventChannelGlobal,
		invoiceEventChannelGlobal, paymentEventChannelGlobal, transactionEventChannelGlobal, peerEventChannelGlobal, blockEventChannelGlobal,
		webSocketResponseChannelGlobal, lightningRequestChannelGlobal, rebalanceRequestChannelGlobal)

	go commons.ManagedChannelGroupCache(commons.ManagedChannelGroupChannel, ctx)
	go commons.ManagedChannelStateCache(commons.ManagedChannelStateChannel, ctx, channelBalanceEventChannelGlobal)
	go commons.ManagedSettingsCache(commons.ManagedSettingsChannel, ctx)
	go commons.ManagedNodeCache(commons.ManagedNodeChannel, ctx)
	go commons.ManagedNodeAliasCache(commons.ManagedNodeAliasChannel, ctx)
	go commons.ManagedChannelCache(commons.ManagedChannelChannel, ctx)
	go commons.ManagedTaggedCache(commons.ManagedTaggedChannel, ctx)
	go commons.ManagedTriggerCache(commons.ManagedTriggerChannel, ctx)
	go tags.ManagedTagCache(tags.ManagedTagChannel, ctx)
	// TODO FIXME cyclic dependency so if you need this in tests then initialise it in the test
	//go automation.ManagedRebalanceCache(automation.ManagedRebalanceChannel, ctx)

	// initialise package level var for keeping state of subsciptions
	commons.RunningServices = make(map[commons.ServiceType]*commons.Services, 0)
	for _, serviceType := range commons.GetServiceTypes() {
		commons.RunningServices[serviceType] = &commons.Services{ServiceType: serviceType}
	}
	commons.RunningServices[commons.LndService].SetNodeConnectionDetailCustomSettings(testNodeId1, commons.NodeConnectionDetailCustomSettings(commons.NodeConnectionDetailCustomSettingsMax))
	commons.RunningServices[commons.LndService].SetNodeConnectionDetailCustomSettings(testNodeId2, commons.NodeConnectionDetailCustomSettings(commons.NodeConnectionDetailCustomSettingsMax))

	return db, cancel, broadcaster, nil
}

func addNode(db *sqlx.DB, testPublicKey string, cancel context.CancelFunc) (int, error) {
	var testNodeId int
	err := db.QueryRowx("INSERT INTO node (public_key, chain, network, created_on) VALUES ($1, $2, $3, $4) RETURNING node_id;",
		testPublicKey, commons.Bitcoin, commons.SigNet, time.Now().UTC()).Scan(&testNodeId)
	if err != nil {
		cancel()
		return 0, errors.Wrapf(err, "Inserting default node for testing with publicKey: %v", TestPublicKey1)
	}
	log.Debug().Msgf("Added test node with publicKey: %v nodeId: %v", TestPublicKey1, testNodeId)
	return testNodeId, nil
}

func addNodeConnectionDetails(db *sqlx.DB, testNodeId int, cancel context.CancelFunc) error {
	_, err := db.Exec(`INSERT INTO node_connection_details
			(node_id, name, implementation, status_id, ping_system, custom_settings, created_on, updated_on)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`,
		testNodeId, fmt.Sprintf("Node_%v", testNodeId), commons.LND, commons.Active, 0, 0, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		cancel()
		return errors.Wrapf(err, "Inserting default node_connection_details for testing with nodeId: %v", testNodeId)
	}
	log.Debug().Msgf("Added test active node connection details with nodeId: %v", testNodeId)
	return nil
}

func AddChannel(db *sqlx.DB, shortChannelId string, lndShortChannelId uint64,
	fundingTransactionHash string, foundingOutputIndex int,
	testNodeId1 int, testNodeId2 int, cancel context.CancelFunc) error {

	_, err := db.Exec(`INSERT INTO channel (
			  short_channel_id, funding_transaction_hash, funding_output_index,
			  closing_transaction_hash, closing_node_id,
			  lnd_short_channel_id, first_node_id, second_node_id, initiating_node_id, accepting_node_id, capacity, private,
			  status_id, created_on, updated_on
			) values (
			  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
			);`,
		shortChannelId, fundingTransactionHash, foundingOutputIndex, nil, nil,
		lndShortChannelId, testNodeId1, testNodeId2, nil, nil, 1_000_000,
		false, commons.Open, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		cancel()
		return errors.Wrapf(err, "Inserting default channel for testing with shortChannelId: %v", shortChannelId)
	}
	return nil
}
