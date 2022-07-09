package e2e

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	// "time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	// "github.com/ory/dockertest/v3"
	// dc "github.com/ory/dockertest/v3/docker"
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

	e2eNetwork, err := cli.NetworkCreate(ctx, "e2e", types.NetworkCreate{})
	if err != nil {
		log.Fatalf("Could not create e2e network: %s", err)
	}

	buildImage("docker/btcd/", "e2e/btcd", cli, ctx)

	buildImage("docker/lnd/", "e2e/lnd", cli, ctx)

	networkingConfig := network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	networkingConfig.EndpointsConfig[e2eNetwork.ID] = &network.EndpointSettings{Links: []string{"e2e-btcd:blockchain"}}

	btcd, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "e2e/btcd",
		Env:   []string{"NETWORK=simnet"},
	}, &container.HostConfig{
		Binds: []string{
			"e2e-shared:/rpc",
			"e2e-bitcoin:/data",
		},
	}, &networkingConfig, nil, "e2e-btcd")
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, btcd.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	alice, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "e2e/lnd",
		Env:   []string{"NETWORK=simnet"},
	}, &container.HostConfig{
		Binds: []string{
			"e2e-shared:/rpc",
			"e2e-lnd:/root/.lnd",
		},
	}, &networkingConfig, nil, "e2e-alice")
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, alice.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	// select {
	// case err := <-errCh:
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// case <-statusCh:
	// }

	out, err := cli.ContainerLogs(ctx, btcd.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	// pool, err := dockertest.NewPool("")
	// if err != nil {
	// 	log.Fatalf("Could not connect to docker: %s", err)
	// }

	// network, err := pool.CreateNetwork("lnd-test")
	// if err != nil {
	// 	log.Fatalf("Could not create network: %s", err)
	// }

	// defer func() {
	// 	network.Close()
	// }()

	// pool.MaxWait = 1 * time.Minute

	// removeAfterExitOption := func(config *dc.HostConfig) {
	// 	// set AutoRemove to true so that stopped container goes away by itself
	// 	config.AutoRemove = true
	// 	config.RestartPolicy = dc.RestartPolicy{
	// 		Name: "no",
	// 	}
	// }

	// btcdOptions := &dockertest.RunOptions{
	// 	Name:     "blockchain",
	// 	Mounts:   []string{"e2e-shared:/rpc", "e2e-bitcoin:/data"},
	// 	Env:      []string{"NETWORK=simnet"},
	// 	Networks: []*dockertest.Network{network},
	// }
	// // pulls an image, creates a container based on it and runs it
	// btcd, err := pool.BuildAndRunWithOptions("./docker/btcd/Dockerfile", btcdOptions)
	// if err != nil {
	// 	log.Fatalf("Could not start btcd: %s", err)
	// }

	// aliceOptions := &dockertest.RunOptions{
	// 	Name:     "lnd-alice",
	// 	Mounts:   []string{"e2e-shared:/rpc", "e2e-lnd:/root/.lnd"},
	// 	Env:      []string{"NETWORK=simnet"},
	// 	Networks: []*dockertest.Network{network},
	// }
	// // pulls an image, creates a container based on it and runs it
	// alice, err := pool.BuildAndRunWithOptions("./docker/lnd/Dockerfile", aliceOptions, removeAfterExitOption)
	// if err != nil {
	// 	log.Fatalf("Could not start alice: %s", err)
	// }

	// var aliceAddress string
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	// if err := pool.Retry(func() error {
	//docker exec -it alice lncli --network=simnet newaddress np2wkh

loop:
	c := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: []string{"lncli", "--network=simnet", "newaddress", "np2wkh"}}
	execID, _ := cli.ContainerExecCreate(ctx, alice.ID, c)

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
		// return errors.New("RPC not returning valid JSON")
		log.Println("loop")
		goto loop
	}

	if address.Address == "" {
		// return errors.New("Not valid address")
		log.Println("loop")
		goto loop
	}
	// content, _, _ := res.Reader.ReadLine()
	// log.Println(string(content))
	log.Println("Alice receive address created")
	log.Println(address.Address)
	// aliceAddress = address.Address

	// }); err != nil {
	// 	log.Fatalf("Could exec command on Alice: %s", err)
	// }

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// if err := pool.Purge(btcd); err != nil {
	// 	log.Fatalf("Could not purge btcd: %s", err)
	// }

	// if err := pool.Purge(alice); err != nil {
	// 	log.Fatalf("Could not purge alice: %s", err)
	// }

	os.Exit(code)
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func buildImage(path string, name string, cli *client.Client, ctx context.Context) {
	tar, err := archive.TarWithOptions(path, &archive.TarOptions{})
	if err != nil {
		log.Fatalf("Creating %s archive: %v", name, err)
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{name},
		Remove:     true,
	}

	res, err := cli.ImageBuild(ctx, tar, opts)
	if err != nil {
		log.Fatalf("Building %s docker image: %v", name, err)
	}
	defer res.Body.Close()
	err = printBuildOutput(res.Body)
	if err != nil {
		log.Fatalf("Printing build output for %s docker image: %v", name, err)
	}
}

func printBuildOutput(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func TestSomething(t *testing.T) {
	// db.Query()
}
