import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./tag_cell.module.scss";
import Tag, { TagProps } from "components/tags/Tag";
import { Link, useLocation } from "react-router-dom";

export type TagCellProps = TagProps & {
  cellWrapperClassName?: string;
  totalCell?: boolean;
  tagId?: number;
  editLink?: true;
};

const TagCell = ({ cellWrapperClassName, totalCell, editLink, ...tagProps }: TagCellProps) => {
  const location = useLocation();

  function EditLinkWrapper() {
    if (editLink) {
      return (
        <Link
          to={`/update-tag/${tagProps.tagId}`}
          state={{ background: location }}
          className={classNames(cellStyles.action, styles.updateLink)}
        >
          <Tag {...tagProps} />
        </Link>
      );
    } else {
      return <Tag {...tagProps} />;
    }
  }

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
        <EditLinkWrapper />
      </div>
    </div>
  );
};

export default TagCell;
