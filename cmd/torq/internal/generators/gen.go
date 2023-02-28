package main

import (
	"os"

	"github.com/lncapital/torq/internal/views"
)

// This program generates typescript files via go generate
// make generate-ts <- run make cmd at root
func main() {
	channelColumns := views.GetTableViewColumnDefinitionsForPage(views.PageChannels)
	println(os.Getwd())
	err := os.WriteFile("web/src/features/channels/channelsColumns.generated.ts", []byte(channelColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	forwardsColumns := views.GetTableViewColumnDefinitionsForPage(views.PageForwards)
	err = os.WriteFile("web/src/features/forwards/forwardsColumns.generated.ts", []byte(forwardsColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	invoicesColumns := views.GetTableViewColumnDefinitionsForPage(views.PageInvoices)
	err = os.WriteFile("web/src/features/transact/Invoices/invoicesColumns.generated.ts", []byte(invoicesColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	paymentsColumns := views.GetTableViewColumnDefinitionsForPage(views.PagePayments)
	err = os.WriteFile("web/src/features/transact/Payments/paymentsColumns.generated.ts", []byte(paymentsColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	onChainTransactionsColumns := views.GetTableViewColumnDefinitionsForPage(views.PageOnChainTransactions)
	err = os.WriteFile("web/src/features/transact/OnChain/onChainColumns.generated.ts", []byte(onChainTransactionsColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	tagsColumns := views.GetTableViewColumnDefinitionsForPage(views.PageTags)
	err = os.WriteFile("web/src/pages/tags/tagsPage/tagsColumns.generated.ts", []byte(tagsColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	channelClosedColumns := views.GetTableViewColumnDefinitionsForPage(views.PageChannelsClosed)
	err = os.WriteFile("web/src/features/channelsClosed/channelsClosedColumns.generated.ts", []byte(channelClosedColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	channelPendingColumns := views.GetTableViewColumnDefinitionsForPage(views.PageChannelsPending)
	err = os.WriteFile("web/src/features/channelsPending/channelsPendingColumns.generated.ts", []byte(channelPendingColumns), 0600)
	if err != nil {
		println(err.Error())
		panic(err)
	}
}
