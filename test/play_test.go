package e2e

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func TestPlaywrightVideo(t *testing.T) {

	if os.Getenv("E2E") == "" {
		log.Println("Skipping e2e tests as E2E environment variable not set")
		return
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not launch playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch Chromium: %v", err)
	}
	page, err := browser.NewPage(playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.BrowserNewContextOptionsRecordVideo{
			Dir: playwright.String("e2e_videos/"),
		},
	})
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	gotoPage := func(url string) {
		fmt.Printf("Visiting %s\n", url)
		if _, err = page.Goto(url); err != nil {
			log.Fatalf("could not goto: %v", err)
		}
		fmt.Printf("Visited %s\n", url)
	}
	gotoPage("http://localhost:3000")

	// page redirects to login
	// _, err = page.WaitForNavigation(playwright.PageWaitForNavigationOptions{URL: "http://localhost:3000/login"})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	page.Fill(".login-form .password-field", "password")

	page.Click(".login-form .submit-button")

	ff, err := page.IsVisible("text=Forwarding fees")
	if err != nil {
		log.Fatal(err)
	}
	if !ff {
		log.Fatalln("Forwarding fees not found")
	}

	page.Click("text=Settings")
	ws, err := page.IsVisible("text=Week starts on")
	if err != nil {
		log.Fatal(err)
	}
	if !ws {
		log.Fatalln("Week starts on not found")
	}

	page.Fill("#address input[type=text]", "100.100.100.100:10009")

	tlsFileBuf, err := os.ReadFile("tls.cert")
	if err != nil {
		log.Fatal(err)
	}
	tlsFile := playwright.InputFile{Name: "tls4.cert", Buffer: tlsFileBuf}
	page.SetInputFiles("#tls input[type=file]", []playwright.InputFile{tlsFile})

	macaroonFileBuf, err := os.ReadFile("readonly.macaroon")
	if err != nil {
		log.Fatal(err)
	}
	macaroonFile := playwright.InputFile{Name: "readonly4.macaroon", Buffer: macaroonFileBuf}
	page.SetInputFiles("#macaroon input[type=file]", []playwright.InputFile{macaroonFile})

	page.Click("text=Save node details")

	time.Sleep(5 * time.Second)

	// log.Println(page.Content())

	if err := page.Close(); err != nil {
		log.Fatalf("failed to close page: %v", err)
	}
	path, err := page.Video().Path()
	if err != nil {
		log.Fatalf("failed to get video path: %v", err)
	}
	fmt.Printf("Saved to %s\n", path)
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
