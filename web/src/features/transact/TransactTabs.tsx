import React from "react";
import { TableControlsTab, TableControlsTabsGroup } from "../tablePage/TablePageTemplate";

function TransactTabs() {
  return (
    <TableControlsTabsGroup>
      <TableControlsTab to={"/Transactions/all"} title={"All"} />
      <TableControlsTab to={"/Transactions/payments"} title={"Payments"} />
      <TableControlsTab to={"/Transactions/invoices"} title={"Invoices"} />
      <TableControlsTab to={"/Transactions/onchain"} title={"On-Chain"} />
    </TableControlsTabsGroup>
  );
}

export default TransactTabs;
