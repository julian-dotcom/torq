import "./table_controls.scss";
import {
  Navigation20Regular as NavigationIcon,
  ArrowJoin20Regular as GroupIcon,
  // Search20Regular as SearchIcon,
  Save20Regular as SaveIcon,
} from "@fluentui/react-icons";


import TimeIntervalSelect from "../../timeIntervalSelect/TimeIntervalSelect";
import DefaultButton from "../../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import { toggleNav } from "../../navigation/navSlice";
import SortControls from "./sort/SortControls";
import {
  createTableViewAsync, fetchChannelsAsync,
  selectCurrentView,
  selectedViewIndex,
  updateTableViewAsync
} from "../tableSlice";
import FilterPopover from "./filter/FilterPopover";

import ViewsPopover from "./views/ViewsPopover";
import ColumnsPopover from "./columns/ColumnsPopover";
import {useEffect} from "react";
import {selectTimeInterval} from "../../timeIntervalSelect/timeIntervalSlice";
import {format} from "date-fns";

function TableControls() {
  const dispatch = useAppDispatch();
  const currentView = useAppSelector(selectCurrentView);
  const currentViewIndex = useAppSelector(selectedViewIndex);
  const currentPeriod = useAppSelector(selectTimeInterval);

    useEffect(() => {
      const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
      const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
      dispatch(fetchChannelsAsync({ from: from, to: to }));
    }, [currentPeriod])

  const saveView = () => {
    let viewMod = { ...currentView }
    viewMod.saved = true
    if (currentView.id === undefined || null) {
      dispatch(createTableViewAsync({ view: viewMod, index: currentViewIndex }))
      return
    }
    dispatch(updateTableViewAsync({ view: viewMod, index: currentViewIndex }))
  }
  return (
    <div className="table-controls">
      <div className="left-container">
        <div className="upper-container">
          <DefaultButton
            icon={<NavigationIcon />}
            text={"Menu"}
            onClick={() => dispatch(toggleNav())}
            className={"show-nav-btn collapse-tablet"}
          />
          <ViewsPopover />
          {!currentView.saved && (<DefaultButton
            icon={<SaveIcon />}
            text={"Save"}
            onClick={saveView}
            className={"collapse-tablet danger"}
          />)}
        </div>
        <div className="lower-container">
          <ColumnsPopover />
          <FilterPopover />
          <SortControls />
          <DefaultButton
            icon={<GroupIcon />}
            text={"Group"}
            className={"collapse-tablet"}
          />
          {/*<DefaultButton*/}
          {/*  icon={<SearchIcon />}*/}
          {/*  text={"Search"}*/}
          {/*  className={"small-tablet"}*/}
          {/*/>*/}
        </div>
      </div>
      <div className="right-container">
        <TimeIntervalSelect />
      </div>
    </div>
  );
}

export default TableControls;
