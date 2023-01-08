import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";

export type TagCellProps = {
  cellWrapperClassName?: string;
  totalCell?: boolean;
  children?: React.ReactNode;
};

const CellWrapper = ({ cellWrapperClassName, totalCell, children }: TagCellProps) => {
  return (
    <div>
      <div className={classNames(cellStyles.cell, { [cellStyles.totalCell]: totalCell }, cellWrapperClassName)}>
        {children}
      </div>
    </div>
  );
};

export default CellWrapper;
