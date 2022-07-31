import Table from "../table/tableContent/Table";
import { useGetChannelsQuery, useGetTableViewsQuery } from "apiSlice";
import { Link, Navigate, useNavigate } from "react-router-dom";
import {
  Filter20Regular as FilterIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  ArrowJoin20Regular as GroupIcon,
} from "@fluentui/react-icons";
import Sidebar, { SidebarSection } from "../sidebar/Sidebar";
import { useParams } from "react-router-dom";
import { orderBy, cloneDeep } from "lodash";
import { applyFilters, Clause, deserialiseQuery } from "features/sidebar/sections/filter/filter";
import { groupByFn } from "features/sidebar/sections/group/groupBy";

import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "../tablePageTemplate/TablePageTemplate";
import { useEffect, useMemo, useState } from "react";
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
import { selectTimeInterval } from "../timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import SortSection, { SortByOptionType } from "../sidebar/sections/sort/SortSection";
import GroupBySection from "../sidebar/sections/group/GroupBySection";
import clone from "../../clone";

type sections = {
  filter: boolean;
  sort: boolean;
  group: boolean;
  columns: boolean;
};

function ForwardsPage() {
  // initial getting of the table views from the database

  const dispatch = useAppDispatch();

  const currentPeriod = useAppSelector(selectTimeInterval);
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");

  const viewResponse = useGetTableViewsQuery();
  const chanResponse = useGetChannelsQuery({ from: from, to: to });

  const columns = useAppSelector(selectAllColumns);
  const sortBy = useAppSelector(selectSortBy);
  const groupBy = useAppSelector(selectGroupBy) || "channels";
  const filters = useAppSelector(selectFilters);

  const activeColumns = clone<Array<ColumnMetaData>>(useAppSelector(selectActiveColumns)) || [];
  const data = clone<ColumnMetaData[]>(chanResponse.data) || [];

  // useEffect(() => {
  //   if (viewId === "" && !viewResponse.isLoading) {
  //     // window.history.replaceState(null, "", "/analyse/forwards/" + viewResponse.data[0].title);
  //     navigate("/analyse/forwards/" + (viewResponse.data[0].view.title || "").replace(/\s+/g, "-").toLowerCase(), {
  //       replace: true,
  //     });
  //   }
  // }, [viewResponse.isLoading]);

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
      <TableControlsTabsGroup>{!viewResponse.isLoading && <ViewsPopover />}</TableControlsTabsGroup>
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

  const channels = useMemo(() => {
    let channels = cloneDeep(data ? data : ([] as any[]));

    if (channels.length > 0) {
      channels = groupByFn(channels, groupBy || "channels");
    }
    if (filters) {
      let f = deserialiseQuery(clone<Clause>(filters));
      channels = applyFilters(f, channels);
    }
    return orderBy(
      channels,
      sortBy.map((s) => s.value),
      sortBy.map((s) => s.direction) as ["asc" | "desc"]
    );
  }, [chanResponse.data, filters, groupBy, sortBy]);

  /* const channels = data || []; */
  if (channels.length > 0) {
    for (const channel of channels) {
      for (const column of activeColumns) {
        column.total = (column.total ?? 0) + channel[column.key];
        column.max = Math.max(column.max ?? 0, channel[column.key] ?? 0);
      }
    }

    const turnover_total_col = activeColumns.find((col) => col.key === "turnover_total");
    const turnover_out_col = activeColumns.find((col) => col.key === "turnover_out");
    const turnover_in_col = activeColumns.find((col) => col.key === "turnover_in");
    const amount_total_col = activeColumns.find((col) => col.key === "amount_total");
    const amount_out_col = activeColumns.find((col) => col.key === "amount_out");
    const amount_in_col = activeColumns.find((col) => col.key === "amount_in");
    const capacity_col = activeColumns.find((col) => col.key === "capacity");
    if (turnover_total_col) {
      turnover_total_col.total = (amount_total_col?.total ?? 0) / (capacity_col?.total ?? 1);
    }

    if (turnover_out_col) {
      turnover_out_col.total = (amount_out_col?.total ?? 0) / (capacity_col?.total ?? 1);
    }

    if (turnover_in_col) {
      turnover_in_col.total = (amount_in_col?.total ?? 0) / (capacity_col?.total ?? 1);
    }
  }

  return (
    <TablePageTemplate
      title={"Forwards"}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
    >
      <Table
        activeColumns={activeColumns}
        data={channels}
        isLoading={chanResponse.isLoading || chanResponse.isFetching || chanResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default ForwardsPage;
