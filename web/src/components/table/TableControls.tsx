import "./table_controls.scss";
import {
  ColumnTriple20Regular as ColumnsIcon,
  Filter20Regular as FilterIcon,
  Navigation20Regular as NavigationIcon,
  ArrowJoin20Regular as GroupIcon,
  Search20Regular as SearchIcon,
  Options20Regular as OptionsIcon,
} from "@fluentui/react-icons";

import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import Dropdown from "../formElements/Dropdown";
import DefaultButton from "../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import { toggleNav } from "../navigation/navSlice";
import SortControls from "./SortControls";
import FilterPopover from "./filter/FilterPopover";
import { fetchChannelsAsync } from "./tableSlice";
import { selectTimeInterval } from "../timeIntervalSelect/timeIntervalSlice";
import { format } from "date-fns";

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
          <Dropdown />
          <DefaultButton
            icon={<OptionsIcon />}
            text={""}
            className={"collapse-tablet mobile-options"}
          />
        </div>
        <div className="lower-container">
          <DefaultButton
            icon={<ColumnsIcon />}
            text={"Columns"}
            className={"collapse-tablet"}
          />
          <div>
            <SortControls />
          </div>
          <FilterPopover/>
          <DefaultButton
            icon={<GroupIcon />}
            text={"Group"}
            className={"collapse-tablet"}
          />
          <DefaultButton
            icon={<SearchIcon />}
            text={"Search"}
            className={"small-tablet"}
          />
        </div>
      </div>
      <div className="right-container">
        <TimeIntervalSelect />
      </div>
    </div>
  );
}

export default TableControls;
