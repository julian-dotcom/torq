import { Link } from "react-router-dom";
import {
  ArrowJoin20Regular as GroupIcon,
  ArrowSortDownLines20Regular as SortIcon,
  ColumnTriple20Regular as ColumnsIcon,
  Filter20Regular as FilterIcon,
  // Save20Regular as SaveIcon,
  Options20Regular as OptionsIcon,
} from "@fluentui/react-icons";
import Sidebar from "features/sidebar/Sidebar";
import TablePageTemplate, {
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
// import ViewsPopover from "features/viewManagement/ViewsPopover";
import ColumnsSection from "features/sidebar/sections/columns/ColumnsSection";
import FilterSection from "features/sidebar/sections/filter/FilterSection";
import SortSection from "features/sidebar/sections/sort/SortSection";
import GroupBySection from "features/sidebar/sections/group/GroupBySection";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import { useView } from "../viewManagement/useView";
import {
  AllForwardsColumns,
  DefaultForwardsView,
  ForwardsFilterTemplate,
  ForwardsSortByTemplate,
} from "./forwardsDefaults";
import { SectionContainer } from "features/section/SectionContainer";
import useTranslations from "services/i18n/useTranslations";
import { forwardsCellRenderer } from "./forwardsCells";
import { useAppSelector } from "store/hooks";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import { useGetForwardsQuery } from "apiSlice";
import Table from "features/table/Table";
import { Forward } from "./forwardsTypes";
import ViewsPopover from "../viewManagement/ViewsPopover";
// import Button, { buttonColor } from "components/buttons/Button";

type sections = {
  filter: boolean;
  sort: boolean;
  group: boolean;
  columns: boolean;
};
function ForwardsPage() {
  const { t } = useTranslations();
  const [view, selectView, isViewsLoaded, allViews] = useView("forwards", AllForwardsColumns, 0, DefaultForwardsView);

  const currentPeriod = useAppSelector(selectTimeInterval);
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");

  const forwardsResponse = useGetForwardsQuery<{
    data: Array<Forward>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ from: from, to: to }, { skip: !isViewsLoaded });

  // useEffect(() => {
  //   const views: ViewInterface<ForwardResponse>[] = [];
  //   if (forwardsViews) {
  //     forwardsViews?.map((v: ViewResponse<ForwardResponse>) => {
  //       views.push(v.view);
  //     });
  //
  //     dispatch(updateViews({ views, index: 0 }));
  //   } else {
  //     dispatch(updateViews({ views: [{ ...DefaultView, title: "Default View" }], index: 0 }));
  //   }
  // }, [forwardsViews, isLoading]);

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

  // const currentView = useAppSelector(selectCurrentView);
  // const saveView = () => {
  //   const viewMod = { ...currentView };
  //   viewMod.saved = true;
  //   if (currentView.id === undefined || null) {
  //     createTableView({ view: viewMod, index: currentViewIndex, page: "forwards" });
  //     return;
  //   }
  //   updateTableView(viewMod);
  // };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          {/*<ViewsPopover views={views} />*/}
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
        <TableControlsButton onClickHandler={() => setSidebarExpanded(!sidebarExpanded)} icon={OptionsIcon} />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const sidebar = (
    <Sidebar title={"Options"} closeSidebarHandler={closeSidebarHandler()}>
      <ViewsPopover
        views={allViews}
        page={"forwards"}
        selectedView={0}
        onSelectView={selectView}
        ViewTemplate={DefaultForwardsView}
      />
      <SectionContainer
        title={"Columns"}
        icon={ColumnsIcon}
        expanded={activeSidebarSections.columns}
        handleToggle={sidebarSectionHandler("columns")}
      >
        <ColumnsSection columns={AllForwardsColumns} view={view} />
      </SectionContainer>
      <SectionContainer
        title={"Filter"}
        icon={FilterIcon}
        expanded={activeSidebarSections.filter}
        handleToggle={sidebarSectionHandler("filter")}
      >
        <FilterSection columns={AllForwardsColumns} view={view} defaultFilter={ForwardsFilterTemplate} />
      </SectionContainer>
      <SectionContainer
        title={"Sort"}
        icon={SortIcon}
        expanded={activeSidebarSections.sort}
        handleToggle={sidebarSectionHandler("sort")}
      >
        <SortSection columns={AllForwardsColumns} view={view} defaultSortBy={ForwardsSortByTemplate} />
      </SectionContainer>
      <SectionContainer
        title={t.group}
        icon={GroupIcon}
        expanded={activeSidebarSections.group}
        handleToggle={sidebarSectionHandler("group")}
      >
        <GroupBySection view={view} />
      </SectionContainer>
    </Sidebar>
  );

  const breadcrumbs = [
    <span key="b1">Analyse</span>,
    <Link key="b2" to={"/analyse/forwards"}>
      Forwards
    </Link>,
  ];

  return (
    <TablePageTemplate
      title={"Forwards"}
      titleContent={<TimeIntervalSelect />}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
    >
      <Table
        activeColumns={view.columns}
        data={forwardsResponse?.data || []}
        cellRenderer={forwardsCellRenderer}
        isLoading={forwardsResponse.isLoading || forwardsResponse.isFetching || forwardsResponse.isUninitialized}
        showTotals={true}
      />
      {/*<ForwardsDataWrapper selectedView={0} />*/}
    </TablePageTemplate>
  );
}

export default ForwardsPage;
