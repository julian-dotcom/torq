import classNames from "classnames";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./tags_cell.module.scss";
import { Link, useLocation } from "react-router-dom";
import { Tag as TagType } from "pages/tagsPage/tagsTypes";
import Tag, { TagSize } from "components/tags/Tag";

export type TagsCellProps = {
  tags: TagType[];
  totalCell?: boolean;
};

function EditLinkWrapper(props: { tag: TagType }) {
  const location = useLocation();

  return (
    <Link
      to={`/update-tag/${props.tag.tagId}`}
      state={{ background: location }}
      className={classNames(cellStyles.action, styles.updateLink)}
    >
      <Tag label={props.tag.name} colorVariant={props.tag.style} sizeVariant={TagSize.tiny} />
    </Link>
  );
}

const TagsCell = ({ tags, totalCell }: TagsCellProps) => {
  return (
    <div>
      <div className={classNames(cellStyles.cell, styles.tagCell, { [cellStyles.totalCell]: totalCell })}>
        {(tags || []).map((tag) => (
          <EditLinkWrapper tag={tag} key={"tag-" + tag.tagId} />
        ))}
      </div>
    </div>
  );
};

export default TagsCell;
