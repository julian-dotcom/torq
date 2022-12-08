import { Link } from "react-router-dom";
import { Options20Regular as OptionsIcon } from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import TimeIntervalSelect from "features/timeIntervalSelect/TimeIntervalSelect";
import {
  AllForwardsColumns,
  DefaultForwardsView,
  ForwardsFilterTemplate,
  ForwardsSortByTemplate,
} from "./forwardsDefaults";
import useTranslations from "services/i18n/useTranslations";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { selectForwardsView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { selectTimeInterval } from "features/timeIntervalSelect/timeIntervalSlice";
import { addDays, format } from "date-fns";
import { useGetForwardsQuery } from "apiSlice";
import { Forward } from "./forwardsTypes";
import { forwardsCellRenderer } from "./forwardsCells";
import Table from "features/table/Table";
import { useFilterData, useSortData } from "../viewManagement/hooks";
// import Button, { buttonColor } from "components/buttons/Button";

function ForwardsPage() {
  const { t } = useTranslations();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();

  const { viewResponse, selectedViewIndex } = useAppSelector(selectForwardsView);
  const currentPeriod = useAppSelector(selectTimeInterval);
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(addDays(new Date(currentPeriod.to), 1), "yyyy-MM-dd");

  const forwardsResponse = useGetForwardsQuery<{
    data: Array<Forward>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ from: from, to: to }, { skip: !isSuccess });

  const filteredData = useFilterData(forwardsResponse.data, viewResponse.view.filters);
  const sortedData = useSortData(filteredData, viewResponse.view.sortBy);

  // Apply frontend based filters
  // TODO: Move this to a custom reach hook, e.g. useFilteredData
  // const data = viewResponse.view.filters
  //   ? applyFilters(deserialiseQuery(viewResponse.view.filters), forwardsResponse.data || [])
  //   : forwardsResponse.data;

  // Logic for toggling the sidebar
  const [sidebarExpanded, setSidebarExpanded] = useState(false);

  const closeSidebarHandler = () => {
    return () => {
      setSidebarExpanded(false);
    };
  };

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
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      selectedViewIndex={selectedViewIndex}
      allColumns={AllForwardsColumns}
      defaultView={DefaultForwardsView}
      filterableColumns={AllForwardsColumns}
      filterTemplate={ForwardsFilterTemplate}
      sortableColumns={AllForwardsColumns}
      sortByTemplate={ForwardsSortByTemplate}
      enableGroupBy={true}
    />
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
        activeColumns={viewResponse.view.columns || []}
        data={sortedData}
        cellRenderer={forwardsCellRenderer}
        isLoading={forwardsResponse.isLoading || forwardsResponse.isFetching || forwardsResponse.isUninitialized}
        showTotals={true}
      />
      {/*<ForwardsDataWrapper viewResponse={viewResponse} loadingViews={!isSuccess} />*/}
    </TablePageTemplate>
  );
}

export default ForwardsPage;
