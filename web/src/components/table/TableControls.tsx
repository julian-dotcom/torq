import "./table_controls.scss";
import DefaultButton from "../buttons/Button";
import { useSelector, useDispatch } from "react-redux";
import {
  ColumnTriple20Regular as ColumnsIcon,
  ArrowSortDown20Regular as SortIcon,
  Filter20Filled as FilterIcon,
  Navigation20Regular as NavigationIcon
} from "@fluentui/react-icons";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import Dropdown from "../formElements/Dropdown";

function TableControls() {

  const dispatch = useDispatch();
  // const navHidden: number = useSelector((state:{navHidden:number}) => {return state.navHidden});

  const toggleNav = () => {
    dispatch({type: 'toggleNav'})
  }

  return (
    <div className="table-controls">
      <div className="left-container">
        <DefaultButton icon={<NavigationIcon/>} text={"Menu"} onClick={toggleNav} className={"show-nav-btn"}/>
        <Dropdown/>
        <DefaultButton icon={<ColumnsIcon/>} text={"Columns"}/>
        <DefaultButton icon={<SortIcon/>} text={"Sort"}/>
        <DefaultButton icon={<FilterIcon/>} text={"Filter"}/>
      </div>
      <div className="right-container">
        <TimeIntervalSelect/>
      </div>
    </div>
  );
}

export default TableControls;
