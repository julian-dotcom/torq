import {
  Options20Regular as OptionsIcon,
  ArrowDownload20Regular as DownloadCsvIcon,
  ArrowSync20Regular as RefreshIcon,
  Add20Regular as NewPeerIcon,
} from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButtonGroup,
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
import { useGetPeersQuery } from "features/peers/peersApi";
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
import * as Routes from "constants/routes";
import { useNavigate } from "react-router-dom";
import { useLocation } from "react-router";
import { userEvents } from "utils/userEvents";

function PeersPage() {
  const { t } = useTranslations();
  const { track } = userEvents();
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const activeNetwork = useAppSelector(selectActiveNetwork);
  const [updateTableView] = useUpdateTableViewMutation();
  const navigate = useNavigate();
  const location = useLocation();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectPeersViews);
  const peersView = useAppSelector(selectViews)("peers");

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
    track("Toggle Table Sidebar", { page: "Peers" });
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
    <TableControlSection intercomTarget={"table-page-controls"}>
      <TableControlsButtonGroup intercomTarget={"table-page-controls-left"}>
        <Button
          intercomTarget={"new-peer-button"}
          buttonColor={ColorVariant.success}
          hideMobileText={true}
          icon={<NewPeerIcon />}
          onClick={() => {
            track("Navigate to Connect Peer");
            navigate(Routes.CONNECT_PEER, { state: { background: location } });
          }}
        >
          {t.peersPage.connectPeer}
        </Button>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup intercomTarget={"table-page-controls-right"}>
        <Button
          intercomTarget={"download-csv"}
          buttonColor={ColorVariant.primary}
          title={t.download}
          hideMobileText={true}
          icon={<DownloadCsvIcon />}
          onClick={() => {
            track("Downloads Table as CSV", {
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
          intercomTarget={"refresh-table"}
          buttonColor={ColorVariant.primary}
          icon={<RefreshIcon />}
          onClick={() => {
            track("Refresh Table", { page: "Peers" });
            peersResponse.refetch();
          }}
        />
        <Button
          intercomTarget={"table-settings"}
          onClick={() => {
            track("Toggle Table Sidebar", { page: "Peers" });
            setSidebarExpanded(!sidebarExpanded);
          }}
          icon={<OptionsIcon />}
          id={"tableControlsButton"}
        >
          {t.options}
        </Button>
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

  const breadcrumbs = [t.manage, <span key="b1">{t.peers}</span>];

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
        intercomTarget={"peers-table"}
        cellRenderer={peerCellRenderer}
        data={data}
        activeColumns={viewResponse.view.columns || []}
        isLoading={peersResponse.isLoading || peersResponse.isFetching || peersResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default PeersPage;
