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
import { useGetTableViewsQuery } from "../viewManagement/viewsApiSlice";
import { selectForwardsView } from "../viewManagement/viewSlice";
import ViewsSidebar from "../viewManagement/ViewsSidebar";
import ForwardsDataWrapper from "./ForwardsDataWrapper";
// import Button, { buttonColor } from "components/buttons/Button";

function ForwardsPage() {
  const { t } = useTranslations();

  const { isSuccess } = useGetTableViewsQuery<{ isSuccess: boolean }>();

  const { viewResponse, selectedViewIndex } = useAppSelector(selectForwardsView);
  console.log("viewResponse", viewResponse);
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
      <ForwardsDataWrapper viewResponse={viewResponse} loadingViews={!isSuccess} />
    </TablePageTemplate>
  );
}

export default ForwardsPage;
