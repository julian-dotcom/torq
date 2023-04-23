import Table from "features/table/Table";
import { useGetOnChainTxQuery } from "./onChainApi";
import { Link, useNavigate } from "react-router-dom";
import {
  Options20Regular as OptionsIcon,
  Add20Regular as NewOnChainAddressIcon,
  ArrowSync20Regular as RefreshIcon,
} from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import Button, { ColorVariant } from "components/buttons/Button";
import { NEW_ADDRESS } from "constants/routes";
import { useLocation } from "react-router";
import useTranslations from "services/i18n/useTranslations";
import { OnChainTx } from "./types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import {
  DefaultOnChainView,
  FilterableOnChainColumns,
  OnChainFilterTemplate,
  OnChainSortTemplate,
  SortableOnChainColumns,
} from "features/transact/OnChain/onChainDefaults";
import { AllOnChainTransactionsColumns } from "features/transact/OnChain/onChainColumns.generated";
import { usePagination } from "components/table/pagination/usePagination";
import { useGetTableViewsQuery, useUpdateTableViewMutation } from "features/viewManagement/viewsApiSlice";
import { useAppSelector } from "store/hooks";
import { selectOnChainView, selectViews } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectActiveNetwork } from "features/network/networkSlice";
import { TableResponses, ViewResponse } from "features/viewManagement/types";
import { userEvents } from "utils/userEvents";

function useMaximums(data: Array<OnChainTx>): OnChainTx | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: OnChainTx, current: OnChainTx) => {
    return {
      ...prev,
      alias: "Max",
      amount: Math.max(prev.amount, current.amount),
      totalFees: Math.max(prev.totalFees, current.totalFees),
      txHash: Math.max(prev.txHash, current.txHash),
    };
  });
}

function OnChainPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();
  const { track } = userEvents();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectOnChainView);
  const channelViews = useAppSelector(selectViews)("channelsClosed");
  const [updateTableView] = useUpdateTableViewMutation();

  const [getPagination, limit, offset] = usePagination("onChain");
  const activeNetwork = useAppSelector(selectActiveNetwork);

  const onChainTxResponse = useGetOnChainTxQuery(
    {
      limit: limit,
      offset: offset,
      order: viewResponse.view.sortBy,
      filter: viewResponse.view.filters ? viewResponse.view.filters : undefined,
      network: activeNetwork,
    },
    { skip: !isSuccess, pollingInterval: 10000 }
  );

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
    track("Toggle Table Sidebar", { page: "OnChain" });
  };

  const maxRow = useMaximums(onChainTxResponse.data?.data || []);

  function handleNameChange(name: string) {
    const view = channelViews.views[selectedViewIndex] as ViewResponse<TableResponses>;
    if (view.id) {
      updateTableView({
        id: view.id,
        view: { ...view.view, title: name },
      });
    }
  }

  const tableControls = (
    <TableControlSection intercomTarget={"table-page-controls"}>
      <TableControlsButtonGroup intercomTarget={"table-page-controls-left"}>
        <Button
          intercomTarget="new-address"
          buttonColor={ColorVariant.success}
          icon={<NewOnChainAddressIcon />}
          hideMobileText={true}
          onClick={() => {
            navigate(NEW_ADDRESS, { state: { background: location } });
            track("Navigate to New OnChain Address");
          }}
        >
          {t.newAddress}
        </Button>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup intercomTarget={"table-page-controls-right"}>
        <Button
          intercomTarget="refresh-table"
          buttonColor={ColorVariant.primary}
          icon={<RefreshIcon />}
          onClick={() => {
            track("Refresh Table", { page: "OnChain" });
            onChainTxResponse.refetch();
          }}
        />
        <Button
          intercomTarget="table-settings"
          onClick={() => {
            setSidebarExpanded(!sidebarExpanded);
            track("Toggle Table Sidebar", { page: "OnChain" });
          }}
          hideMobileText={true}
          icon={<OptionsIcon />}
          id={"tableControlsButton"}
        >
          {t.Options}
        </Button>
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const sidebar = (
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      selectedViewIndex={selectedViewIndex}
      allColumns={AllOnChainTransactionsColumns}
      defaultView={DefaultOnChainView}
      filterableColumns={FilterableOnChainColumns}
      filterTemplate={OnChainFilterTemplate}
      sortableColumns={SortableOnChainColumns}
      sortByTemplate={OnChainSortTemplate}
    />
  );

  const breadcrumbs = [
    <span key="b1">{t.transactions}</span>,
    <Link key="b2" to={"/transactions/on-chain"}>
      {t.onChainTx}
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={viewResponse.view.title}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={getPagination(onChainTxResponse?.data?.pagination?.total || 0)}
      onNameChange={handleNameChange}
      isDraft={viewResponse.id === undefined}
    >
      <Table
        intercomTarget={"on-chain-table"}
        cellRenderer={DefaultCellRenderer}
        data={onChainTxResponse?.data?.data || []}
        activeColumns={viewResponse.view.columns}
        isLoading={onChainTxResponse.isLoading || onChainTxResponse.isFetching || onChainTxResponse.isUninitialized}
        maxRow={maxRow}
      />
    </TablePageTemplate>
  );
}

export default OnChainPage;
