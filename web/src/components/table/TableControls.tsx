import "./table_controls.scss";
import DefaultButton from "../buttons/Button";
import {useAppDispatch } from '../../store/hooks';
import {toggleNav} from '../navigation/navSlice'
import {
  ColumnTriple20Regular as ColumnsIcon,
  ArrowSortDownLines20Regular as SortIcon,
  Filter20Regular as FilterIcon,
  Navigation20Regular as NavigationIcon,
  ArrowJoin20Regular as GroupIcon,
  Search20Regular as SearchIcon,
  Options20Regular as OptionsIcon,
} from "@fluentui/react-icons";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import Dropdown from "../formElements/Dropdown";

function TableControls() {

  const dispatch = useAppDispatch()

  return (
    <div className="table-controls">
      <div className="left-container">
        <div className="upper-container">
          <DefaultButton
            icon={<NavigationIcon/>}
            text={"Menu"}
            onClick={() => dispatch(toggleNav())}
            className={"show-nav-btn collapse-tablet"}/>
          <Dropdown/>
          <DefaultButton icon={<OptionsIcon/>} text={""} className={"collapse-tablet mobile-options"}/>
        </div>
        <div className="lower-container">
          <DefaultButton icon={<ColumnsIcon/>} text={"Columns"} className={"collapse-tablet"}/>
          <DefaultButton icon={<SortIcon/>} text={"Sort"} className={"collapse-tablet"}/>
          <DefaultButton icon={<FilterIcon/>} text={"Filter"} className={"collapse-tablet"}/>
          <DefaultButton icon={<GroupIcon/>} text={"Group"} className={"collapse-tablet"}/>
          <DefaultButton icon={<SearchIcon/>} text={"Search"} className={"small-tablet"}/>
        </div>
      </div>
      <div className="right-container">
        <TimeIntervalSelect/>
      </div>
    </div>
  );
}

export default TableControls;
