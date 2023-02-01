// This program generates typescript files via go generate
package main

import (
	"os"

	"github.com/lncapital/torq/internal/views"
)

func main() {
	channelColumns := views.GetTableViewColumnDefinitionsForPage(views.PageChannels)
	err := os.WriteFile("../../web/src/features/channels/channelsColumns.ts", []byte(channelColumns), 0644)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	forwardsColumns := views.GetTableViewColumnDefinitionsForPage(views.PageForwards)
	err = os.WriteFile("../../web/src/features/forwards/forwardsColumns.ts", []byte(forwardsColumns), 0644)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	invoicesColumns := views.GetTableViewColumnDefinitionsForPage(views.PageInvoices)
	err = os.WriteFile("../../web/src/features/transact/Invoices/invoicesColumns.ts", []byte(invoicesColumns), 0644)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	paymentsColumns := views.GetTableViewColumnDefinitionsForPage(views.PagePayments)
	err = os.WriteFile("../../web/src/features/transact/Payments/paymentsColumns.ts", []byte(paymentsColumns), 0644)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	onChainTransactionsColumns := views.GetTableViewColumnDefinitionsForPage(views.PageOnChainTransactions)
	err = os.WriteFile("../../web/src/features/transact/OnChain/onChainColumns.ts", []byte(onChainTransactionsColumns), 0644)
	if err != nil {
		println(err.Error())
		panic(err)
	}

	tagsColumns := views.GetTableViewColumnDefinitionsForPage(views.PageTags)
	err = os.WriteFile("../../web/src/pages/tags/tagsPage/tagsColumns.ts", []byte(tagsColumns), 0644)
	if err != nil {
		println(err.Error())
		panic(err)
	}
}
