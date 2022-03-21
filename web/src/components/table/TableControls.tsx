import "./table_controls.scss";
import DefaultButton from "../buttons/Button";
import {
  ColumnTriple20Regular as ColumnsIcon,
  ArrowSortDown20Regular as SortIcon,
  Filter20Filled as FilterIcon,
} from "@fluentui/react-icons";
import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import TableHeader from "./TableHeader";

function TableControls() {
  return (
    <div className="table-controls">
      <div className="left-container">
        <TableHeader title="Top Revenue Today" />
        <TimeIntervalSelect />
      </div>
      <div className="right-container">
        <DefaultButton icon={<ColumnsIcon />} text={"Columns"} />
        <DefaultButton icon={<SortIcon />} text={"Sort"} />
        <DefaultButton icon={<FilterIcon />} text={"Filter"} />
      </div>
    </div>
  );
}

export default TableControls;
