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

const defautDelayMS = 2000          // 2s
const defaultMaxDurationMS = 120000 // 60s

const networkName = "e2e"
const aliceName = "e2e-alice"
const bobName = "e2e-bob"
const carolName = "e2e-carol"
const btcdName = "e2e-btcd"

var ctx context.Context
var cli *client.Client
var btcd dockercontainer.ContainerCreateCreatedBody
var alice dockercontainer.ContainerCreateCreatedBody
var bob dockercontainer.ContainerCreateCreatedBody
var carol dockercontainer.ContainerCreateCreatedBody

func TestMain(m *testing.M) {

	if os.Getenv("E2E") == "" {
		log.Println("Skipping e2e tests as E2E environment variable not set")
		return
	}

	ctx = context.Background()

	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// cleanup any old networks or containers that might have been left around from a failed run
	log.Println("Checking if any old containers or networks are present")
	cleanup(cli, ctx)

	log.Println("Building btcd image from dockerfile")
	buildImage("docker/btcd/", "e2e/btcd", cli, ctx)
	log.Println("Building lnd image from dockerfile")
	buildImage("docker/lnd/", "e2e/lnd", cli, ctx)

	networkingConfig := createNetwork(ctx, cli, "e2e")

	log.Println("Starting btcd")
	_ = createContainer(cli, ctx, "e2e/btcd", btcdName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			// "e2e-bitcoin:/data",
		}, networkingConfig)

	log.Println("Starting Alice")
	alice = createContainer(cli, ctx, "e2e/lnd", aliceName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			// "e2e-lnd-alice:/root/.lnd",
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
		err := execJSONReturningCommand(ctx, cli, alice, cmd, &address)
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
	btcd = createContainer(cli, ctx, "e2e/btcd", btcdName,
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
		err := execJSONReturningCommand(ctx, cli, btcd, cmd, &blockchainInfo)
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

	// start Bob and Carol AFTER btcd has restarted
	log.Println("Starting Bob")
	bob = createContainer(cli, ctx, "e2e/lnd", bobName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			// "e2e-lnd-bob:/root/.lnd",
		}, networkingConfig)

	log.Println("Get Bob's pubkey")

	bobInspection, err := cli.ContainerInspect(ctx, bob.ID)
	if err != nil {
		log.Fatalf("Getting Bob's IP Address: %v", err)
	}
	bobIPAddress := bobInspection.NetworkSettings.Networks["e2e"].IPAddress
	log.Println("Bob's IP address is:")
	log.Println(bobIPAddress)

	bobPubkey, err := getPubKey(ctx, cli, bob)
	if err != nil {
		log.Fatalf("Getting Bob's pubkey: %v", err)
	}
	log.Printf("Bob's pubkey is: %s\n", bobPubkey)

	log.Println("Connecting Bob to Alice")

	err = connectPeer(ctx, cli, alice, bobPubkey, bobIPAddress)
	if err != nil {
		log.Fatalf("Connecting Bob to Alice: %v", err)
	}

	log.Println("Verifing Bob is a peer of Alice")

	bobPeerExists, err := checkPeerExists(ctx, cli, alice, bobPubkey)
	if err != nil || !bobPeerExists {
		log.Fatalf("Checking that Bob is a peer of Alice: %v", err)
	}

	log.Println("Bob confirmed as peer of Alice")

	log.Println("Getting Alice's pubkey")
	alicePubkey, err := getPubKey(ctx, cli, alice)
	if err != nil {
		log.Fatalf("Getting Alice's pubkey: %v", err)
	}

	log.Printf("Alice's pubkey is: %s\n", alicePubkey)

	log.Println("Verifing Alice is a peer of Bob")

	alicePeerExists, err := checkPeerExists(ctx, cli, bob, alicePubkey)
	if err != nil || !alicePeerExists {
		log.Fatalf("Checking that Alice is a peer of Bob: %v", err)
	}
	log.Println("Alice confirmed as peer of Bob")

	log.Println("Create the Alice<->Bob channel")

	aliceBobChannelPoint, err := createChannel(ctx, cli, alice, bobPubkey, "12000000", btcd)
	if err != nil {
		log.Fatalf("Creating Alice<->Bob channel: %v", err)
	}

	log.Println("Generating invoice for payment to Bob")

	bobEncodedInvoice, err := generateInvoice(ctx, cli, bob, "1000000")
	if err != nil {
		log.Fatalf("Creating Bob invoice: %v", err)
	}

	log.Printf("Encoded payment request: %s\n", bobEncodedInvoice)

	log.Println("Alice paying invoice sending payment to Bob")

	err = payInvoice(ctx, cli, alice, bobEncodedInvoice)
	if err != nil {
		log.Fatalf("Sending Alice->Bob payment: %v", err)
	}

	log.Println("Checking payment received by Bob")
	bobChannelBalance, err := getChannelBalance(ctx, cli, bob)
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
		err := execJSONReturningCommand(ctx, cli, alice, cmd, &closeChannel)
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

	log.Println("Starting Carol")
	carol = createContainer(cli, ctx, "e2e/lnd", carolName,
		[]string{"NETWORK=simnet"},
		[]string{
			"e2e-shared:/rpc",
			// "e2e-lnd-bob:/root/.lnd",
		}, networkingConfig)

	log.Println("Getting Carol's pubkey")
	carolPubkey, err := getPubKey(ctx, cli, carol)
	if err != nil {
		log.Fatalf("Getting Carol's pubkey: %v", err)
	}
	log.Printf("Carol's pubkey: %s\n", carolPubkey)

	carolInspection, err := cli.ContainerInspect(ctx, carol.ID)
	if err != nil {
		log.Fatalf("Getting Carol's IP Address: %v", err)
	}
	carolIPAddress := carolInspection.NetworkSettings.Networks["e2e"].IPAddress
	log.Println("Carol's IP address is:")
	log.Println(carolIPAddress)

	log.Println("Connecting Carol to Bob")

	err = connectPeer(ctx, cli, bob, carolPubkey, carolIPAddress)
	if err != nil {
		log.Fatalf("Connecting Carol to Bob: %v", err)
	}

	log.Println("Verifing Carol is a peer of Bob")

	carolPeerExists, err := checkPeerExists(ctx, cli, bob, carolPubkey)
	if err != nil || !carolPeerExists {
		log.Fatalf("Checking that Carol is a peer of Bob: %v", err)
	}

	log.Println("Carol confirmed as peer of Bob")

	log.Println("Verifing Bob is a peer of Carol")
	carolBobPeerExists, err := checkPeerExists(ctx, cli, carol, bobPubkey)
	if err != nil || !carolBobPeerExists {
		log.Fatalf("Checking that Bob is a peer of Carol: %v", err)
	}
	log.Println("Bob confirmed as peer of Carol")

	err = mineBlocks(ctx, cli, btcd, 30)
	if err != nil {
		log.Fatalf("Mining blocks: %v\n", err)
	}
	log.Println("Create the Bob<->Carol channel")

	_, err = createChannel(ctx, cli, bob, carolPubkey, "100000", btcd)
	if err != nil {
		log.Fatalf("Creating Bob<->Carol channel: %v", err)
	}

	log.Println("Recreate the Alice<->Bob channel")

	aliceBobChannelPoint, err = createChannel(ctx, cli, alice, bobPubkey, "1000000", btcd)
	if err != nil {
		log.Fatalf("Creating Alice<->Bob channel: %v", err)
	}

	log.Println("Generating invoice for payment to Carol")

	carolEncodedInvoice, err := generateInvoice(ctx, cli, carol, "10")
	if err != nil {
		log.Fatalf("Creating Carol invoice: %v", err)
	}

	log.Printf("Encoded payment request: %s\n", carolEncodedInvoice)

	log.Println("Alice paying invoice sending payment via Bob to Carol")

	err = payInvoice(ctx, cli, alice, carolEncodedInvoice)
	if err != nil {
		log.Fatalf("Sending Alice->Bob->Carol payment: %v", err)
	}

	log.Println("Checking payment received by Carol")
	carolChannelBalance, err := getChannelBalance(ctx, cli, carol)
	if err != nil {
		log.Fatalf("Checking Carol's balance: %v", err)
	}

	log.Println("Payment received by Carol")
	log.Printf("Carol's channel balance: %s\n", carolChannelBalance)

	log.Println("Cluster setup complete, ready to run tests")

	code := m.Run()

	// try to cleanup after run
	// can't defer this as os.Exit doesn't care for defer
	// cleanup(cli, ctx)
	os.Exit(code)
}

func getChannelBalance(ctx context.Context, cli *client.Client,
	initiator dockercontainer.ContainerCreateCreatedBody) (balance string, err error) {
	err = retry(func() error {
		var channelBalance struct {
			Balance string `json:"balance"`
		}
		cmd := []string{"lncli", "--network=simnet", "channelbalance"}
		err := execJSONReturningCommand(ctx, cli, initiator, cmd, &channelBalance)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", initiator.ID)
		}
		if channelBalance.Balance == "" {
			return errors.New("Payment not received")
		}
		if channelBalance.Balance == "0" {
			return errors.New("Payment not received")
		}
		balance = channelBalance.Balance
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Getting channel balance")
	}
	return balance, nil
}

func payInvoice(ctx context.Context, cli *client.Client,
	initiator dockercontainer.ContainerCreateCreatedBody, invoice string) error {
	err := retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "sendpayment", "--force", "--pay_req=" + invoice}
		var stderr bytes.Buffer
		_, stderr, err := execCommand(ctx, cli, initiator, cmd)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", initiator.ID)
		}
		if len(stderr.Bytes()) > 0 {
			return errors.New("Payment not sent")
		}
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return errors.Wrap(err, "Sending payment")
	}
	return nil
}

func generateInvoice(ctx context.Context, cli *client.Client,
	initiator dockercontainer.ContainerCreateCreatedBody, amount string) (encodedInvoice string, err error) {
	err = retry(func() error {
		var addInvoice struct {
			EncodedPayReq string `json:"payment_request"`
		}
		cmd := []string{"lncli", "--network=simnet", "addinvoice", "--amt=" + amount}
		err := execJSONReturningCommand(ctx, cli, initiator, cmd, &addInvoice)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", initiator.ID)
		}
		if addInvoice.EncodedPayReq == "" {
			return errors.New("Invoice not generated")
		}
		encodedInvoice = addInvoice.EncodedPayReq
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Creating invoice")
	}
	return encodedInvoice, nil
}

func checkPeerExists(ctx context.Context, cli *client.Client, initiator dockercontainer.ContainerCreateCreatedBody,
	remotePubkey string) (bool, error) {
	var listPeers struct {
		Peers []struct {
			Pubkey   string `json:"pub_key"`
			SyncType string `json:"sync_type"`
		} `json:"peers"`
	}
	err := retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "listpeers"}
		err := execJSONReturningCommand(ctx, cli, initiator, cmd, &listPeers)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on Alice %s", initiator.ID)
		}
		for _, peer := range listPeers.Peers {
			if peer.Pubkey == remotePubkey {
				// peer found, return from retry
				if peer.SyncType != "ACTIVE_SYNC" {
					return noRetryError{}
				}
				return nil
			}
		}
		return errors.New("peer not found")
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return false, errors.Wrap(err, "Checking nodes are peers")
	}
	return true, nil
}

func connectPeer(ctx context.Context, cli *client.Client, initiator dockercontainer.ContainerCreateCreatedBody,
	remotePubkey string, remoteIPAddress string) error {
	err := retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "connect", remotePubkey + "@" + remoteIPAddress}
		var stderr bytes.Buffer
		_, stderr, err := execCommand(ctx, cli, initiator, cmd)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", initiator.ID)
		}
		if len(stderr.Bytes()) > 0 {
			if !strings.Contains(string(stderr.Bytes()), "already connected to peer") {
				return errors.New("Peer didn't connect")
			}
		}

		// immediately check if the peer is connected as sometimes it seems to succeed and didn't
		peerConnected, err := checkPeerExists(ctx, cli, initiator, remotePubkey)
		if err != nil || !peerConnected {
			return errors.New("Peer didn't connect")
		}

		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return errors.Wrap(err, "Connecting peers")
	}
	return nil
}

func createChannel(ctx context.Context, cli *client.Client, initiator dockercontainer.ContainerCreateCreatedBody,
	remotePubkey string, amount string, btcd dockercontainer.ContainerCreateCreatedBody) (channelPoint string, err error) {

	var fundingTxId string
	err = retry(func() error {
		var openChannel struct {
			FundingTxId string `json:"funding_txid"`
		}
		cmd := []string{"lncli", "--network=simnet", "openchannel", "--node_key=" + remotePubkey, "--local_amt=" + amount}
		err := execJSONReturningCommand(ctx, cli, initiator, cmd, &openChannel)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", initiator.ID)
		}
		if openChannel.FundingTxId == "" {
			return errors.New("Channel not created")
		}
		fundingTxId = openChannel.FundingTxId
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Creating channel")
	}
	log.Printf("Funding transaction ID: %s\n", fundingTxId)

	log.Println("Include funding transaction in block thereby opening the channel")

	err = mineBlocks(ctx, cli, btcd, 30)
	if err != nil {
		return "", errors.Wrap(err, "Mining blocks")
	}

	log.Println("Blocks mined")
	log.Println("Checking channel is open")

	err = retry(func() error {
		var listChannels struct {
			Channels []struct {
				ChannelPoint string `json:"channel_point"`
			} `json:"channels"`
		}
		cmd := []string{"lncli", "--network=simnet", "listchannels"}
		err := execJSONReturningCommand(ctx, cli, initiator, cmd, &listChannels)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", initiator.ID)
		}
		if len(listChannels.Channels) == 0 {
			return errors.New("Channel not open")
		}
		if listChannels.Channels[0].ChannelPoint == "" {
			return errors.New("Channel not open")
		}
		channelPoint = listChannels.Channels[0].ChannelPoint
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Creating channel")
	}
	log.Printf("Channel point: %s\n", channelPoint)

	return channelPoint, nil

}

func getPubKey(ctx context.Context, cli *client.Client, container dockercontainer.ContainerCreateCreatedBody) (pubkey string, err error) {
	var getInfo struct {
		IdentityPubkey string `json:"identity_pubkey"`
	}
	err = retry(func() error {
		cmd := []string{"lncli", "--network=simnet", "getinfo"}
		err = execJSONReturningCommand(ctx, cli, container, cmd, &getInfo)
		if err != nil {
			return errors.Wrapf(err, "Running exec command on %s", container.ID)
		}
		if getInfo.IdentityPubkey == "" {
			return errors.New("Invalid Pubkey")
		}
		pubkey = getInfo.IdentityPubkey
		return nil
	}, defautDelayMS, defaultMaxDurationMS)
	if err != nil {
		return "", errors.Wrap(err, "Calling getinfo")
	}
	return pubkey, nil
}

func getOnchainBalance(ctx context.Context, cli *client.Client, container dockercontainer.ContainerCreateCreatedBody) (balance string, err error) {
	err = retry(func() error {
		var walletBalance struct {
			ConfirmedBalance string `json:"confirmed_balance"`
		}
		cmd := []string{"lncli", "--network=simnet", "walletbalance"}
		err := execJSONReturningCommand(ctx, cli, container, cmd, &walletBalance)
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
		err := execJSONReturningCommand(ctx, cli, btcd, cmd, &output)
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

func execJSONReturningCommand(ctx context.Context, cli *client.Client,
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
	if len(os.Getenv("DEBUG")) > 0 {
		log.Printf("%s\n", string(bufStdout.Bytes()))
		log.Printf("%s\n", string(bufStderr.Bytes()))
	}
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

type noRetryError struct{}

func (nre noRetryError) Error() string {
	return "Skip retries"
}

func retry(operation func() error, delayMilliseconds int, maxWaitMilliseconds int) error {
	totalWaited := 0
	for {
		if totalWaited > maxWaitMilliseconds {
			return errors.New("Exceeded maximum wait period")
		}
		err := operation()
		var noRetry noRetryError
		if errors.As(err, &noRetry) {
			return err
		}
		if err == nil {
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
	if len(os.Getenv("DEBUG")) > 0 {
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

func TestPayCarolFromAlice(t *testing.T) {
	log.Println("Generating invoice for payment to Carol")

	carolEncodedInvoice, err := generateInvoice(ctx, cli, carol, "10")
	if err != nil {
		log.Fatalf("Creating Carol invoice: %v", err)
	}

	log.Printf("Encoded payment request: %s\n", carolEncodedInvoice)

	log.Println("Alice paying invoice sending payment via Bob to Carol")

	err = payInvoice(ctx, cli, alice, carolEncodedInvoice)
	if err != nil {
		log.Fatalf("Sending Alice->Bob->Carol payment: %v", err)
	}

	log.Println("Checking payment received by Carol")
	carolChannelBalance, err := getChannelBalance(ctx, cli, carol)
	if err != nil {
		log.Fatalf("Checking Carol's balance: %v", err)
	}

	log.Println("Payment received by Carol")
	log.Printf("Carol's channel balance: %s\n", carolChannelBalance)
}
