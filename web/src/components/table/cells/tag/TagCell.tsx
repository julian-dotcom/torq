import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./tag_cell.module.scss";
import Tag, { TagProps } from "components/tags/Tag";

export type TagCellProps = TagProps & {
  cellWrapperClassName?: string;
  totalCell?: boolean;
};

const CheckboxCell = ({ cellWrapperClassName, totalCell, ...tagProps }: TagCellProps) => {
  return (
    <div
      className={classNames(
        cellStyles.cell,
        styles.tagCell,
        { [cellStyles.totalCell]: totalCell },
        cellWrapperClassName
      )}
    >
      <Tag {...tagProps} />
    </div>
  );
};

export default CheckboxCell;
