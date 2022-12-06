import { Link } from "react-router-dom";
import {
  MoneySettings20Regular as AdjustFeesIcon,
  Options20Regular as OptionsIcon,
  // Save20Regular as SaveIcon,
  ArrowRouting20Regular as ChannelsIcon,
} from "@fluentui/react-icons";
import TablePageTemplate, {
  TableControlSection,
  TableControlsButton,
  TableControlsButtonGroup,
  TableControlsTabsGroup,
} from "features/templates/tablePageTemplate/TablePageTemplate";
import { useState } from "react";
import Button, { buttonColor } from "components/buttons/Button";
import { useNavigate } from "react-router-dom";
import { useLocation } from "react-router";
import { UPDATE_CHANNEL, OPEN_CHANNEL } from "constants/routes";
import { channel } from "./channelsTypes";
import useTranslations from "services/i18n/useTranslations";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import Table from "features/table/Table";
import {
  AllChannelsColumns,
  ChannelsFilterTemplate,
  ChannelsSortTemplate,
  DefaultChannelsView,
  SortableChannelsColumns,
} from "./channelsDefaults";
import { useGetChannelsQuery } from "apiSlice";
import { useAppSelector } from "store/hooks";
import { useGetTableViewsQuery } from "features/viewManagement/viewsApiSlice";
import { selectChannelView } from "features/viewManagement/viewSlice";
import ViewsSidebar from "features/viewManagement/ViewsSidebar";

function ChannelsPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();
  const { viewResponse, selectedViewIndex } = useAppSelector(selectChannelView);

  const channelsResponse = useGetChannelsQuery<{
    data: Array<channel>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>(undefined, { skip: !isSuccess });

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
      </TableControlsButtonGroup>
      <TableControlsButtonGroup>
        <Button
          buttonColor={buttonColor.green}
          text={t.openChannel}
          className={"collapse-tablet"}
          icon={<ChannelsIcon />}
          onClick={() => {
            navigate(OPEN_CHANNEL, { state: { background: location } });
          }}
        />
        <Button
          buttonColor={buttonColor.green}
          text={t.updateChannelPolicy.title}
          icon={<AdjustFeesIcon />}
          onClick={() => {
            navigate(UPDATE_CHANNEL, { state: { background: location } });
          }}
        />
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
        cellRenderer={DefaultCellRenderer}
        data={channelsResponse?.data || []}
        activeColumns={viewResponse.view.columns || []}
        isLoading={channelsResponse.isLoading || channelsResponse.isFetching || channelsResponse.isUninitialized}
      />
    </TablePageTemplate>
  );
}

export default ChannelsPage;
