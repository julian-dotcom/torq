import { Link } from "react-router-dom";
import {
  MoneySettings20Regular as AdjustFeesIcon,
  // Filter20Regular as FilterIcon,
  // ArrowSortDownLines20Regular as SortIcon,
  // ColumnTriple20Regular as ColumnsIcon,
  // ArrowJoin20Regular as GroupIcon,
  Options20Regular as OptionsIcon,
  // Save20Regular as SaveIcon,
  ArrowRouting20Regular as ChannelsIcon,
} from "@fluentui/react-icons";
// import Sidebar from "features/sidebar/Sidebar";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
// import { Clause, FilterCategoryType, FilterInterface } from "features/sidebar/sections/filter/filter";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
// import { useAppDispatch, useAppSelector } from "store/hooks";
// import ViewsPopover from "features/viewManagement/ViewsPopover";
// import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
// import FilterSection from "features/sidebar/sections/filter/FilterSection";
// import SortSection, { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
// import GroupBySection from "features/sidebar/sections/group/GroupBySection";
// import ChannelsDataWrapper from "./ChannelsDataWrapper";
// import { SectionContainer } from "features/section/SectionContainer";
// import { ColumnMetaData } from "features/table/types";
import Button, { buttonColor } from "components/buttons/Button";
import { useNavigate } from "react-router-dom";
import { useLocation } from "react-router";
import { UPDATE_CHANNEL, OPEN_CHANNEL } from "constants/routes";
import { channel } from "./channelsTypes";
import useTranslations from "services/i18n/useTranslations";
import { AllViewsResponse } from "features/viewManagement/types";
import { Sections } from "features/sidebar/sections/types";
import DefaultCellRenderer from "../table/DefaultCellRenderer";
import Table from "../table/Table";
// import useLocalStorage from "../helpers/useLocalStorage";
import { DefaultChannelsView } from "./channelsDefaults";
import { useGetChannelsQuery } from "../../apiSlice";

function ChannelsPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  // const [limit, setLimit] = useLocalStorage("invoicesLimit", 100);
  // const [offset, setOffset] = useState(0);

  const allViews = useGetTableViewsQuery<{
    data: AllViewsResponse;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();
  const views = allViews?.data ? allViews.data["channel"] : [DefaultChannelsView];
  const [selectedView, setSelectedView] = useState(0);

  const channelsResponse = useGetChannelsQuery<{
    data: Array<channel>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();

  //
  // const { data: channelsViews, isLoading } = useGetTableViewsQuery<{
  //   data: Array<ViewInterface<channel>>;
  //   isLoading: true;
  // }>(undefined, { skip: !allViews.isSuccess });
  // // const { data: channelsViews, isLoading } = useGetTableViewsQuery({ page: "channels" });

  // const views = useEffect(() => {
  //   const views: ViewInterface<channel>[] = [];
  //   if (channelsViews) {
  //     channelsViews?.map((v: ViewResponse<channel>) => {
  //       views.push(v.view);
  //     });
  //   } else {
  //     dispatch(updateViews({ views: [{ ...DefaultView, title: "Default View" }], index: 0 }));
  //   }
  //   return views;
  // }, [channelsViews, isLoading]);

  // const activeColumns = useAppSelector(selectActiveColumns) || [];
  // const columns = useAppSelector(selectAllColumns);
  // const sortBy = useAppSelector(selectSortBy) || [];
  // const groupBy = useAppSelector(selectGroupBy) || "channels";
  // const filters = useAppSelector(selectFilters);

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  // General logic for toggling the sidebar sections
  const initialSectionState: Sections = {
    filter: false,
    sort: false,
    columns: false,
    group: false,
  };

  const [activeSidebarSections, setActiveSidebarSections] = useState(initialSectionState);

  const sidebarSectionHandler = (section: keyof Sections) => {
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

  // const [updateTableView] = useUpdateTableViewMutation();
  // const [createTableView] = useCreateTableViewMutation();
  // const currentViewIndex = useAppSelector(selectedViewIndex);
  // const currentView = useAppSelector(selectCurrentView);
  //
  // const saveView = () => {
  //   const viewMod = { ...currentView };
  //   viewMod.saved = true;
  //   if (currentView.id === undefined || null) {
  //     createTableView({ view: viewMod, index: currentViewIndex, page: "channels" });
  //     return;
  //   }
  //   updateTableView(viewMod);
  // };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          {/*<ViewsPopover*/}
          {/*  page="channels"*/}
          {/*  selectViews={selectViews}*/}
          {/*  updateViews={updateViews}*/}
          {/*  updateSelectedView={updateSelectedView}*/}
          {/*  selectedViewIndex={selectedViewIndex}*/}
          {/*  updateViewsOrder={updateViewsOrder}*/}
          {/*  DefaultView={DefaultView}*/}
          {/*/>*/}
          {/*{!currentView.saved && (*/}
          {/*  <Button*/}
          {/*    buttonColor={buttonColor.green}*/}
          {/*    icon={<SaveIcon />}*/}
          {/*    text={"Save"}*/}
          {/*    onClick={saveView}*/}
          {/*    className={"collapse-tablet"}*/}
          {/*  />*/}
          {/*)}*/}
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
        <Button
          buttonColor={buttonColor.green}
          text={"Open Channel"}
          className={"collapse-tablet"}
          icon={<ChannelsIcon />}
          onClick={() => {
            navigate(OPEN_CHANNEL, { state: { background: location } });
          }}
        />
        <Button
          buttonColor={buttonColor.green}
          text={t.updateChannelPolicy.title}
          icon={<AdjustFeesIcon />}
          onClick={() => {
            navigate(UPDATE_CHANNEL, { state: { background: location } });
          }}
        />
        <TableControlsButton onClickHandler={() => setSidebarExpanded(!sidebarExpanded)} icon={OptionsIcon} />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  // const updateColumnsHandler = (columns: ColumnMetaData<channel>[]) => {
  //   dispatch(updateColumns({ columns }));
  // };
  //
  // const handleFilterUpdate = (filters: Clause) => {
  //   dispatch(updateFilters({ filters: filters.toJSON() }));
  // };
  //
  // const handleSortUpdate = (updated: SortByOptionType[]) => {
  //   dispatch(updateSortBy({ sortBy: updated }));
  // };
  //
  // const handleGroupByUpdate = (updated: string) => {
  //   dispatch(updateGroupBy({ groupBy: updated }));
  // };

  // const sidebar = (
  //   <Sidebar title={t.tableLayout.tableOptionsTitle} closeSidebarHandler={closeSidebarHandler()}>
  //     <SectionContainer
  //       title={t.columns}
  //       icon={ColumnsIcon}
  //       expanded={activeSidebarSections.columns}
  //       handleToggle={sidebarSectionHandler("columns")}
  //     >
  //       <ColumnsSection columns={columns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
  //     </SectionContainer>
  //
  //     <SectionContainer
  //       title={t.filter}
  //       icon={FilterIcon}
  //       expanded={activeSidebarSections.filter}
  //       handleToggle={sidebarSectionHandler("filter")}
  //     >
  //       <FilterSection
  //         columnsMeta={columns}
  //         filters={filters}
  //         filterUpdateHandler={handleFilterUpdate}
  //         defaultFilter={defaultFilter}
  //       />
  //     </SectionContainer>
  //
  //     <SectionContainer
  //       title={t.sort}
  //       icon={SortIcon}
  //       expanded={activeSidebarSections.sort}
  //       handleToggle={sidebarSectionHandler("sort")}
  //     >
  //       <SortSection columns={columns} orderBy={sortBy} updateSortByHandler={handleSortUpdate} />
  //     </SectionContainer>
  //
  //     <SectionContainer
  //       title={t.group}
  //       icon={GroupIcon}
  //       expanded={activeSidebarSections.group}
  //       handleToggle={sidebarSectionHandler("group")}
  //     >
  //       <GroupBySection groupBy={groupBy} groupByHandler={handleGroupByUpdate} />
  //     </SectionContainer>
  //   </Sidebar>
  // );

  const breadcrumbs = [
    <span key="b1">Analyse</span>,
    <Link key="b2" to={"/analyse/channels"}>
      {t.channels}
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={t.channels}
      titleContent={""}
      breadcrumbs={breadcrumbs}
      // sidebarExpanded={sidebarExpanded}
      // sidebar={sidebar}
      tableControls={tableControls}
    >
      <Table
        cellRenderer={DefaultCellRenderer}
        data={channelsResponse?.data || []}
        activeColumns={views[selectedView].columns || []}
        isLoading={channelsResponse.isLoading || channelsResponse.isFetching || channelsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default ChannelsPage;
