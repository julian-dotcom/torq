import { Link } from "react-router-dom";
import {
  MoneySettings20Regular as AdjustFeesIcon,
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  ArrowJoin20Regular as GroupIcon,
  Options20Regular as OptionsIcon,
  Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import Sidebar from "features/sidebar/Sidebar";
import { useCreateTableViewMutation, useGetTableViewsQuery, useUpdateTableViewMutation } from "apiSlice";
import { Clause, FilterCategoryType, FilterInterface } from "features/sidebar/sections/filter/filter";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useEffect, useState } from "react";
import { useAppDispatch, useAppSelector } from "store/hooks";
import { selectCurrentView, selectedViewIndex } from "features/channels/ChannelsSlice";
import {
  selectViews,
  updateViews,
  updateSelectedView,
  updateViewsOrder,
  DefaultView,
  updateColumns,
  selectActiveColumns,
  selectAllColumns,
  selectFilters,
  updateFilters,
  selectSortBy,
  updateSortBy,
  selectGroupBy,
  updateGroupBy,
} from "./ChannelsSlice";
import ViewsPopover from "features/viewManagement/ViewsPopover";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import FilterSection from "features/sidebar/sections/filter/FilterSection";
import SortSection, { SortByOptionType } from "features/sidebar/sections/sort/SortSectionOld";
import GroupBySection from "features/sidebar/sections/group/GroupBySection";
import ChannelsDataWrapper from "./ChannelsDataWrapper";
import { SectionContainer } from "features/section/SectionContainer";
import { ColumnMetaData } from "features/table/Table";
import Button, { buttonColor } from "features/buttons/Button";
import { useNavigate } from "react-router-dom";
import { useLocation } from "react-router";
import { UPDATE_CHANNEL } from "constants/routes";
import { Sections } from "./channelsTypes";
import useTranslations from "services/i18n/useTranslations";
import { ViewResponse } from "features/viewManagement/ViewsPopover";
import { ViewInterface } from "features/table/Table";

function ChannelsPage() {
  const dispatch = useAppDispatch();
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const { data: channelsViews, isLoading } = useGetTableViewsQuery({page: 'channels'});

  useEffect(() => {
    const views: ViewInterface[] = [];
    if (channelsViews) {
      channelsViews?.map((v: ViewResponse) => {
        views.push(v.view)
      });

      dispatch(updateViews({ views, index: 0 }));
    } else {
      dispatch(updateViews({ views: [{...DefaultView, title: "Default View"}], index: 0 }));
    }
  }, [channelsViews, isLoading]);

  const activeColumns = useAppSelector(selectActiveColumns) || [];
  const columns = useAppSelector(selectAllColumns);
  const sortBy = useAppSelector(selectSortBy);
  const groupBy = useAppSelector(selectGroupBy) || "channels";
  const filters = useAppSelector(selectFilters);

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

  const [updateTableView] = useUpdateTableViewMutation();
  const [createTableView] = useCreateTableViewMutation();
  const currentViewIndex = useAppSelector(selectedViewIndex);
  const currentView = useAppSelector(selectCurrentView);
  const saveView = () => {
    const viewMod = { ...currentView };
    viewMod.saved = true;
    if (currentView.id === undefined || null) {
      createTableView({ view: viewMod, index: currentViewIndex, page: 'channels' });
      return;
    }
    updateTableView(viewMod);
  };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <ViewsPopover
            page="channels"
            selectViews={selectViews}
            updateViews={updateViews}
            updateSelectedView={updateSelectedView}
            selectedViewIndex={selectedViewIndex}
            updateViewsOrder={updateViewsOrder}
            DefaultView={DefaultView}
          />
          {!currentView.saved && (
            <Button
              buttonColor={buttonColor.green}
              icon={<SaveIcon />}
              text={"Save"}
              onClick={saveView}
              className={"collapse-tablet"}
            />
          )}
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
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

  const updateColumnsHandler = (columns: ColumnMetaData[]) => {
    dispatch(updateColumns({ columns }));
  };

  const handleFilterUpdate = (filters: Clause) => {
    dispatch(updateFilters({ filters: filters.toJSON() }));
  };

  const handleSortUpdate = (updated: SortByOptionType[]) => {
    dispatch(updateSortBy({ sortBy: updated }));
  };

  const handleGroupByUpdate = (updated: string) => {
    dispatch(updateGroupBy({ groupBy: updated }));
  };

  const defaultFilter: FilterInterface = {
    funcName: "gte",
    category: "number" as FilterCategoryType,
    parameter: 0,
    key: "capacity",
  };

  const sidebar = (
    <Sidebar title={t.tableLayout.tableOptionsTitle} closeSidebarHandler={closeSidebarHandler()}>
      <SectionContainer
        title={t.columns}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={columns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
      </SectionContainer>

      <SectionContainer
        title={t.filter}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        handleToggle={sidebarSectionHandler("filter")}
      >
        <FilterSection
          columnsMeta={columns}
          filters={filters}
          filterUpdateHandler={handleFilterUpdate}
          defaultFilter={defaultFilter}
        />
      </SectionContainer>

      <SectionContainer
        title={t.sort}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={columns} orderBy={sortBy} updateSortByHandler={handleSortUpdate} />
      </SectionContainer>

      <SectionContainer
        title={t.group}
        icon={GroupIcon}
        expanded={activeSidebarSections.group}
        handleToggle={sidebarSectionHandler("group")}
      >
        <GroupBySection groupBy={groupBy} groupByHandler={handleGroupByUpdate} />
      </SectionContainer>
    </Sidebar>
  );

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
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
    >
      <>
        <ChannelsDataWrapper activeColumns={activeColumns} />
      </>
    </TablePageTemplate>
  );
}

export default ChannelsPage;
