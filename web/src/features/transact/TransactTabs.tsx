import React from "react";
import { TableControlsTab, TableControlsTabsGroup } from "../tablePageTemplate/TablePageTemplate";

function TransactTabs() {
  return (
    <TableControlsTabsGroup>
      <TableControlsTab to={"/transactions/all"} title={"All"} />
      <TableControlsTab to={"/transactions/payments"} title={"Payments"} />
      <TableControlsTab to={"/transactions/invoices"} title={"Invoices"} />
      <TableControlsTab to={"/transactions/onchain"} title={"On-Chain"} />
    </TableControlsTabsGroup>
  );
}

export default TransactTabs;
