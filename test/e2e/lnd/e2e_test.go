package e2e

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"log"
	"os"
	"testing"
	"time"
	// "github.com/ory/dockertest/v3"
	// dc "github.com/ory/dockertest/v3/docker"
)

const defautDelayMS = 500          // 500ms
const defaultMaxDurationMS = 30000 // 30s

const networkName = "e2e"
const aliceName = "e2e-alice"
const bobName = "e2e-bob"
const carolName = "e2e-carol"
const btcdName = "e2e-btcd"

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

	// cleanup any old networks or containers that might have been left around from a failed run
	log.Println("Checking if any old container or networks are present")
	cleanup(cli, ctx)

	buildImage("docker/btcd/", "e2e/btcd", cli, ctx)
	buildImage("docker/lnd/", "e2e/lnd", cli, ctx)

	networkingConfig := createNetwork(ctx, cli, "e2e")

	_ = createContainer(cli, ctx, "e2e/btcd", btcdName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			"e2e-bitcoin:/data",
		}, networkingConfig)

	alice := createContainer(cli, ctx, "e2e/lnd", aliceName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			"e2e-lnd:/root/.lnd",
		}, networkingConfig)

	// Example looking at container logs
	// out, err := cli.ContainerLogs(ctx, btcd.ID, types.ContainerLogsOptions{ShowStdout: true})
	// if err != nil {
	// 	panic(err)
	// }
	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	log.Println("Creating new mining address on Alice")

	var aliceAddress string
	err = retry(func() error {
		var address struct {
			Address string `json:"address"`
		}
		cmd := []string{"lncli", "--network=simnet", "newaddress", "np2wkh"}
		err = execJSONReturningCommand(cli, ctx, alice, cmd, &address)
		if err != nil {
			errors.Wrapf(err, "Running exec command on alice %s", alice.ID)
		}
		if address.Address == "" {
			return errors.New("Not a valid address")
		}
		aliceAddress = address.Address
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Getting alice mining address: %v", err)
	}
	log.Println("Alice receive address created")
	log.Println(aliceAddress)

	log.Println("Recreating btcd container with Alice's mining address")
	findAndRemoveContainer(cli, ctx, btcdName)
	log.Println("Starting new btcd container")
	btcd := createContainer(cli, ctx, "e2e/btcd", btcdName,
		[]string{
			"NETWORK=simnet",
			"MINING_ADDRESS=" + aliceAddress},
		[]string{
			"e2e-shared:/rpc",
			"e2e-bitcoin:/data",
		}, networkingConfig)

	log.Println("Generate 400 blocks (we need at least \"100 >=\" blocks because of coinbase block maturity and \"300 ~=\" in order to activate segwit)")

	err = retry(func() error {
		var output []string
		cmd := []string{"/start-btcctl.sh", "generate", "400"}
		err = execJSONReturningCommand(cli, ctx, btcd, cmd, &output)
		log.Printf("%s", output)
		if err != nil {
			errors.Wrapf(err, "Running exec command on btcd %s", btcd.ID)
		}
		if len(output) == 0 {
			return errors.New("Blocks not mined")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("btcd mining blocks: %v", err)
	}
	// bufStdout, _, err := execCommand(ctx, cli, container, cmd)
	// if err != nil {
	// 	return errors.Wrap(err, "Exec command on container")
	// }
	//
	code := m.Run()

	// try to cleanup after run
	// can't defer this as os.Exit doesn't care for defer
	// cleanup(cli, ctx)
	os.Exit(code)
}

func execJSONReturningCommand(cli *client.Client, ctx context.Context,
	container dockercontainer.ContainerCreateCreatedBody,
	cmd []string, returnObject interface{}) error {

	bufStdout, _, err := execCommand(ctx, cli, container, cmd)
	if err != nil {
		return errors.Wrap(err, "Exec command on container")
	}

	err = json.Unmarshal(bufStdout.Bytes(), returnObject)
	if err != nil {
		return errors.Wrap(err, "json unmarshal")
	}
	return nil
}

func execCommand(ctx context.Context, cli *client.Client,
	container dockercontainer.ContainerCreateCreatedBody,
	cmd []string) (bufStdout bytes.Buffer, bufStderr bytes.Buffer, err error) {

	c := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: cmd}
	execID, _ := cli.ContainerExecCreate(ctx, container.ID, c)

	res, err := cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return bufStdout, bufStderr, errors.Wrap(err, "Container exec start")
	}

	err = cli.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return bufStdout, bufStderr, errors.Wrap(err, "Container exec start")
	}

	// stdcopy.StdCopy(os.Stdout, stderr, res.Reader)
	stdcopy.StdCopy(&bufStdout, &bufStderr, res.Reader)
	return bufStdout, bufStderr, nil
}

func createContainer(cli *client.Client, ctx context.Context,
	image string, name string, env []string, binds []string,
	networkingConfig network.NetworkingConfig) dockercontainer.ContainerCreateCreatedBody {

	btcd, err := cli.ContainerCreate(ctx, &dockercontainer.Config{
		Image: image,
		Env:   env,
	}, &dockercontainer.HostConfig{
		Binds: binds,
	}, &networkingConfig, nil, name)
	if err != nil {
		log.Fatalf("Creating %s container: %v", name, err)
	}
	if err := cli.ContainerStart(ctx, btcd.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatalf("Starting %s container: %v", name, err)
	}
	return btcd
}

func createNetwork(ctx context.Context, cli *client.Client, name string) network.NetworkingConfig {
	e2eNetwork, err := cli.NetworkCreate(ctx, name, types.NetworkCreate{})
	if err != nil {
		log.Fatalf("Creating %s network: %v", name, err)
	}
	networkingConfig := network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	networkingConfig.EndpointsConfig[e2eNetwork.ID] = &network.EndpointSettings{Links: []string{"e2e-btcd:blockchain"}}
	return networkingConfig
}

func retry(operation func() error, delayMilliseconds int, maxWaitMilliseconds int) error {
	totalWaited := 0
	for {
		if totalWaited > maxWaitMilliseconds {
			return errors.New("Exceeded maximum wait period")
		}
		if operation() == nil {
			break
		}
		log.Println("Waiting...")
		time.Sleep(time.Duration(delayMilliseconds) * time.Millisecond)
		totalWaited += delayMilliseconds
	}
	return nil
}

func cleanup(cli *client.Client, ctx context.Context) {
	findAndRemoveContainer(cli, ctx, btcdName)
	findAndRemoveContainer(cli, ctx, aliceName)
	findAndRemoveNetwork(cli, ctx, networkName)
}

func findAndRemoveNetwork(cli *client.Client, ctx context.Context, name string) {
	network, err := findNetworkByName(cli, ctx, name)
	if err != nil {
		log.Fatalf("Removing old %s network: %v", name, err)
	}
	if network != nil {
		log.Printf("Old %s network found; removing\n", name)
		if err := cli.NetworkRemove(ctx, network.ID); err != nil {
			log.Fatalf("Removing old %s network: %v", name, err)
		}
	}
}

func findNetworkByName(cli *client.Client, ctx context.Context, name string) (*types.NetworkResource, error) {
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	for _, network := range networks {
		if network.Name == name {
			return &network, nil
		}
	}
	return nil, nil
}

func findAndRemoveContainer(cli *client.Client, ctx context.Context, name string) {
	container, err := findContainerByName(cli, ctx, name)
	if err != nil {
		log.Fatalf("Removing %s container: %v", name, err)
	}
	if container != nil {
		log.Printf("%s container found; removing\n", name)

		if container.State == "running" {
			if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
				log.Fatalf("Stopping %s container: %v", name, err)
			}
		}
		if err := cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
			log.Fatalf("Removing %s container: %v", name, err)
		}
	}
}

func findContainerByName(cli *client.Client, ctx context.Context, name string) (*types.Container, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		for _, containerName := range container.Names {
			// internal docker names have leading slashes; trim off
			if containerName[1:] == name {
				return &container, nil
			}
		}
	}
	return nil, nil
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

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
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
