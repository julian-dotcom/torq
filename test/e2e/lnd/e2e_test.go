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
	"strconv"
	"strings"
	"testing"
	"time"
	// "github.com/ory/dockertest/v3"
	// dc "github.com/ory/dockertest/v3/docker"
)

const defautDelayMS = 1000         // 1s
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

	log.Println("Building btcd image from dockerfile")
	buildImage("docker/btcd/", "e2e/btcd", cli, ctx, false)
	log.Println("Building lnd image from dockerfile")
	buildImage("docker/lnd/", "e2e/lnd", cli, ctx, false)

	networkingConfig := createNetwork(ctx, cli, "e2e")

	log.Println("Starting btcd")
	_ = createContainer(cli, ctx, "e2e/btcd", btcdName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			// "e2e-bitcoin:/data",
		}, networkingConfig)

	log.Println("Starting Alice")
	alice := createContainer(cli, ctx, "e2e/lnd", aliceName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			// "e2e-lnd-alice:/root/.lnd",
		}, networkingConfig)

	log.Println("Starting Bob")
	bob := createContainer(cli, ctx, "e2e/lnd", bobName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			// "e2e-lnd-bob:/root/.lnd",
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
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
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
			// "e2e-bitcoin:/data",
		}, networkingConfig)

	log.Println("Generate 400 blocks (we need at least \"100 >=\" blocks because of coinbase block maturity and \"300 ~=\" in order to activate segwit)")

	err = mineBlocks(ctx, cli, btcd, 400)
	if err != nil {
		log.Fatalf("Mining blocks: %v", err)
	}

	log.Println("Blocks mined")
	log.Println("Checking that segwit is active")

	err = retry(func() error {
		var blockchainInfo struct {
			Bip9Softforks struct {
				Segwit struct {
					Status string `json:"status"`
				} `json:"segwit"`
			} `json:"bip9_softforks"`
		}
		cmd := []string{"/start-btcctl.sh", "getblockchaininfo"}
		err = execJSONReturningCommand(cli, ctx, btcd, cmd, &blockchainInfo)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on btcd %s", btcd.ID)
		}
		if blockchainInfo.Bip9Softforks.Segwit.Status != "active" {
			return errors.New("Segwit not active")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("btcd checking segwit is active: %v", err)
	}
	log.Println("Segwit is active")
	log.Println("Checking Alice's balance")

	aliceBalance, err := getOnchainBalance(ctx, cli, alice)
	if err != nil {
		log.Fatalf("Getting Alice's balance: %v", err)
	}

	log.Printf("Alice's onchain balance is: %s\n", aliceBalance)

	log.Println("Get Bob's pubkey")

	var getInfo struct {
		IdentityPubkey string `json:"identity_pubkey"`
	}
	var bobPubkey string
	err = retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "getinfo"}
		err = execJSONReturningCommand(cli, ctx, bob, cmd, &getInfo)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Bob %s", bob.ID)
		}
		if getInfo.IdentityPubkey == "" {
			return errors.New("Invalid Pubkey")
		}
		bobPubkey = getInfo.IdentityPubkey
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Getting Bob's pubkey: %v", err)
	}
	log.Printf("Bob's pubkey is: %s\n", bobPubkey)

	bobInspection, err := cli.ContainerInspect(ctx, bob.ID)
	if err != nil {
		log.Fatalf("Getting Bob's IP Address: %v", err)
	}
	bobIPAddress := bobInspection.NetworkSettings.Networks["e2e"].IPAddress
	log.Println("Bob's IP address is:")
	log.Println(bobIPAddress)

	log.Println("Connecting Bob to Alice")

	err = retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "connect", bobPubkey + "@" + bobIPAddress}
		var stderr bytes.Buffer
		_, stderr, err = execCommand(ctx, cli, alice, cmd)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
		}
		if len(stderr.Bytes()) > 0 {
			return errors.New("Bob not connected to Alice")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Checking that Bob is a peer of Alice: %v", err)
	}

	log.Println("Verifing Bob is a peer of Alice")

	var listPeers struct {
		Peers []struct {
			Pubkey string `json:"pub_key"`
		} `json:"peers"`
	}
	err = retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "listpeers"}
		err = execJSONReturningCommand(cli, ctx, alice, cmd, &listPeers)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
		}
		if len(listPeers.Peers) == 0 {
			return errors.New("Bob not a peer")
		}
		if listPeers.Peers[0].Pubkey != bobPubkey {
			return errors.New("Bob not a peer")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Checking that Bob is a peer of Alice: %v", err)
	}

	log.Println("Bob confirmed as peer of Alice")

	log.Println("Getting Alice's pubkey")
	var alicePubkey string
	err = retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "getinfo"}
		err = execJSONReturningCommand(cli, ctx, alice, cmd, &getInfo)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
		}
		if getInfo.IdentityPubkey == "" {
			return errors.New("Invalid Pubkey")
		}
		alicePubkey = getInfo.IdentityPubkey
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Getting Alice's pubkey: %v", err)
	}

	log.Printf("Alice's pubkey is: %s\n", alicePubkey)

	log.Println("Verifing Alice is a peer of Bob")

	err = retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "listpeers"}
		err = execJSONReturningCommand(cli, ctx, bob, cmd, &listPeers)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on bob %s", bob.ID)
		}
		if len(listPeers.Peers) == 0 {
			return errors.New("Alice not a peer")
		}
		if listPeers.Peers[0].Pubkey != alicePubkey {
			return errors.New("Alice not a peer")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Checking that Alice is a peer of Bob: %v", err)
	}
	log.Println("Alice confirmed as peer of Bob")

	log.Println("Create the Alice<->Bob channel")

	var aliceFundingTxId string
	err = retry(func() error {
		var openChannel struct {
			FundingTxId string `json:"funding_txid"`
		}
		cmd := []string{"lncli", "--network=simnet", "openchannel", "--node_key=" + bobPubkey, "--local_amt=1000000"}
		err = execJSONReturningCommand(cli, ctx, alice, cmd, &openChannel)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
		}
		if openChannel.FundingTxId == "" {
			return errors.New("Channel not created")
		}
		aliceFundingTxId = openChannel.FundingTxId
		return nil
	}, 4500, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Creating Alice<->Bob channel: %v", err)
	}
	log.Printf("Funding transaction ID: %s\n", aliceFundingTxId)

	log.Println("Include funding transaction in block thereby opening the channel")

	err = mineBlocks(ctx, cli, btcd, 3)
	if err != nil {
		log.Fatalf("Mining blocks: %v", err)
	}

	log.Println("Blocks mined")
	log.Println("Checking channel with Bob is open")

	var aliceBobChannelPoint string
	err = retry(func() error {
		var listChannels struct {
			Channels []struct {
				ChannelPoint string `json:"channel_point"`
			} `json:"channels"`
		}
		cmd := []string{"lncli", "--network=simnet", "listchannels"}
		err = execJSONReturningCommand(cli, ctx, alice, cmd, &listChannels)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
		}
		if len(listChannels.Channels) == 0 {
			return errors.New("Channel not open")
		}
		if listChannels.Channels[0].ChannelPoint == "" {
			return errors.New("Channel not open")
		}
		aliceBobChannelPoint = listChannels.Channels[0].ChannelPoint
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Creating Alice<->Bob channel: %v", err)
	}
	log.Printf("Alice<->Bob channel point: %s\n", aliceBobChannelPoint)

	log.Println("Generating invoice for payment to Bob")

	var bobEncodedInvoice string
	err = retry(func() error {
		var addInvoice struct {
			EncodedPayReq string `json:"payment_request"`
		}
		cmd := []string{"lncli", "--network=simnet", "addinvoice", "--amt=100000"}
		err = execJSONReturningCommand(cli, ctx, bob, cmd, &addInvoice)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Bob %s", bob.ID)
		}
		if addInvoice.EncodedPayReq == "" {
			return errors.New("Invoice not generated")
		}
		bobEncodedInvoice = addInvoice.EncodedPayReq
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Creating Bob invoice: %v", err)
	}

	log.Printf("Encoded payment request: %s\n", bobEncodedInvoice)

	log.Println("Alice paying invoice sending payment to Bob")

	err = retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "sendpayment", "--force", "--pay_req=" + bobEncodedInvoice}
		var stderr bytes.Buffer
		_, stderr, err = execCommand(ctx, cli, alice, cmd)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
		}
		if len(stderr.Bytes()) > 0 {
			return errors.New("Payment not sent")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Sending Alice->Bob payment: %v", err)
	}

	log.Println("Checking payment received by Bob")
	var bobChannelBalance string
	err = retry(func() error {
		var channelBalance struct {
			Balance string `json:"balance"`
		}
		cmd := []string{"lncli", "--network=simnet", "channelbalance"}
		err = execJSONReturningCommand(cli, ctx, bob, cmd, &channelBalance)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Bob %s", bob.ID)
		}
		if channelBalance.Balance == "" {
			return errors.New("Payment not received")
		}
		if channelBalance.Balance == "0" {
			return errors.New("Payment not received")
		}
		bobChannelBalance = channelBalance.Balance
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Checking Bob's balance: %v", err)
	}

	log.Println("Payment received by Bob")
	log.Printf("Bob's channel balance: %s\n", bobChannelBalance)

	log.Println("Close Alice<->Bob channel to gain on chain funds for Bob")

	var aliceBobClosingTxId string
	err = retry(func() error {
		var closeChannel struct {
			ClosingTxId string `json:"closing_txid"`
		}
		fundingTxId := aliceBobChannelPoint[:strings.IndexByte(aliceBobChannelPoint, ':')]
		outputIndex := aliceBobChannelPoint[strings.IndexByte(aliceBobChannelPoint, ':')+1:]
		cmd := []string{"lncli", "--network=simnet", "closechannel", "--funding_txid=" + fundingTxId, "--output_index=" + outputIndex}
		err = execJSONReturningCommand(cli, ctx, alice, cmd, &closeChannel)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", alice.ID)
		}
		if closeChannel.ClosingTxId == "" {
			return errors.New("Channel not closed")
		}
		aliceBobClosingTxId = closeChannel.ClosingTxId
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		log.Fatalf("Closing Alice<->Bob channel: %v", err)
	}

	log.Printf("Alice<->Bob channel closing transaction id: %s\n", aliceBobClosingTxId)

	log.Println("Mining some blocks to confirm closing transaction")

	err = mineBlocks(ctx, cli, btcd, 3)
	if err != nil {
		log.Fatalf("Mining blocks: %v", err)
	}

	bobOnChainBalance, err := getOnchainBalance(ctx, cli, bob)
	if err != nil {
		log.Fatalf("Getting Bob's balance: %v", err)
	}
	log.Printf("Bob's onchain balance: %s\n", bobOnChainBalance)

	code := m.Run()

	// try to cleanup after run
	// can't defer this as os.Exit doesn't care for defer
	// cleanup(cli, ctx)
	os.Exit(code)
}

func getOnchainBalance(ctx context.Context, cli *client.Client, container dockercontainer.ContainerCreateCreatedBody) (balance string, err error) {
	err = retry(func() error {
		var walletBalance struct {
			ConfirmedBalance string `json:"confirmed_balance"`
		}
		cmd := []string{"lncli", "--network=simnet", "walletbalance"}
		err := execJSONReturningCommand(cli, ctx, container, cmd, &walletBalance)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", container.ID)
		}
		if walletBalance.ConfirmedBalance == "" {
			return errors.New("Balance not confirmed")
		}
		if walletBalance.ConfirmedBalance == "0" {
			return errors.New("Balance not confirmed")
		}
		balance = walletBalance.ConfirmedBalance
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Getting balance")
	}
	return balance, nil

}

func mineBlocks(ctx context.Context, cli *client.Client, btcd dockercontainer.ContainerCreateCreatedBody, numberOfBlocks int) error {
	err := retry(func() error {
		var output []string
		cmd := []string{"/start-btcctl.sh", "generate", strconv.Itoa(numberOfBlocks)}
		err := execJSONReturningCommand(cli, ctx, btcd, cmd, &output)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on btcd %s", btcd.ID)
		}
		if len(output) == 0 {
			return errors.New("Blocks not mined")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		errors.Wrap(err, "btcd mining blocks")
	}
	return nil
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

	// stdcopy.StdCopy(os.Stdout, os.Stderr, res.Reader)
	stdcopy.StdCopy(&bufStdout, &bufStderr, res.Reader)
	// DEBUG Tip: uncomment below to see raw output of commands
	// log.Printf("%s\n", string(bufStdout.Bytes()))
	// log.Printf("%s\n", string(bufStderr.Bytes()))
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
		log.Println("Checking...")
		time.Sleep(time.Duration(delayMilliseconds) * time.Millisecond)
		totalWaited += delayMilliseconds
	}
	return nil
}

func cleanup(cli *client.Client, ctx context.Context) {
	findAndRemoveContainer(cli, ctx, btcdName)
	findAndRemoveContainer(cli, ctx, aliceName)
	findAndRemoveContainer(cli, ctx, bobName)
	findAndRemoveContainer(cli, ctx, carolName)
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

func buildImage(path string, name string, cli *client.Client, ctx context.Context, printOutput bool) {
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
	if printOutput {
		err = printBuildOutput(res.Body)
		if err != nil {
			log.Fatalf("Printing build output for %s docker image: %v", name, err)
		}
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
