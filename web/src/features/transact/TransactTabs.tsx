import TabButton from "components/buttons/TabButton";
import { TableControlsTabsGroup } from "features/templates/tablePageTemplate/TablePageTemplate";
import ViewsPopover from "features/viewManagement/ViewsPopover";
import { ViewInterface } from "features/table/Table";
import { RootState } from "store/store";
import { ActionCreatorWithPayload } from "@reduxjs/toolkit";

type TransactTabs = {
  page: string;
  selectViews: (state: RootState) => ViewInterface[];
  // selectViews: ViewInterface[],
  updateViews: ActionCreatorWithPayload<
    {
      views: ViewInterface[];
      index: number;
    },
    string
  >;
  updateSelectedView: ActionCreatorWithPayload<
    {
      index: number;
    },
    string
  >;
  selectedViewIndex: (state: RootState) => number;
  DefaultView: ViewInterface;
  updateViewsOrder: ActionCreatorWithPayload<
    {
      views: ViewInterface[];
      index: number;
    },
    string
  >;
};

function TransactTabs({
  page,
  selectViews,
  updateViews,
  updateSelectedView,
  selectedViewIndex,
  updateViewsOrder,
  DefaultView,
}: TransactTabs) {
  return (
    <TableControlsTabsGroup>
      {/*<TabButton to={"/transactions/all"} title={"All"} />*/}
      <TabButton to={"/transactions/payments"} title={"Payments"} />
      <TabButton to={"/transactions/invoices"} title={"Invoices"} />
      <TabButton to={"/transactions/onchain"} title={"On-Chain"} />
      {
        <ViewsPopover
          page={page}
          selectViews={selectViews}
          updateSelectedView={updateSelectedView}
          selectedViewIndex={selectedViewIndex}
          updateViews={updateViews}
          updateViewsOrder={updateViewsOrder}
          DefaultView={DefaultView}
        />
      }
    </TableControlsTabsGroup>
  );
}

export default TransactTabs;
