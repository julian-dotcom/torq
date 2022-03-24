import "./table_controls.scss";
import DefaultButton from "../buttons/Button";
import {
  ColumnTriple20Regular as ColumnsIcon,
  ArrowSortDown20Regular as SortIcon,
  Filter20Filled as FilterIcon,
} from "@fluentui/react-icons";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import TableHeader from "./TableHeader";
import Dropdown from "../formElements/Dropdown";

function TableControls() {
  return (
    <div className="table-controls">
      <div className="left-container">
        {/*<TableHeader title="Top Revenue Today"/>*/}
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
