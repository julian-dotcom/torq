import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./tag_cell.module.scss";
import Tag, { TagProps } from "components/tags/Tag";
import {Link, useLocation } from "react-router-dom";
import { UPDATE_TAG } from "constants/routes";
import { useDeleteTagMutation } from "pages/tagsPage/tagsApi"
import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";

export type TagCellProps = TagProps & {
  cellWrapperClassName?: string;
  totalCell?: boolean;
  tagId?: number;
};

const CheckboxCell = ({ cellWrapperClassName, totalCell, ...tagProps }: TagCellProps) => {
  const [deleteTagMutation, deleteTagResponse] = useDeleteTagMutation();
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
        {tagProps?.tagId && tagProps?.tagId < 0 && (
          <Link to={`${UPDATE_TAG}?tagId=${tagProps.tagId}`} state={{ background: location }} className={classNames(cellStyles.action, styles.updateLink)}>
            <Tag {...tagProps} />
          </Link>
        )}
        {tagProps?.tagId && tagProps?.tagId > 0 && (
          <div className={classNames(styles.tagLinks)}>
            <Link to={`${UPDATE_TAG}?tagId=${tagProps.tagId}`} state={{ background: location }} className={classNames(cellStyles.action, styles.updateLink)}>
              <Tag {...tagProps} />
            </Link>
            <ButtonWrapper
              className={styles.customButtonWrapperStyles}
              rightChildren={
              <Button className={styles.tagbutton}
                onClick={() => {
                  deleteTagMutation(tagProps?.tagId || 0)
                }}
                buttonColor={ColorVariant.error}>
                Delete
              </Button>
              }
            />
          </div>
        )}
        </div>

    </div>
  );
};

export default CheckboxCell;
