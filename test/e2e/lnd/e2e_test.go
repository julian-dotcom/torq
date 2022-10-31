package e2e

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/lncapital/torq/virtual_network"
	"github.com/playwright-community/playwright-go"
)

const torqPort = "4927"

const btcdName = "e2e-btcd"

const bobName = "e2e-bob"
const aliceName = "e2e-alice"
const carolName = "e2e-carol"

var ctx context.Context
var cli *client.Client
var torq dockercontainer.ContainerCreateCreatedBody
var btcd dockercontainer.ContainerCreateCreatedBody
var alice dockercontainer.ContainerCreateCreatedBody
var bob dockercontainer.ContainerCreateCreatedBody
var carol dockercontainer.ContainerCreateCreatedBody
var bobIPAddress string

func TestMain(m *testing.M) {

	if os.Getenv("E2E") == "" {
		log.Println("Skipping e2e tests as E2E environment variable not set")
		return
	}

	var err error
	pwRunOpts := &playwright.RunOptions{
		Browsers: []string{"chromium"},
	}
	err = playwright.Install(pwRunOpts)
	if err != nil {
		log.Fatalf("Installing playwright: %v\n", err)
	}

	ctx = context.Background()

	cli, err = client.NewClientWithOpts()
	if err != nil {
		log.Fatalf("Getting new docker client: %v\n", err)
	}

	de := virtual_network.DockerDevEnvironment{
		Client:           cli,
		NetworkName:      "e2e",
		SharedVolumeName: "e2e-shared",
	}

	// cleanup any old networks or containers that might have been left around from a failed run

	log.Println("Checking if any old containers or networks are present")

	// Add config for Torq database
	torqDbCont := de.AddContainer("e2e-torq-db",
		"timescale/timescaledb:latest-pg14",
		nil,
		[]string{"POSTGRES_PASSWORD=password"},
		nil,
		"")

	// Add config for Torq
	torqConf := de.AddContainer(
		"e2e-torq",
		"e2e/torq",
		nil,
		nil,
		[]string{
			"--db.host", torqDbCont.Name,
			"--db.password", "password",
			"--torq.password", "password",
			"--torq.port", torqPort,
			"start"},
		torqPort,
	)

	// Add config for btcd
	btcdConf := de.AddContainer(
		btcdName,
		"e2e/btcd",
		[]string{
			de.SharedVolumeName + ":/rpc",
			btcdName + ":/data",
		},
		[]string{"NETWORK=simnet"},
		nil,
		"",
	)

	// Add config for alice
	aliceConf := de.AddContainer(
		aliceName,
		"e2e/lnd",
		[]string{
			de.SharedVolumeName + ":/rpc",
			aliceName + ":/root/.lnd",
		},
		[]string{"NETWORK=simnet"},
		nil,
		"",
	)

	// Add config for bob
	bobConf := de.AddContainer(
		bobName,
		"e2e/lnd",
		[]string{
			de.SharedVolumeName + ":/rpc",
			bobName + ":/root/.lnd",
		},
		[]string{"NETWORK=simnet"},
		nil,
		"10011",
	)

	// Add config for carol
	carolConf := de.AddContainer(
		carolName,
		"e2e/lnd",
		[]string{
			de.SharedVolumeName + ":/rpc",
			carolName + ":/root/.lnd",
		},
		[]string{"NETWORK=simnet"},
		nil,
		"",
	)

	// Clean up old containers and network before initiating new ones.
	if err = de.CleanupContainers(ctx); err != nil {
		log.Fatalf("Couldn't cleanup containers: %v", err)
	}
	if err = de.CleanupDefaultVolumes(ctx); err != nil {
		log.Fatalf("Couldn't cleanup volumes %v", err)
	}
	if err = de.FindAndRemoveNetwork(ctx, de.NetworkName); err != nil {
		log.Fatalf("Couldn't cleanup network: %v", err)
	}

	// Create the shared network
	networkingConfig, err := de.CreateNetwork(ctx)
	if err != nil {
		log.Fatal(err)
	}
	de.NetworkingConfig = networkingConfig

	log.Println("Creating e2e network")

	// Start the database
	err = de.InitContainer(ctx, torqDbCont)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Building Torq image")
	rootPath, _ := filepath.Abs("../../../")
	// path to Dockerfile in root of project
	err = de.BuildImage(ctx, rootPath, "e2e/torq")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Building btcd image from dockerfile")
	btcdPath, _ := filepath.Abs("../../../virtual_network/docker/btcd/")
	err = de.BuildImage(ctx, btcdPath, "e2e/btcd")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Building lnd image from dockerfile")
	lndPath, _ := filepath.Abs("../../../virtual_network/docker/lnd/")
	err = de.BuildImage(ctx, lndPath, "e2e/lnd")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting btcd")
	err = de.InitContainer(ctx, btcdConf)
	if err != nil {
		log.Fatal(err)
	}
	btcd = btcdConf.Instance

	log.Println("Starting Alice")
	err = de.InitContainer(ctx, aliceConf)
	if err != nil {
		log.Fatal(err)
	}
	alice = aliceConf.Instance

	// Example looking at container logs
	//out, err := cli.ContainerLogs(ctx, btcd.ID, types.ContainerLogsOptions{ShowStdout: true})
	//if err != nil {
	//	panic(err)
	//}
	//stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	log.Println("Creating new mining address on Alice")

	aliceAddress, err := virtual_network.GetNewAddress(ctx, cli, alice)
	if err != nil {
		log.Fatalf("Getting alice mining address: %v", err)
	}
	log.Println("Alice receive address created")
	log.Println(aliceAddress)

	log.Println("Shutting Alice down before btcd restart")
	if err = de.FindAndRemoveContainer(ctx, aliceConf.Name); err != nil {
		log.Fatalf("Unable to find and remove alice container: %v", err)
	}

	log.Println("Recreating btcd container with Alice's mining address")
	if err = de.FindAndRemoveContainer(ctx, btcdConf.Name); err != nil {
		log.Fatalf("Unable to find and remove btcd container: %v", err)
	}

	log.Println("Starting new btcd container")
	// Update the container config with the minind addres instead of adding a new one
	btcdConf.Env = []string{"NETWORK=simnet", "MINING_ADDRESS=" + aliceAddress}
	err = de.InitContainer(ctx, btcdConf)
	if err != nil {
		log.Fatal(err)
	}
	btcd = btcdConf.Instance

	log.Println("Generate 400 blocks (we need at least \"100 >=\" blocks because of coinbase block maturity and \"300 ~=\" in order to activate segwit)")

	err = virtual_network.MineBlocks(ctx, cli, btcd, 400)
	if err != nil {
		log.Fatalf("Mining blocks: %v", err)
	}

	log.Println("Blocks mined")

	log.Println("Recreating Alice now that btcd is back online")

	const aliceVolumeName = "e2e-alice"
	aliceConf.Binds = []string{
		de.SharedVolumeName + ":/rpc",
		aliceVolumeName + ":/root/.lnd",
	}

	// Initiation new Alice (based on the same configuration
	err = de.InitContainer(ctx, aliceConf)
	if err != nil {
		log.Fatal(err)
	}
	alice = aliceConf.Instance

	log.Println("Checking that segwit is active")
	err = virtual_network.SegWitActive(ctx, cli, btcd)
	if err != nil {
		log.Fatalf("btcd checking segwit is active: %v", err)
	}

	log.Println("Segwit is active")
	log.Println("Checking Alice's balance")

	// Skipping balance check for now and assuming it has worked for speed
	// aliceBalance, err := getOnchainBalance(ctx, cli, alice)
	// if err != nil {
	// 	log.Fatalf("Getting Alice's balance: %v", err)
	// }

	// log.Printf("Alice's onchain balance is: %s\n", aliceBalance)

	// Starting carol
	log.Println("Starting Carol")
	err = de.InitContainer(ctx, carolConf)
	if err != nil {
		log.Fatal(err)
	}
	carol = carolConf.Instance

	// start Bob and Carol AFTER btcd has restarted
	log.Println("Starting Bob")
	err = de.InitContainer(ctx, bobConf)
	if err != nil {
		log.Fatal(err)
	}
	bob = bobConf.Instance

	log.Println("Get Bob's pubkey")
	bobInspection, err := cli.ContainerInspect(ctx, bob.ID)
	if err != nil {
		log.Fatalf("Getting Bob's IP Address: %v", err)
	}
	bobIPAddress = bobInspection.NetworkSettings.Networks["e2e"].IPAddress
	log.Println("Bob's IP address is:")
	log.Println(bobIPAddress)

	bobPubkey, err := virtual_network.GetPubKey(ctx, cli, bob)
	if err != nil {
		log.Fatalf("Getting Bob's pubkey: %v", err)
	}
	log.Printf("Bob's pubkey is: %s\n", bobPubkey)

	log.Println("Connecting Bob to Alice")

	err = virtual_network.ConnectPeer(ctx, cli, alice, bobPubkey, bobIPAddress)
	if err != nil {
		log.Fatalf("Connecting Bob to Alice: %v", err)
	}

	log.Println("Verifing Bob is a peer of Alice")

	bobPeerExists, err := virtual_network.CheckPeerExists(ctx, cli, alice, bobPubkey)
	if err != nil || !bobPeerExists {
		log.Fatalf("Checking that Bob is a peer of Alice: %v", err)
	}

	log.Println("Bob confirmed as peer of Alice")

	log.Println("Getting Alice's pubkey")
	alicePubkey, err := virtual_network.GetPubKey(ctx, cli, alice)
	if err != nil {
		log.Fatalf("Getting Alice's pubkey: %v", err)
	}

	log.Printf("Alice's pubkey is: %s\n", alicePubkey)

	log.Println("Verifing Alice is a peer of Bob")

	alicePeerExists, err := virtual_network.CheckPeerExists(ctx, cli, bob, alicePubkey)
	if err != nil || !alicePeerExists {
		log.Fatalf("Checking that Alice is a peer of Bob: %v", err)
	}
	log.Println("Alice confirmed as peer of Bob")

	log.Println("Create the Alice<->Bob channel")

	aliceBobChannelPoint, err := virtual_network.CreateChannel(ctx, cli, alice, bobPubkey, "12000000", btcd)
	if err != nil {
		log.Fatalf("Creating Alice<->Bob channel: %v", err)
	}

	log.Println("Generating invoice for payment to Bob")

	bobEncodedInvoice, err := virtual_network.GenerateInvoice(ctx, cli, bob, "4100000")
	if err != nil {
		log.Fatalf("Creating Bob invoice: %v", err)
	}

	log.Printf("Encoded payment request: %s\n", bobEncodedInvoice)

	log.Println("Alice paying invoice sending payment to Bob")

	err = virtual_network.PayInvoice(ctx, cli, alice, bobEncodedInvoice)
	if err != nil {
		log.Fatalf("Sending Alice->Bob payment: %v", err)
	}

	log.Println("Checking payment received by Bob")
	bobChannelBalance, err := virtual_network.GetChannelBalance(ctx, cli, bob)
	if err != nil {
		log.Fatalf("Checking Bob's balance: %v", err)
	}

	log.Println("Payment received by Bob")
	log.Printf("Bob's channel balance: %s\n", bobChannelBalance)

	log.Println("Close Alice<->Bob channel to gain on chain funds for Bob")

	aliceBobClosingTxId, err := virtual_network.CloseChannel(ctx, cli, alice, aliceBobChannelPoint)
	if err != nil {
		log.Fatalf("Closing Alice<->Bob channel: %v", err)
	}

	log.Printf("Alice<->Bob channel closing transaction id: %s\n", aliceBobClosingTxId)

	log.Println("Mining some blocks to confirm closing transaction")

	err = virtual_network.MineBlocks(ctx, cli, btcd, 3)
	if err != nil {
		log.Fatalf("Mining blocks: %v", err)
	}

	bobOnChainBalance, err := virtual_network.GetOnchainBalance(ctx, cli, bob)
	if err != nil {
		log.Fatalf("Getting Bob's balance: %v", err)
	}
	log.Printf("Bob's onchain balance: %s\n", bobOnChainBalance)

	// Starting torq here means that the database should be ready and Torq should be up before test needs it
	// Better solution would be to check that the DB is ready and that Torq is ready
	log.Println("Starting Torq")

	err = de.InitContainer(ctx, torqConf)
	if err != nil {
		log.Fatal(err)
	}
	torq = torqConf.Instance

	log.Println("Getting Carol's pubkey")
	carolPubkey, err := virtual_network.GetPubKey(ctx, cli, carol)
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

	err = virtual_network.ConnectPeer(ctx, cli, bob, carolPubkey, carolIPAddress)
	if err != nil {
		log.Fatalf("Connecting Carol to Bob: %v", err)
	}

	log.Println("Verifing Carol is a peer of Bob")

	carolPeerExists, err := virtual_network.CheckPeerExists(ctx, cli, bob, carolPubkey)
	if err != nil || !carolPeerExists {
		log.Fatalf("Checking that Carol is a peer of Bob: %v", err)
	}

	log.Println("Carol confirmed as peer of Bob")

	log.Println("Verifing Bob is a peer of Carol")
	carolBobPeerExists, err := virtual_network.CheckPeerExists(ctx, cli, carol, bobPubkey)
	if err != nil || !carolBobPeerExists {
		log.Fatalf("Checking that Bob is a peer of Carol: %v", err)
	}
	log.Println("Bob confirmed as peer of Carol")

	err = virtual_network.MineBlocks(ctx, cli, btcd, 30)
	if err != nil {
		log.Fatalf("Mining blocks: %v\n", err)
	}
	log.Println("Create the Bob<->Carol channel")

	_, err = virtual_network.CreateChannel(ctx, cli, bob, carolPubkey, "100000", btcd)
	if err != nil {
		log.Fatalf("Creating Bob<->Carol channel: %v", err)
	}

	log.Println("Recreate the Alice<->Bob channel")

	_, err = virtual_network.CreateChannel(ctx, cli, alice, bobPubkey, "1000000", btcd)
	if err != nil {
		log.Fatalf("Creating Alice<->Bob channel: %v", err)
	}

	log.Println("Generating invoice for payment to Carol")

	carolEncodedInvoice, err := virtual_network.GenerateInvoice(ctx, cli, carol, "10")
	if err != nil {
		log.Fatalf("Creating Carol invoice: %v", err)
	}

	log.Printf("Encoded payment request: %s\n", carolEncodedInvoice)

	log.Println("Alice paying invoice sending payment via Bob to Carol")

	err = virtual_network.PayInvoice(ctx, cli, alice, carolEncodedInvoice)
	if err != nil {
		log.Fatalf("Sending Alice->Bob->Carol payment: %v", err)
	}

	log.Println("Checking payment received by Carol")
	carolChannelBalance, err := virtual_network.GetChannelBalance(ctx, cli, carol)
	if err != nil {
		log.Fatalf("Checking Carol's balance: %v", err)
	}

	log.Println("Payment received by Carol")
	log.Printf("Carol's channel balance: %s\n", carolChannelBalance)

	log.Println("Cluster setup complete, ready to run tests")

	code := m.Run()

	// try to cleanup after run
	// can't defer this as os.Exit doesn't care for defer
	if code == 0 {
		if err = de.CleanupContainers(ctx); err != nil {
			log.Fatalf("Unable to cleanup containers: %v", err)
		}
	}

	os.Exit(code)
}

func TestPayCarolFromAlice(t *testing.T) {
	log.Println("Generating invoice for payment to Carol")

	carolEncodedInvoice, err := virtual_network.GenerateInvoice(ctx, cli, carol, "10")
	if err != nil {
		log.Fatalf("Creating Carol invoice: %v", err)
	}

	log.Printf("Encoded payment request: %s\n", carolEncodedInvoice)

	log.Println("Alice paying invoice sending payment via Bob to Carol")

	err = virtual_network.PayInvoice(ctx, cli, alice, carolEncodedInvoice)
	if err != nil {
		log.Fatalf("Sending Alice->Bob->Carol payment: %v", err)
	}

	log.Println("Checking payment received by Carol")
	carolChannelBalance, err := virtual_network.GetChannelBalance(ctx, cli, carol)
	if err != nil {
		log.Fatalf("Checking Carol's balance: %v", err)
	}

	log.Println("Payment received by Carol")
	log.Printf("Carol's channel balance: %s\n", carolChannelBalance)
}

func TestPlaywrightVideo(t *testing.T) {

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not launch playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("could not launch Chromium: %v", err)
	}
	page, err := browser.NewPage(playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.BrowserNewContextOptionsRecordVideo{
			Dir: playwright.String("e2e_videos/"),
			Size: &playwright.BrowserNewContextOptionsRecordVideoSize{
				Width:  playwright.Int(1920),
				Height: playwright.Int(1080),
			},
		},
	})

	defer func() {
		if err := page.Close(); err != nil {
			t.Fatalf("failed to close page: %v", err)
		}
		path, err := page.Video().Path()
		if err != nil {
			t.Fatalf("failed to get video path: %v", err)
		}
		fmt.Printf("Saved to %s\n", path)
		if err = browser.Close(); err != nil {
			t.Fatalf("could not close browser: %v", err)
		}
		if err = pw.Stop(); err != nil {
			t.Fatalf("could not stop Playwright: %v", err)
		}
	}()

	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}
	gotoPage := func(url string) {
		fmt.Printf("Visiting %s\n", url)
		if _, err = page.Goto(url); err != nil {
			t.Fatalf("could not goto: %v", err)
		}
		fmt.Printf("Visited %s\n", url)
	}
	gotoPage("http://localhost:" + torqPort)

	// page redirects to login
	// _, err = page.WaitForNavigation(playwright.PageWaitForNavigationOptions{URL: "http://localhost:3000/login"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	click := func(target string) {
		if err = page.Click(target); err != nil {
			t.Fatalf("Couldn't find '%v' on page: %v", target, err)
		}
	}

	fill := func(target string, value string) {
		if err = page.Fill(target, value); err != nil {
			t.Fatalf("Couldn't fill '%v' with value '%v': %v", target, value, err)
		}
	}

	fill(".login-form .password-field", "password")

	click(".login-form .submit-button")

	_, err = page.Locator("text=Forwarding fees")
	if err != nil {
		t.Fatal(err)
	}

	click("text=Settings")
	ws, err := page.IsVisible("text=Week starts on")
	if err != nil {
		t.Fatal(err)
	}
	if !ws {
		t.Fatalf("Week starts on not found\n")
	}

	click("text=Add Node")

	fill("#address input[type=text]", bobIPAddress+":10011")

	tlsFileReader, _, err := cli.CopyFromContainer(ctx, bobName, "/root/.lnd/tls.cert")
	if err != nil {
		t.Fatalf("Copying tls file: %v\n", err)
	}
	// file comes out as a tar, untar it
	tlsTar := tar.NewReader(tlsFileReader)
	// hdr gives you the header of the tar file
	_, err = tlsTar.Next()
	if err == io.EOF || err != nil {
		// EOF == end of tar archive
		t.Fatalf("Reading tls tar header: %v\n", err)
	}
	tlsBuf := new(bytes.Buffer)
	_, err = tlsBuf.ReadFrom(tlsTar)
	if err != nil {
		t.Fatalf("Reading tls tar: %v\n", err)
	}

	pTlsFile := playwright.InputFile{Name: "tls.cert", Buffer: tlsBuf.Bytes()}
	if err = page.SetInputFiles("#tls input[type=file]", []playwright.InputFile{pTlsFile}); err != nil {
		t.Fatalf("Couldn't set tls file on settings page: %v", err)
	}

	macaroonFileReader, _, err := cli.CopyFromContainer(ctx, bobName, "/root/.lnd/data/chain/bitcoin/simnet/admin."+
		"macaroon")
	if err != nil {
		t.Fatalf("Copying macaroon file: %v\n", err)
	}
	// file comes out as a tar, untar it
	macaroonTar := tar.NewReader(macaroonFileReader)
	// hdr gives you the header of the tar file
	_, err = macaroonTar.Next()
	if err == io.EOF || err != nil {
		// EOF == end of tar archive
		t.Fatalf("Reading macaroon tar header: %v\n", err)
	}
	macaroonBuf := new(bytes.Buffer)
	_, err = macaroonBuf.ReadFrom(macaroonTar)
	if err != nil {
		t.Fatalf("Reading macaroon tar: %v\n", err)
	}

	pMacaroonFile := playwright.InputFile{Name: "readonly.macaroon", Buffer: macaroonBuf.Bytes()}
	if err = page.SetInputFiles("#macaroon input[type=file]", []playwright.InputFile{pMacaroonFile}); err != nil {
		t.Fatalf("Couldn't set macaroon file on settings page: %v", err)
	}

	click("id=save-node")

	click("text=Summary")

	click("text=Forwards")

	click("text=Default View")

	click("text=Channels")

	click("text=Default View")

	click("text=Transactions")

	click("id=collapse-navigation")

	click("text=Invoices")

	click("text=On-Chain")

	click("id=tableControlsButton")

	click("text=Filter")

	click("text=Add filter")

	click("text=Sort")
}
