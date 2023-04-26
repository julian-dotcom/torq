import classNames from "classnames";
import styles from "./tag.module.scss";
import { Delete16Regular as RemoveIcon } from "@fluentui/react-icons";
import { TagProps, TagColor, TagSize, TagSizeClasses, TagColorClasses } from "./Tag";
import { MouseEvent } from "react";

interface clickable {
  onClick: () => void;
}

export default function Tag(props: TagProps & clickable) {
  const sizeClass = TagSizeClasses.get(props.sizeVariant || TagSize.small);
  const colorClass = TagColorClasses.get(props.colorVariant || TagColor.primary);
  const customColor = {
    backgroundColor: props.customBackgroundColor,
    color: props.customTextColor,
  };

  const handleClick = (_: MouseEvent) => {
    props.onClick();
  };
  return (
    <div
      onClick={handleClick}
      className={classNames(styles.removeTag, styles.tagWrapper, colorClass, sizeClass)}
      style={customColor}
    >
      <div className={classNames(styles.icon)}>
        <RemoveIcon />
      </div>
      <div className={styles.tagLabel}>{props.label}</div>
    </div>
  );
}
