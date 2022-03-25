import "./table_controls.scss";
import DefaultButton from "../buttons/Button";
import { useSelector, useDispatch } from "react-redux";
import {
  ColumnTripleRegular as ColumnsIcon,
  ArrowSortDownLines16Regular as SortIcon,
  Filter16Regular as FilterIcon,
  NavigationRegular as NavigationIcon,
  ArrowJoinRegular as GroupIcon
} from "@fluentui/react-icons";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import Dropdown from "../formElements/Dropdown";

function TableControls() {

  const dispatch = useDispatch();

  const toggleNav = () => {
    dispatch({type: 'toggleNav'})
  }

  return (
    <div className="table-controls">
      <div className="left-container">
        <div className="upper-container">
          <DefaultButton icon={<NavigationIcon/>} text={"Menu"} onClick={toggleNav} className={"show-nav-btn"}/>
          <Dropdown/>
        </div>
        <div className="lower-container">
          <DefaultButton icon={<ColumnsIcon/>} text={"Columns"}/>
          <DefaultButton icon={<SortIcon/>} text={"Sort"}/>
          <DefaultButton icon={<FilterIcon/>} text={"Filter"}/>
          <DefaultButton icon={<GroupIcon/>} text={"Group"}/>
        </div>
      </div>
      <div className="right-container">
        <TimeIntervalSelect/>
      </div>
    </div>
  );
}

export default TableControls;
