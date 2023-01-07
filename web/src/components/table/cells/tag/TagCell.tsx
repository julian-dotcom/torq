import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./tag_cell.module.scss";
import Tag, { TagProps } from "components/tags/Tag";
import { Link, useLocation } from "react-router-dom";
import { UPDATE_TAG } from "constants/routes";

export type TagCellProps = TagProps & {
  cellWrapperClassName?: string;
  totalCell?: boolean;
  tagId?: number;
};

const CheckboxCell = ({ cellWrapperClassName, totalCell, ...tagProps }: TagCellProps) => {
  const location = useLocation();
  return (
    <div>
      <div
        className={classNames(
          cellStyles.cell,
          styles.tagCell,
          { [cellStyles.totalCell]: totalCell },
          cellWrapperClassName
        )}
      >
        <Link
          to={`${UPDATE_TAG}?tagId=${tagProps.tagId}`}
          state={{ background: location }}
          className={classNames(cellStyles.action, styles.updateLink)}
        >
          <Tag {...tagProps} />
        </Link>
      </div>
    </div>
  );
};

export default CheckboxCell;
