import Table from "features/table/Table";
import { useGetOnChainTxQuery } from "./onChainApi";
import { Link, useNavigate } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Options20Regular as OptionsIcon,
  LinkEdit20Regular as NewOnChainAddressIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import Sidebar from "features/sidebar/Sidebar";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
  TableControlsButton,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import { SectionContainer } from "features/section/SectionContainer";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import FilterSection from "features/sidebar/sections/filter/FilterSection";
import SortSection from "features/sidebar/sections/sort/SortSection";
import Button, { buttonColor } from "components/buttons/Button";
import { NEW_ADDRESS } from "constants/routes";
import { useLocation } from "react-router";
import useTranslations from "services/i18n/useTranslations";
import { OnChainResponse } from "./types";
import DefaultCellRenderer from "../../table/DefaultCellRenderer";
import {
  AllOnChainColumns,
  DefaultOnChainView,
  OnChainFilterTemplate,
  OnChainSortTemplate,
  SortableOnChainColumns,
} from "./onChainDefaults";
import { useView } from "../../viewManagement/useView";
import { FilterInterface } from "../../sidebar/sections/filter/filter";
import { usePagination } from "../../../components/table/pagination/usePagination";

type sections = {
  filter: boolean;
  sort: boolean;
  columns: boolean;
};

function OnChainPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const [view, selectView, isViewsLoaded] = useView("onChain", AllOnChainColumns, 0, DefaultOnChainView);
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
      order: view.sortBy,
      filter: view.filters.length ? (view.filters.toJSON() as FilterInterface) : undefined,
    },
    { skip: !isViewsLoaded }
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

  // General logic for toggling the sidebar sections
  const initialSectionState: sections = {
    filter: false,
    sort: false,
    columns: false,
  };
  const [activeSidebarSections, setActiveSidebarSections] = useState(initialSectionState);

  const sidebarSectionHandler = (section: keyof sections) => {
    return () => {
      setActiveSidebarSections({
        ...activeSidebarSections,
        [section]: !activeSidebarSections[section],
      });
    };
  };

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
    <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SectionContainer
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={AllOnChainColumns} view={view} />
      </SectionContainer>
      <SectionContainer
        title={"Filter"}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        handleToggle={sidebarSectionHandler("filter")}
      >
        <FilterSection columns={AllOnChainColumns} view={view} defaultFilter={OnChainFilterTemplate} />
      </SectionContainer>
      <SectionContainer
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={SortableOnChainColumns} view={view} defaultSortBy={OnChainSortTemplate} />
      </SectionContainer>
    </Sidebar>
  );

  const breadcrumbs = [
    <span key="b1">Transactions</span>,
    <Link key="b2" to={"/transactions/on-chain"}>
      On-Chain Tx
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
        activeColumns={view.columns}
        isLoading={onChainTxResponse.isLoading || onChainTxResponse.isFetching || onChainTxResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default OnChainPage;
