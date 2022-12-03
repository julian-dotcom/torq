import Table from "features/table/Table";
import { useGetOnChainTxQuery } from "./onChainApi";
import { Link, useNavigate } from "react-router-dom";
import {
  Options20Regular as OptionsIcon,
  LinkEdit20Regular as NewOnChainAddressIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
  TableControlsButton,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import Button, { buttonColor } from "components/buttons/Button";
import { NEW_ADDRESS } from "constants/routes";
import { useLocation } from "react-router";
import useTranslations from "services/i18n/useTranslations";
import { OnChainResponse } from "./types";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import {
  AllOnChainColumns,
  DefaultOnChainView,
  OnChainFilterTemplate,
  OnChainSortTemplate,
  SortableOnChainColumns,
} from "./onChainDefaults";
import { FilterInterface } from "features/sidebar/sections/filter/filter";
import { usePagination } from "components/table/pagination/usePagination";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { useAppSelector } from "store/hooks";
import { selectOnChainView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";

function OnChainPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const viewResponse = useAppSelector(selectOnChainView);
  const [getPagination, limit, offset] = usePagination("onChain");

  const onChainTxResponse = useGetOnChainTxQuery<{
    data: OnChainResponse;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>(
    {
      limit: limit,
      offset: offset,
      order: viewResponse.view.sortBy,
      filter: viewResponse.view.filters.length ? (viewResponse.view.filters.toJSON() as FilterInterface) : undefined,
    },
    { skip: !isSuccess }
  );

  // let data: any = [];
  //
  // if (onchainResponse?.data?.data) {
  //   data = onchainResponse?.data?.data.map((invoice: any) => {
  //     const invoice_state = statusTypes[invoice.invoice_state];
  //
  //     return {
  //       ...invoice,
  //       invoice_state,
  //     };
  //   });
  // }

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  const closeSidebarHandler = () => {
    return () => {
      setSidebarExpanded(false);
    };
  };

  // const saveView = () => {
  //   const viewMod = { ...currentView };
  //   viewMod.saved = true;
  //   if (currentView.id === undefined || null) {
  //     createTableView({ view: viewMod, index: currentViewIndex, page: "onChain" });
  //     return;
  //   }
  //   updateTableView(viewMod);
  // };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <Button
          buttonColor={buttonColor.green}
          text={t.newAddress}
          icon={<NewOnChainAddressIcon />}
          className={"collapse-tablet"}
          onClick={() => {
            navigate(NEW_ADDRESS, { state: { background: location } });
          }}
        />
        <TableControlsButton
          onClickHandler={() => setSidebarExpanded(!sidebarExpanded)}
          icon={OptionsIcon}
          id={"tableControlsButton"}
        />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const sidebar = (
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      allColumns={AllOnChainColumns}
      defaultView={DefaultOnChainView}
      filterableColumns={AllOnChainColumns}
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
      title={"OnChain"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      pagination={getPagination(onChainTxResponse?.data?.pagination?.total || 0)}
    >
      <Table
        cellRenderer={DefaultCellRenderer}
        data={onChainTxResponse?.data?.data || []}
        activeColumns={viewResponse.view.columns}
        isLoading={onChainTxResponse.isLoading || onChainTxResponse.isFetching || onChainTxResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default OnChainPage;
