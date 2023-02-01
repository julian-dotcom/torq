import { Link } from "react-router-dom";
import { Options20Regular as OptionsIcon, ArrowRouting20Regular as ChannelsIcon } from "@fluentui/react-icons";
import mixpanel from "mixpanel-browser";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import { useNavigate } from "react-router-dom";
import { useLocation } from "react-router";
import { channel } from "./channelsTypes";
import { OPEN_CHANNEL } from "constants/routes";
import useTranslations from "services/i18n/useTranslations";
import Table from "features/table/Table";
import {
  ChannelsFilterTemplate,
  ChannelsSortTemplate,
  DefaultChannelsView,
  SortableChannelsColumns,
} from "features/channels/channelsDefaults";
import { AllChannelsColumns } from "features/channels/channelsColumns";
import { useGetChannelsQuery } from "apiSlice";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { selectChannelView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";
import { useState } from "react";
import { useFilterData, useSortData } from "features/viewManagement/hooks";
import { useGroupBy } from "features/sidebar/sections/group/groupBy";
import channelsCellRenderer from "./channelsCellRenderer";
import { selectActiveNetwork } from "features/network/networkSlice";

function useMaximums(data: Array<channel>): channel | undefined {
  if (!data.length) {
    return undefined;
  }

  return data.reduce((prev: channel, current: channel) => {
    return {
      ...prev,
      alias: "Max",
      feeBase: Math.max(prev.feeBase, current.feeBase),
      capacity: Math.max(prev.capacity, current.capacity),
      commitFee: Math.max(prev.commitFee, current.commitFee),
      commitmentType: Math.max(prev.commitmentType, current.commitmentType),
      commitWeight: Math.max(prev.commitWeight, current.commitWeight),
      feePerKw: Math.max(prev.feePerKw, current.feePerKw),
      feeRateMilliMsat: Math.max(prev.feeRateMilliMsat, current.feeRateMilliMsat),
      fundingOutputIndex: Math.max(prev.fundingOutputIndex, current.fundingOutputIndex),
      gauge: Math.max(prev.gauge, current.gauge),
      lifetime: Math.max(prev.lifetime, current.lifetime),
      lndShortChannelId: Math.max(prev.lndShortChannelId, current.lndShortChannelId),
      balance: Math.max(prev.balance, current.balance), // NB! This column only exists in the frontend!
      localBalance: Math.max(prev.localBalance, current.localBalance),
      localChanReserveSat: Math.max(prev.localChanReserveSat, current.localChanReserveSat),
      maxHtlc: Math.max(prev.maxHtlc, current.maxHtlc),
      minHtlc: Math.max(prev.minHtlc, current.minHtlc),
      nodeId: Math.max(prev.nodeId, current.nodeId),
      channelId: Math.max(prev.channelId, current.channelId),
      numUpdates: Math.max(prev.numUpdates, current.numUpdates),
      pendingForwardingHTLCsAmount: Math.max(prev.pendingForwardingHTLCsAmount, current.pendingForwardingHTLCsAmount),
      pendingForwardingHTLCsCount: Math.max(prev.pendingForwardingHTLCsCount, current.pendingForwardingHTLCsCount),
      pendingLocalHTLCsAmount: Math.max(prev.pendingLocalHTLCsAmount, current.pendingLocalHTLCsAmount),
      pendingLocalHTLCsCount: Math.max(prev.pendingLocalHTLCsCount, current.pendingLocalHTLCsCount),
      pendingTotalHTLCsAmount: Math.max(prev.pendingTotalHTLCsAmount, current.pendingTotalHTLCsAmount),
      pendingTotalHTLCsCount: Math.max(prev.pendingTotalHTLCsCount, current.pendingTotalHTLCsCount),
      remoteBalance: Math.max(prev.remoteBalance, current.remoteBalance),
      remoteFeeBase: Math.max(prev.remoteFeeBase, current.remoteFeeBase),
      remoteChanReserveSat: Math.max(prev.remoteChanReserveSat, current.remoteChanReserveSat),
      remoteFeeRateMilliMsat: Math.max(prev.remoteFeeRateMilliMsat, current.remoteFeeRateMilliMsat),
      remoteMaxHtlc: Math.max(prev.remoteMaxHtlc, current.remoteMaxHtlc),
      remoteMinHtlc: Math.max(prev.remoteMinHtlc, current.remoteMinHtlc),
      remotePubkey: Math.max(prev.remotePubkey, current.remotePubkey),
      remoteTimeLockDelta: Math.max(prev.remoteTimeLockDelta, current.remoteTimeLockDelta),
      timeLockDelta: Math.max(prev.timeLockDelta, current.timeLockDelta),
      totalSatoshisReceived: Math.max(prev.totalSatoshisReceived, current.totalSatoshisReceived),
      totalSatoshisSent: Math.max(prev.totalSatoshisSent, current.totalSatoshisSent),
      unsettledBalance: Math.max(prev.unsettledBalance, current.unsettledBalance),
    };
  });
}

function ChannelsPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectChannelView);
  const activeNetwork = useAppSelector(selectActiveNetwork);

  const channelsResponse = useGetChannelsQuery<{
    data: Array<channel>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>({ network: activeNetwork }, { skip: !isSuccess });

  const filteredData = useFilterData(channelsResponse.data, viewResponse.view.filters);
  const sortedData = useSortData(filteredData, viewResponse.view.sortBy);
  const data = useGroupBy<channel>(sortedData, viewResponse.view.groupBy);
  const maxRow = useMaximums(data);

  // Logic for toggling the sidebar
  const closeSidebarHandler = () => {
    setSidebarExpanded(false);
    mixpanel.track("Toggle Table Sidebar", { page: "Channels" });
  };

  const tableControls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<ChannelsIcon />}
            onClick={() => {
              mixpanel.track("Navigate to Open Channel");
              navigate(OPEN_CHANNEL, { state: { background: location } });
            }}
          >
            {t.openChannel}
          </Button>
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
        <TableControlsButton
          onClickHandler={() => {
            mixpanel.track("Toggle Table Sidebar", { page: "Channels" });
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
      allColumns={AllChannelsColumns}
      defaultView={DefaultChannelsView}
      filterableColumns={AllChannelsColumns}
      filterTemplate={ChannelsFilterTemplate}
      sortableColumns={SortableChannelsColumns}
      sortByTemplate={ChannelsSortTemplate}
    />
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
      <Table
        cellRenderer={channelsCellRenderer}
        data={data}
        activeColumns={viewResponse.view.columns || []}
        isLoading={channelsResponse.isLoading || channelsResponse.isFetching || channelsResponse.isUninitialized}
        maxRow={maxRow}
      />
    </TablePageTemplate>
  );
}

export default ChannelsPage;
