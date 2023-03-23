import {
  Options20Regular as OptionsIcon,
  ArrowDownload20Regular as DownloadCsvIcon,
  ArrowSync20Regular as RefreshIcon,
} from "@fluentui/react-icons";
import mixpanel from "mixpanel-browser";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { Peer } from "features/peers/peersTypes";
import useTranslations from "services/i18n/useTranslations";
import Table from "features/table/Table";
import {
  PeersFilterTemplate,
  PeersSortTemplate,
  FilterablePeersColumns,
  SortablePeersColumns,
  DefaultPeersView,
} from "features/peers/peersDefaults";
import { AllPeersColumns } from "features/peers/peersColumns.generated";
import { useGetPeersQuery } from "apiSlice";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery, useUpdateTableViewMutation } from "features/viewManagement/viewsApiSlice";
import { selectPeersViews, selectViews } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { useState } from "react";
import { useFilterData, useSortData } from "features/viewManagement/hooks";
import { selectActiveNetwork } from "features/network/networkSlice";
import { TableResponses, ViewResponse } from "../viewManagement/types";
import { createCsvFile } from "utils/JsonTableToCsv";
import Button, { ColorVariant } from "components/buttons/Button";
import peerCellRenderer from "./peersCellRenderer";

function PeersPage() {
  const { t } = useTranslations();
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectPeersViews);
  const peersView = useAppSelector(selectViews)("peers");
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const [updateTableView] = useUpdateTableViewMutation();

  const peersResponse = useGetPeersQuery<{
    data: Array<Peer>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork }, { skip: !isSuccess, pollingInterval: 10000 });

  const filteredData = useFilterData(peersResponse.data, viewResponse.view.filters);
  const data = useSortData(filteredData, viewResponse.view.sortBy);

  // Logic for toggling the sidebar
  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
    mixpanel.track("Toggle Table Sidebar", { page: "Peers" });
  };

  function handleNameChange(name: string) {
    const view = peersView.views[selectedViewIndex] as ViewResponse<TableResponses>;
    if (view.id) {
      updateTableView({
        id: view.id,
        view: { ...view.view, title: name },
      });
    }
  }

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup></TableControlsTabsGroup>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
        <Button
          buttonColor={ColorVariant.primary}
          title={t.download}
          hideMobileText={true}
          icon={<DownloadCsvIcon />}
          onClick={() => {
            mixpanel.track("Downloads Table as CSV", {
              downloadTablePage: "Peers",
              downloadTableViewTitle: viewResponse.view?.title,
              downloadTableColumns: viewResponse.view?.columns,
              downloadTableFilters: viewResponse.view?.filters,
              downloadTableSortBy: viewResponse.view?.sortBy,
            });
            createCsvFile(data, viewResponse.view.title || "Peers");
          }}
        />
        <Button
          buttonColor={ColorVariant.primary}
          icon={<RefreshIcon />}
          onClick={() => {
            mixpanel.track("Refresh Table", { page: "Peers" });
            peersResponse.refetch();
          }}
        />
        <TableControlsButton
          onClickHandler={() => {
            mixpanel.track("Toggle Table Sidebar", { page: "Peers" });
            setSidebarExpanded(!sidebarExpanded);
          }}
          icon={OptionsIcon}
          id={"tableControlsButton"}
        />
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  const sidebar = (
    <ViewsSidebar
      onExpandToggle={closeSidebarHandler}
      expanded={sidebarExpanded}
      viewResponse={viewResponse}
      selectedViewIndex={selectedViewIndex}
      allColumns={AllPeersColumns}
      defaultView={DefaultPeersView}
      filterableColumns={FilterablePeersColumns}
      filterTemplate={PeersFilterTemplate}
      sortableColumns={SortablePeersColumns}
      sortByTemplate={PeersSortTemplate}
    />
  );

  const breadcrumbs = [<span key="b1">{t.peers}</span>];

  return (
    <TablePageTemplate
      title={viewResponse.view.title}
      titleContent={""}
      breadcrumbs={breadcrumbs}
      sidebarExpanded={sidebarExpanded}
      sidebar={sidebar}
      tableControls={tableControls}
      onNameChange={handleNameChange}
      isDraft={viewResponse.id === undefined}
    >
      <Table
        cellRenderer={peerCellRenderer}
        data={data}
        activeColumns={viewResponse.view.columns || []}
        isLoading={peersResponse.isLoading || peersResponse.isFetching || peersResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default PeersPage;
