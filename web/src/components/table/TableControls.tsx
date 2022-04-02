import "./table_controls.scss";
import {
  Navigation20Regular as NavigationIcon,
  ArrowJoin20Regular as GroupIcon,
  Search20Regular as SearchIcon,
  Options20Regular as OptionsIcon,
  Save20Regular as SaveIcon,
} from "@fluentui/react-icons";
import { format } from "date-fns";

import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import DefaultButton from "../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import { toggleNav } from "../navigation/navSlice";
import SortControls from "./controls/sort/SortControls";
import {
  createTableViewAsync,
  selectCurrentView,
  selectedViewIndex, updateTableViewAsync
} from "./tableSlice";
import FilterPopover from "./controls/filter/FilterPopover";

import ViewsPopover from "./controls/views/ViewsPopover";
import ColumnsPopover from "./controls/columns/ColumnsPopover";

function TableControls() {
  const dispatch = useAppDispatch();

  const currentView = useAppSelector(selectCurrentView);
  const currentViewIndex = useAppSelector(selectedViewIndex);

  const saveView = () => {
    let viewMod = {...currentView}
    viewMod.saved = true
    if (currentView.id === undefined || null) {
      dispatch(createTableViewAsync({view: viewMod, index: currentViewIndex}))
      return
    }
    dispatch(updateTableViewAsync({view: viewMod, index: currentViewIndex}))
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
          <DefaultButton
            icon={<OptionsIcon />}
            text={""}
            className={"collapse-tablet mobile-options"}
          />
          {!currentView.saved && (<DefaultButton
            icon={<SaveIcon/>}
            text={"Save"}
            onClick={saveView}
            className={"danger"}
          />)}
        </div>
        <div className="lower-container">
          <ColumnsPopover />
          <SortControls />
          <FilterPopover />
          {/*<DefaultButton*/}
          {/*  icon={<GroupIcon />}*/}
          {/*  text={"Group"}*/}
          {/*  className={"collapse-tablet"}*/}
          {/*/>*/}
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
