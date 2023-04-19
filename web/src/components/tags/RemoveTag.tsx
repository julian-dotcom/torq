import classNames from "classnames";
import styles from "./tag.module.scss";
import { Delete16Regular as RemoveIcon } from "@fluentui/react-icons";
import { TagProps, TagColor, TagSize, TagSizeClasses, TagColorClasses } from "./Tag";

export default function Tag(props: TagProps) {
  const sizeClass = TagSizeClasses.get(props.sizeVariant || TagSize.small);
  const colorClass = TagColorClasses.get(props.colorVariant || TagColor.primary);
  const customColor = {
    backgroundColor: props.customBackgroundColor,
    color: props.customTextColor,
  };
  return (
    <div className={styles.removeTagWrapper}>
      <div className={classNames(styles.tagWrapper, colorClass, sizeClass)} style={customColor}>
        <div className={styles.tagLabel}>{props.label}</div>
      </div>
      <div className={classNames(styles.iconWrapper, styles.tagWrapper, sizeClass)} style={customColor}>
        <div className={classNames(styles.icon)}>
          <RemoveIcon />
        </div>
      </div>
    </div>
  );
}
