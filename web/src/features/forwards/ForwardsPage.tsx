import { Link } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  ArrowJoin20Regular as GroupIcon,
  Save20Regular as SaveIcon,
  Options20Regular as OptionsIcon,
} from "@fluentui/react-icons";
import Sidebar, { SidebarSection } from "../sidebar/Sidebar";
import { useUpdateTableViewMutation, useCreateTableViewMutation, useGetTableViewsQuery } from "apiSlice";

import { Clause, FilterCategoryType, FilterInterface } from "features/sidebar/sections/filter/filter";

import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "../templates/tablePageTemplate/TablePageTemplate";
import React, { useState } from "react";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import {
  updateColumns,
  selectActiveColumns,
  selectAllColumns,
  selectFilters,
  updateFilters,
  selectSortBy,
  updateSortBy,
  selectGroupBy,
  updateGroupBy,
} from "./forwardsSlice";
import ViewsPopover from "./views/ViewsPopover";
import ColumnsSection from "../sidebar/sections/columns/ColumnsSection";
import FilterSection from "../sidebar/sections/filter/FilterSection";
import SortSection, { SortByOptionType } from "../sidebar/sections/sort/SortSectionOld";
import GroupBySection from "../sidebar/sections/group/GroupBySection";
import ForwardsDataWrapper from "./ForwardsDataWrapper";
import Button, { buttonColor } from "../buttons/Button";
import { selectCurrentView, selectedViewIndex } from "features/forwards/forwardsSlice";
import classNames from "classnames";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";

type sections = {
  filter: boolean;
  sort: boolean;
  group: boolean;
  columns: boolean;
};

function ForwardsPage() {
  const dispatch = useAppDispatch();

  useGetTableViewsQuery();

  // const viewResponse = useGetTableViewsQuery();
  const currentView = useAppSelector(selectCurrentView);
  const currentViewIndex = useAppSelector(selectedViewIndex);
  const [updateTableView] = useUpdateTableViewMutation();
  const [createTableView] = useCreateTableViewMutation();
  const activeColumns = useAppSelector(selectActiveColumns) || [];
  const columns = useAppSelector(selectAllColumns);
  const sortBy = useAppSelector(selectSortBy);
  const groupBy = useAppSelector(selectGroupBy) || "channels";
  const filters = useAppSelector(selectFilters);

  const saveView = () => {
    const viewMod = { ...currentView };
    viewMod.saved = true;
    if (currentView.id === undefined || null) {
      createTableView({ view: viewMod, index: currentViewIndex });
      return;
    }
    updateTableView(viewMod);
  };

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  // General logic for toggling the sidebar sections
  const initialSectionState: sections = {
    filter: false,
    sort: false,
    columns: false,
    group: false,
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

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          {/*<TimeIntervalSelect />*/}
          {<ViewsPopover />}
        </TableControlsTabsGroup>
        <TableControlsButton onClickHandler={() => setSidebarExpanded(!sidebarExpanded)} icon={OptionsIcon} />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const updateColumnsHandler = (columns: Array<any>) => {
    dispatch(updateColumns({ columns: columns }));
  };

  const handleFilterUpdate = (filters: Clause) => {
    dispatch(updateFilters({ filters: filters.toJSON() }));
  };

  const handleSortUpdate = (updated: Array<SortByOptionType>) => {
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
    <Sidebar title={"Table Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SidebarSection
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={columns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
      </SidebarSection>

      <SidebarSection
        title={"Filter"}
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
      </SidebarSection>

      <SidebarSection
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={columns} orderBy={sortBy} updateSortByHandler={handleSortUpdate} />
      </SidebarSection>

      <SidebarSection
        title={"Group"}
        icon={GroupIcon}
        expanded={activeSidebarSections.group}
        handleToggle={sidebarSectionHandler("group")}
      >
        <GroupBySection groupBy={groupBy} groupByHandler={handleGroupByUpdate} />
      </SidebarSection>
    </Sidebar>
  );

  const breadcrumbs = ["Analyse", <Link to={"/analyse/forwards"}>Forwards</Link>];

  return (
    <TablePageTemplate
      title={"Forwards"}
      titleContent={<TimeIntervalSelect />}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
    >
      <ForwardsDataWrapper activeColumns={activeColumns} />
    </TablePageTemplate>
  );
}

export default ForwardsPage;
