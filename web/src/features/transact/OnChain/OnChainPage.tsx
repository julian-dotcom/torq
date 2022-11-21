import Table from "features/table/Table";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { useGetOnChainTxQuery } from "apiSlice";
import { Link, useNavigate } from "react-router-dom";
import {
  // Filter20Regular as FilterIcon,
  // ArrowSortDownLines20Regular as SortIcon,
  // ColumnTriple20Regular as ColumnsIcon,
  // Options20Regular as OptionsIcon,
  LinkEdit20Regular as NewOnChainAddressIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
// import Sidebar from "features/sidebar/Sidebar";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import Pagination from "components/table/pagination/Pagination";
import useLocalStorage from "features/helpers/useLocalStorage";
// import SortSection, { OrderBy } from "features/sidebar/sections/sort/SortSection";
// import FilterSection from "features/sidebar/sections/filter/FilterSection";
// import { Clause, FilterInterface } from "features/sidebar/sections/filter/filter";
import { useAppDispatch } from "store/hooks";
// import { FilterCategoryType } from "features/sidebar/sections/filter/filter";
// import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
// import { SectionContainer } from "features/section/SectionContainer";
import Button, { buttonColor } from "components/buttons/Button";
import { NEW_ADDRESS } from "constants/routes";
import { useLocation } from "react-router";
import useTranslations from "services/i18n/useTranslations";
import { AllViewsResponse } from "features/viewManagement/types";
import { OnChainResponse } from "./types";
import DefaultCellRenderer from "../../table/DefaultCellRenderer";
import { DefaultOnChainView } from "./onChainDefaults";

type sections = {
  filter: boolean;
  sort: boolean;
  columns: boolean;
};

const statusTypes: any = {
  OPEN: "Open",
  SETTLED: "Settled",
  EXPIRED: "Expired",
};
function OnChainPage() {
  const dispatch = useAppDispatch();
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const [limit, setLimit] = useLocalStorage("invoicesLimit", 100);
  const [offset, setOffset] = useState(0);

  const allViews = useGetTableViewsQuery<{
    data: AllViewsResponse;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();
  const views = allViews?.data ? allViews.data["onChain"] : [DefaultOnChainView];
  const [selectedView, setSelectedView] = useState(0);

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
      // order: invoiceViews[selectedView].sortBy,
    },
    { skip: !allViews.isSuccess }
  );

  // useEffect(() => {
  //   const views: ViewInterface<OnChainTx>[] = [];
  //   if (!isLoading) {
  //     if (onchainViews) {
  //       onchainViews?.map((v: ViewInterface<OnChainTx>) => {
  //         views.push(v.view);
  //       });
  //
  //       dispatch(updateViews({ views, index: 0 }));
  //     } else {
  //       dispatch(updateViews({ views: [{ ...DefaultView, title: "Default View" }], index: 0 }));
  //     }
  //   }
  // }, [onchainViews, isLoading]);
  //
  // const [limit, setLimit] = useLocalStorage("onchainLimit", 100);
  // const [offset, setOffset] = useState(0);
  // const [orderBy, setOrderBy] = useLocalStorage("onchainOrderBy", [
  //   {
  //     key: "date",
  //     direction: "desc",
  //   },
  // ] as OrderBy[]);
  //
  // const activeColumns = useAppSelector(selectActiveColumns) || [];
  // const allColumns = useAppSelector(selectAllColumns);
  //
  // const navigate = useNavigate();
  // const filters = useAppSelector(selectOnChainFilters);
  //
  // const onchainResponse = useGetOnChainTxQuery({
  //   limit: limit,
  //   offset: offset,
  //   order: orderBy,
  // });
  //
  // // Logic for toggling the sidebar
  // const [sidebarExpanded, setSidebarExpanded] = useState(false);
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
  //
  // // General logic for toggling the sidebar sections
  // const initialSectionState: sections = {
  //   filter: false,
  //   sort: false,
  //   columns: false,
  // };
  //
  // const [activeSidebarSections, setActiveSidebarSections] = useState(initialSectionState);
  //
  // const sidebarSectionHandler = (section: keyof sections) => {
  //   return () => {
  //     setActiveSidebarSections({
  //       ...activeSidebarSections,
  //       [section]: !activeSidebarSections[section],
  //     });
  //   };
  // };
  //
  // const closeSidebarHandler = () => {
  //   return () => {
  //     setSidebarExpanded(false);
  //   };
  // };
  //
  // const location = useLocation();
  // const { t } = useTranslations();
  //
  // const [updateTableView] = useUpdateTableViewMutation();
  // const [createTableView] = useCreateTableViewMutation();
  // const currentViewIndex = useAppSelector(selectedViewIndex);
  // const currentView = useAppSelector(selectCurrentView);
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
        {/*<TableControlsButton*/}
        {/*  onClickHandler={() => setSidebarExpanded(!sidebarExpanded)}*/}
        {/*  icon={OptionsIcon}*/}
        {/*  id={"tableControlsButton"}*/}
        {/*/>*/}
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  // const defaultFilter: FilterInterface = {
  //   funcName: "gte",
  //   category: "number" as FilterCategoryType,
  //   parameter: 0,
  //   key: "amount",
  // };
  // const filterColumns = useAppSelector(selectAllColumns);
  //
  // const handleFilterUpdate = (updated: Clause) => {
  //   dispatch(updateOnChainFilters({ filters: updated.toJSON() }));
  // };

  // const sortableColumns = allColumns.filter((column: ColumnMetaData<OnChainTx>) =>
  //   [
  //     "date",
  //     "destAddresses",
  //     "destAddressesCount",
  //     "amount",
  //     "totalFees",
  //     "label",
  //     "lndTxTypeLabel",
  //     "lndShortChanId",
  //   ].includes(column.key)
  // );

  // const handleSortUpdate = (updated: Array<OrderBy>) => {
  //   setOrderBy(updated);
  // };

  // const updateColumnsHandler = (columns: Array<any>) => {
  //   dispatch(updateColumns({ columns: columns }));
  // };

  // const sidebar = (
  //   <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
  //     <SectionContainer
  //       title={"Columns"}
  //       icon={ColumnsIcon}
  //       expanded={activeSidebarSections.columns}
  //       handleToggle={sidebarSectionHandler("columns")}
  //     >
  //       <ColumnsSection columns={allColumns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
  //     </SectionContainer>
  //     <SectionContainer
  //       title={"Filter"}
  //       icon={FilterIcon}
  //       expanded={activeSidebarSections.filter}
  //       handleToggle={sidebarSectionHandler("filter")}
  //     >
  //       <FilterSection
  //         columnsMeta={filterColumns}
  //         filters={filters}
  //         filterUpdateHandler={handleFilterUpdate}
  //         defaultFilter={defaultFilter}
  //       />
  //     </SectionContainer>
  //     <SectionContainer
  //       title={"Sort"}
  //       icon={SortIcon}
  //       expanded={activeSidebarSections.sort}
  //       handleToggle={sidebarSectionHandler("sort")}
  //     >
  //       <SortSection columns={sortableColumns} orderBy={orderBy} updateHandler={handleSortUpdate} />
  //     </SectionContainer>
  //   </Sidebar>
  // );

  const breadcrumbs = [
    <span key="b1">Transactions</span>,
    <Link key="b2" to={"/transactions/on-chain"}>
      On-Chain Tx
    </Link>,
  ];
  const pagination = (
    <Pagination
      limit={limit}
      offset={offset}
      total={onChainTxResponse?.data?.pagination?.total || 0}
      perPageHandler={setLimit}
      offsetHandler={setOffset}
    />
  );
  return (
    <TablePageTemplate
      title={"OnChain"}
      breadcrumbs={breadcrumbs}
      // sidebarExpanded={sidebarExpanded}
      // sidebar={sidebar}
      tableControls={tableControls}
      pagination={pagination}
    >
      <Table
        cellRenderer={DefaultCellRenderer}
        data={onChainTxResponse?.data?.data || []}
        activeColumns={views[selectedView].columns || []}
        isLoading={onChainTxResponse.isLoading || onChainTxResponse.isFetching || onChainTxResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default OnChainPage;
