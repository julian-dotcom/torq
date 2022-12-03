import {
  ArrowJoin20Regular as GroupIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Filter20Regular as FilterIcon,
  // Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import Sidebar from "features/sidebar/Sidebar";
import ViewsPopover from "../viewManagement/ViewsPopover";
import ColumnsSection from "../sidebar/sections/columns/ColumnsSection";
import { ColumnMetaData } from "../table/types";
import { useState } from "react";
import { ViewResponse } from "./types";
import FilterSection from "../sidebar/sections/filter/FilterSection";
import { SectionContainer } from "../section/SectionContainer";
import GroupBySection from "../sidebar/sections/group/GroupBySection";
import SortSection, { OrderBy } from "../sidebar/sections/sort/SortSection";
import useTranslations from "../../services/i18n/useTranslations";

type ViewSidebarProps<T> = {
  expanded: boolean;
  onExpandToggle: (event?: React.MouseEvent<HTMLDivElement, MouseEvent>) => void;
  viewResponse: ViewResponse<T>;
  defaultView: ViewResponse<T>;
  allColumns: Array<ColumnMetaData<T>>;
  filterableColumns: ColumnMetaData<T>[];
  sortableColumns: ColumnMetaData<T>[];
  sortByTemplate: OrderBy;
  filterTemplate: any;
  enableGroupBy?: boolean;
};

export default function ViewsSidebar<T>(props: ViewSidebarProps<T>) {
  const { t } = useTranslations();
  // General logic for toggling the sidebar sections
  const initialSectionState = {
    filter: false,
    sort: false,
    columns: false,
    group: false,
  };

  const [activeSidebarSections, setActiveSidebarSections] = useState(initialSectionState);

  const sidebarSectionHandler = (section: keyof typeof initialSectionState) => {
    return () => {
      setActiveSidebarSections({
        ...activeSidebarSections,
        [section]: !activeSidebarSections[section],
      });
    };
  };

  return (
    <Sidebar title={"Options"} closeSidebarHandler={props.onExpandToggle}>
      <ViewsPopover page={props.viewResponse.page} defaultView={props.defaultView} />
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
          uuid={props.viewResponse.uuid}
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
          filters={props.viewResponse.view.filters}
          uuid={props.viewResponse.uuid}
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
          uuid={props.viewResponse.uuid}
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
            uuid={props.viewResponse.uuid}
          />
        </SectionContainer>
      )}
    </Sidebar>
  );
}
