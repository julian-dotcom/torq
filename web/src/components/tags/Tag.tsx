import classNames from "classnames";
import styles from "./tag.module.scss";
import { LockClosed16Regular as LockedIcon, Delete16Regular as DeleteIcon } from "@fluentui/react-icons";
import { useUntagChannelMutation, useUntagNodeMutation } from "pages/tags/tagsApi";
import { userEvents } from "utils/userEvents";

export enum TagSize {
  normal = "normal",
  small = "small",
  tiny = "tiny",
}

export const TagSizeClasses = new Map<TagSize, string>([
  [TagSize.normal, styles.normal],
  [TagSize.small, styles.small],
  [TagSize.tiny, styles.tiny],
]);

export enum TagColor {
  primary = "primary",
  success = "success",
  warning = "warning",
  error = "error",
  accent1 = "accent1",
  accent2 = "accent2",
  accent3 = "accent3",
  custom = "custom",
}

export const TagColorClasses = new Map<TagColor, string>([
  [TagColor.primary, styles.primary],
  [TagColor.success, styles.success],
  [TagColor.warning, styles.warning],
  [TagColor.error, styles.error],
  [TagColor.accent1, styles.accent1],
  [TagColor.accent2, styles.accent2],
  [TagColor.accent3, styles.accent3],
  [TagColor.custom, styles.custom],
]);

export type BaseTagProps = {
  sizeVariant?: TagSize;
  colorVariant?: TagColor;
  customBackgroundColor?: string;
  customTextColor?: string;
  locked?: boolean;
  label: string;
  icon?: JSX.Element;
  tagId?: number;
  peerId?: never;
  deletable?: never;
  channelId?: never;
};

export type PeerTagProps = {
  sizeVariant?: TagSize;
  colorVariant?: TagColor;
  customBackgroundColor?: string;
  customTextColor?: string;
  locked?: boolean;
  label: string;
  icon?: JSX.Element;
  deletable: boolean;
  tagId: number;
  peerId: number;
  channelId?: never;
};

export type ChannelTagProps = {
  sizeVariant?: TagSize;
  colorVariant?: TagColor;
  customBackgroundColor?: string;
  customTextColor?: string;
  locked?: boolean;
  label: string;
  icon?: JSX.Element;
  deletable: boolean;
  tagId: number;
  peerId?: never;
  channelId?: number;
};

export type TagProps = BaseTagProps | PeerTagProps | ChannelTagProps;

export default function Tag(props: TagProps) {
  const { track } = userEvents();
  const sizeClass = TagSizeClasses.get(props.sizeVariant || TagSize.small);
  const colorClass = TagColorClasses.get(props.colorVariant || TagColor.primary);
  const customColor = {
    backgroundColor: props.customBackgroundColor,
    color: props.customTextColor,
  };

  const [untagChannel] = useUntagChannelMutation();
  const [untagNode] = useUntagNodeMutation();

  const handleUntagPeer = () => {
    if (!props.peerId || !props.tagId) return;
    track("Untag", {
      tagId: props?.tagId,
      tagName: props.label,
      peerId: props?.peerId,
      tagType: "node",
    });
    untagNode({ tagId: props.tagId, nodeId: props.peerId });
  };

  const handleUntagChannel = () => {
    if (!props.channelId || !props.tagId) return;
    track("Untag", {
      tagId: props?.tagId,
      tagName: props.label,
      channelId: props?.channelId,
      tagType: "channel",
    });
    untagChannel({ tagId: props.tagId, channelId: props.channelId });
  };

  // Untag peer or channel
  const handleUntag = () => {
    if (props.peerId) {
      handleUntagPeer();
    } else if (props.channelId) {
      handleUntagChannel();
    }
  };

  return (
    <div
      className={classNames(styles.tagWrapper, colorClass, sizeClass, { [styles.deletable]: props.deletable })}
      style={customColor}
    >
      {props.locked && (
        <div className={classNames(styles.icon)}>
          <LockedIcon />
        </div>
      )}
      {props.icon && <div className={classNames(styles.tagIcon)}>{props.icon}</div>}
      <div className={styles.tagLabel}>{props.label}</div>
      {props.deletable && (
        <div className={classNames(styles.deleteIcon)} onClick={handleUntag}>
          {<DeleteIcon />}
        </div>
      )}
    </div>
  );
}
