// This program generates typescript files via go generate
package main

import (
	"os"

	"github.com/lncapital/torq/internal/views"
)

func main() {
	channelColumns := views.GetTableViewColumnDefinitionsForPage(views.PageChannel)
	err := os.WriteFile("../../web/src/features/channels/channelsColumns.ts", []byte(channelColumns), 0644)
	if err != nil {
		println(err.Error())
		panic(err)
	}
}
