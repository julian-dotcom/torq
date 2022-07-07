package e2e

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"

	// "github.com/docker/docker/api/types/container"
	"encoding/json"

	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
)

func TestMain(m *testing.M) {

	if os.Getenv("E2E") == "" {
		log.Println("Skipping e2e tests as E2E environment variable not set")
		return
	}

	ctx := context.Background()

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	tar, err := archive.TarWithOptions("docker/btcd", &archive.TarOptions{})
	if err != nil {
		log.Fatalf("Creating btcd archive: %s", err)
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"e2e/btcd"},
		Remove:     true,
	}

	res, err := cli.ImageBuild(ctx, tar, opts)
	if err != nil {
		log.Fatalf("Building btcd docker image: %s", err)
	}
	defer res.Body.Close()
	stdcopy.StdCopy(os.Stdout, os.Stderr, res.Body)

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	network, err := pool.CreateNetwork("lnd-test")
	if err != nil {
		log.Fatalf("Could not create network: %s", err)
	}

	defer func() {
		network.Close()
	}()

	pool.MaxWait = 1 * time.Minute

	removeAfterExitOption := func(config *dc.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = dc.RestartPolicy{
			Name: "no",
		}
	}

	btcdOptions := &dockertest.RunOptions{
		Name:     "blockchain",
		Mounts:   []string{"e2e-shared:/rpc", "e2e-bitcoin:/data"},
		Env:      []string{"NETWORK=simnet"},
		Networks: []*dockertest.Network{network},
	}
	// pulls an image, creates a container based on it and runs it
	btcd, err := pool.BuildAndRunWithOptions("./docker/btcd/Dockerfile", btcdOptions)
	if err != nil {
		log.Fatalf("Could not start btcd: %s", err)
	}

	aliceOptions := &dockertest.RunOptions{
		Name:     "lnd-alice",
		Mounts:   []string{"e2e-shared:/rpc", "e2e-lnd:/root/.lnd"},
		Env:      []string{"NETWORK=simnet"},
		Networks: []*dockertest.Network{network},
	}
	// pulls an image, creates a container based on it and runs it
	alice, err := pool.BuildAndRunWithOptions("./docker/lnd/Dockerfile", aliceOptions, removeAfterExitOption)
	if err != nil {
		log.Fatalf("Could not start alice: %s", err)
	}

	// var aliceAddress string
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		//docker exec -it alice lncli --network=simnet newaddress np2wkh

		c := types.ExecConfig{AttachStdout: true, AttachStderr: true,
			Cmd: []string{"lncli", "--network=simnet", "newaddress", "np2wkh"}}
		execID, _ := cli.ContainerExecCreate(ctx, alice.Container.ID, c)

		res, er := cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
		if er != nil {
			log.Printf("Container exec attach on alice: %v\n", err)
		}

		err = cli.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{})
		if err != nil {
			log.Printf("Container exec start on alice: %v\n", err)
		}
		var bufStdout bytes.Buffer
		// stdout := bufio.NewWriter(&bufStdout)
		var bufStderr bytes.Buffer

		// stdcopy.StdCopy(os.Stdout, stderr, res.Reader)
		stdcopy.StdCopy(&bufStdout, &bufStderr, res.Reader)

		var address struct {
			Address string `json:"address"`
		}
		err = json.Unmarshal(bufStdout.Bytes(), &address)
		if err != nil {
			return errors.New("RPC not returning valid JSON")
		}

		if address.Address == "" {
			return errors.New("Not valid address")
		}
		// content, _, _ := res.Reader.ReadLine()
		// log.Println(string(content))
		log.Println("Alice receive address created")
		// aliceAddress = address.Address
		return nil

	}); err != nil {
		log.Fatalf("Could exec command on Alice: %s", err)
	}

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
