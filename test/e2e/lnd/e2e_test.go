package e2e

import (
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	if os.Getenv("E2E") == "" {
		log.Println("Skipping e2e tests as E2E environment variable not set")
		return
	}
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Hostname     string
	// Name         string
	// Repository   string
	// Tag          string
	// Env          []string
	// Entrypoint   []string
	// Cmd          []string
	// Mounts       []string
	// Links        []string
	// ExposedPorts []string
	// ExtraHosts   []string
	// CapAdd       []string
	// SecurityOpt  []string
	// DNS          []string
	// WorkingDir   string
	// NetworkID    string
	// Networks     []*Network // optional networks to join
	// Labels       map[string]string
	// Auth         dc.AuthConfiguration
	// PortBindings map[dc.Port][]dc.PortBinding
	// Privileged   bool
	// User         string
	// Tty          bool
	// Platform     string

	network, err := pool.CreateNetwork("lnd-test")
	if err != nil {
		log.Fatalf("Could not create network: %s", err)
	}

	removeAfterExitOption := func(config *dc.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = dc.RestartPolicy{
			Name: "no",
		}
	}

	btcdOptions := &dockertest.RunOptions{
		Name:     "btcd",
		Mounts:   []string{"e2e-shared:/rpc", "e2e-bitcoin:/data"},
		Env:      []string{"RPCUSER", "RPCPASS", "NETWORK=simnet", "DEBUG", "MINING_ADDRESS"},
		Networks: []*dockertest.Network{network},
	}
	// pulls an image, creates a container based on it and runs it
	btcd, err := pool.BuildAndRunWithOptions("./docker/btcd/Dockerfile", btcdOptions, removeAfterExitOption)
	if err != nil {
		log.Fatalf("Could not start btcd: %s", err)
	}

	aliceOptions := &dockertest.RunOptions{
		Name:     "lnd-alice",
		Mounts:   []string{"e2e-shared:/rpc", "e2e-lnd:/root/.lnd"},
		Env:      []string{"RPCUSER", "RPCPASS", "NETWORK=simnet", "CHAIN", "DEBUG"},
		Networks: []*dockertest.Network{network},
		Links:    []string{"btcd:blockchain"},
	}
	// pulls an image, creates a container based on it and runs it
	alice, err := pool.BuildAndRunWithOptions("./docker/lnd/Dockerfile", aliceOptions, removeAfterExitOption)
	if err != nil {
		log.Fatalf("Could not start alice: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	// if err := pool.Retry(func() error {
	// 	var err error
	// 	db, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql", resource.GetPort("3306/tcp")))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return db.Ping()
	// }); err != nil {
	// 	log.Fatalf("Could not connect to database: %s", err)
	// }

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(btcd); err != nil {
		log.Fatalf("Could not purge btcd: %s", err)
	}

	if err := pool.Purge(alice); err != nil {
		log.Fatalf("Could not purge alice: %s", err)
	}

	os.Exit(code)
}

func TestSomething(t *testing.T) {
	// db.Query()
}
