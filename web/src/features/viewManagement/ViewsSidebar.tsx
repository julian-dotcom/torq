import {
  ArrowJoin20Regular as GroupIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Filter20Regular as FilterIcon,
  TableMultipleRegular as ViewsIcon,
} from "@fluentui/react-icons";
import Sidebar from "features/sidebar/Sidebar";
import ViewsPopover from "./ViewSection";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import { ColumnMetaData } from "features/table/types";
import { useState } from "react";
import { ViewResponse } from "./types";
import FilterSection from "features/sidebar/sections/filter/FilterSection";
import { SectionContainer } from "features/section/SectionContainer";
import GroupBySection from "features/sidebar/sections/group/GroupBySection";
import SortSection, { OrderBy } from "features/sidebar/sections/sort/SortSection";
import useTranslations from "services/i18n/useTranslations";
import { AndClause, deserialiseQuery, OrClause } from "features/sidebar/sections/filter/filter";
import { userEvents } from "utils/userEvents";

type ViewSidebarProps<T> = {
  expanded: boolean;
  onExpandToggle: (event?: React.MouseEvent<HTMLDivElement, MouseEvent>) => void;
  viewResponse: ViewResponse<T>;
  selectedViewIndex: number;
  defaultView: ViewResponse<T>;
  allColumns: Array<ColumnMetaData<T>>;
  filterableColumns: ColumnMetaData<T>[];
  sortableColumns: ColumnMetaData<T>[];
  sortByTemplate: OrderBy;
  filterTemplate: any; // eslint-disable-line @typescript-eslint/no-explicit-any
  enableGroupBy?: boolean;
};

export default function ViewsSidebar<T>(props: ViewSidebarProps<T>) {
  const { t } = useTranslations();
  const { track } = userEvents();
  // General logic for toggling the sidebar sections
  const initialSectionState = {
    views: true,
    filter: false,
    sort: false,
    columns: false,
    group: false,
  };

  const [activeSidebarSections, setActiveSidebarSections] = useState(initialSectionState);

  const sidebarSectionHandler = (section: keyof typeof initialSectionState) => {
    return () => {
      track(`Toggle Sidebar Section`, {
        section: section,
        expanded: !activeSidebarSections[section],
      });
      setActiveSidebarSections({
        ...activeSidebarSections,
        [section]: !activeSidebarSections[section],
      });
    };
  };

  const filterClause = deserialiseQuery(props.viewResponse.view.filters) as AndClause | OrClause;

  return (
    <Sidebar title={"Options"} closeSidebarHandler={() => props.onExpandToggle()}>
      <SectionContainer
        title={t.views}
        icon={ViewsIcon}
        expanded={activeSidebarSections.views}
        handleToggle={sidebarSectionHandler("views")}
      >
        <ViewsPopover page={props.viewResponse.page} defaultView={props.defaultView} />
      </SectionContainer>
      <SectionContainer
        title={t.columns}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection
          allColumns={props.allColumns}
          activeColumns={props.viewResponse.view.columns}
          page={props.viewResponse.page}
          viewIndex={props.selectedViewIndex}
        />
      </SectionContainer>
      <SectionContainer
        title={t.filter}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        handleToggle={sidebarSectionHandler("filter")}
      >
        <FilterSection
          filterableColumns={props.filterableColumns}
          page={props.viewResponse.page}
          defaultFilter={props.filterTemplate}
          filters={filterClause}
          viewIndex={props.selectedViewIndex}
        />
      </SectionContainer>
      <SectionContainer
        title={t.sort}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection
          columns={props.sortableColumns}
          sortBy={props.viewResponse.view.sortBy}
          defaultSortBy={props.sortByTemplate}
          page={props.viewResponse.page}
          viewIndex={props.selectedViewIndex}
        />
      </SectionContainer>
      {props.enableGroupBy && (
        <SectionContainer
          title={t.group}
          icon={GroupIcon}
          expanded={activeSidebarSections.group}
          handleToggle={sidebarSectionHandler("group")}
        >
          <GroupBySection
            page={props.viewResponse.page}
            groupBy={props.viewResponse.view.groupBy}
            viewIndex={props.selectedViewIndex}
          />
        </SectionContainer>
      )}
    </Sidebar>
  );
}
