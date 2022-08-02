import { Link } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  ArrowJoin20Regular as GroupIcon,
} from "@fluentui/react-icons";
import Sidebar, { SidebarSection } from "../sidebar/Sidebar";

import { Clause } from "features/sidebar/sections/filter/filter";

import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "../tablePageTemplate/TablePageTemplate";
import React, { useEffect, useMemo, useState } from "react";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import {
  updateColumns,
  selectActiveColumns,
  selectAllColumns,
  ColumnMetaData,
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
import SortSection, { SortByOptionType } from "../sidebar/sections/sort/SortSection";
import GroupBySection from "../sidebar/sections/group/GroupBySection";
import clone from "../../clone";
import ForwardsDataWrapper from "./ForwardsDataWrapper";

type sections = {
  filter: boolean;
  sort: boolean;
  group: boolean;
  columns: boolean;
};

function ForwardsPage() {
  const dispatch = useAppDispatch();

  // const viewResponse = useGetTableViewsQuery();

  const activeColumns = useAppSelector(selectActiveColumns) || [];
  const columns = useAppSelector(selectAllColumns);
  const sortBy = useAppSelector(selectSortBy);
  const groupBy = useAppSelector(selectGroupBy) || "channels";
  const filters = useAppSelector(selectFilters);

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

  // useEffect(() => {
  //   if (viewId === "" && !viewResponse.isLoading) {
  //     // window.history.replaceState(null, "", "/analyse/forwards/" + viewResponse.data[0].title);
  //     navigate("/analyse/forwards/" + (viewResponse.data[0].view.title || "").replace(/\s+/g, "-").toLowerCase(), {
  //       replace: true,
  //     });
  //   }
  // }, [viewResponse.isLoading]);

  const setSection = (section: keyof sections) => {
    return () => {
      if (activeSidebarSections[section] && sidebarExpanded) {
        setSidebarExpanded(false);
        setActiveSidebarSections(initialSectionState);
      } else {
        setSidebarExpanded(true);
        setActiveSidebarSections({
          ...initialSectionState,
          [section]: true,
        });
      }
    };
  };

  const sidebarSectionHandler = (section: keyof sections) => {
    return () => {
      setActiveSidebarSections({
        ...initialSectionState,
        [section]: !activeSidebarSections[section],
      });
    };
  };

  const closeSidebarHandler = () => {
    return () => {
      setSidebarExpanded(false);
      setActiveSidebarSections(initialSectionState);
    };
  };

  const tableControls = (
    <TableControlSection>
      <TableControlsTabsGroup>{<ViewsPopover />}</TableControlsTabsGroup>
      <TableControlsButtonGroup>
        <TableControlsButton
          onClickHandler={setSection("columns")}
          icon={ColumnsIcon}
          active={activeSidebarSections.columns}
        />
        <TableControlsButton
          onClickHandler={setSection("filter")}
          icon={FilterIcon}
          active={activeSidebarSections.filter}
        />
        <TableControlsButton onClickHandler={setSection("sort")} icon={SortIcon} active={activeSidebarSections.sort} />
        <TableControlsButton
          onClickHandler={setSection("group")}
          icon={GroupIcon}
          active={activeSidebarSections.group}
        />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const updateColumnsHandler = (columns: ColumnMetaData[]) => {
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

  const sidebar = (
    <Sidebar title={"Table Options"} closeSidebarHandler={closeSidebarHandler()}>
      <SidebarSection
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        sectionToggleHandler={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={columns} activeColumns={activeColumns} handleUpdateColumn={updateColumnsHandler} />
      </SidebarSection>

      <SidebarSection
        title={"Filter"}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        sectionToggleHandler={sidebarSectionHandler("filter")}
      >
        <FilterSection filters={filters} filterUpdateHandler={handleFilterUpdate} />
      </SidebarSection>

      <SidebarSection
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        sectionToggleHandler={sidebarSectionHandler("sort")}
      >
        <SortSection columns={columns} sortBy={sortBy} updateSortByHandler={handleSortUpdate} />
      </SidebarSection>

      <SidebarSection
        title={"Group"}
        icon={GroupIcon}
        expanded={activeSidebarSections.group}
        sectionToggleHandler={sidebarSectionHandler("group")}
      >
        <GroupBySection groupBy={groupBy} groupByHandler={handleGroupByUpdate} />
      </SidebarSection>
    </Sidebar>
  );

  const breadcrumbs = ["Analyse", <Link to={"/analyse/forwards"}>Forwards</Link>];

  return (
    <TablePageTemplate
      title={"Forwards"}
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
