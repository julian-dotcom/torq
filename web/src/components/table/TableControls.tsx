import "./table_controls.scss";
import {
  ColumnTriple20Regular as ColumnsIcon,
  Navigation20Regular as NavigationIcon,
  ArrowJoin20Regular as GroupIcon,
  Search20Regular as SearchIcon,
  Options20Regular as OptionsIcon
} from "@fluentui/react-icons";
import { format } from "date-fns";

import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import DefaultButton from "../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import { toggleNav } from "../navigation/navSlice";
import SortControls from "./controls/sort/SortControls";
import { fetchChannelsAsync } from "./tableSlice";
import FilterPopover from "./controls/filter/FilterPopover";
import { selectTimeInterval } from "../timeIntervalSelect/timeIntervalSlice";

import ViewsPopover from "./controls/views/ViewsPopover";
import ColumnsPopover from "./controls/columns/ColumnsPopover";

function TableControls() {
  const dispatch = useAppDispatch();
  const currentPeriod = useAppSelector(selectTimeInterval);
  const from = format(new Date(currentPeriod.from), "yyyy-MM-dd");
  const to = format(new Date(currentPeriod.to), "yyyy-MM-dd");
  dispatch(fetchChannelsAsync({ from: from, to: to }));

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
