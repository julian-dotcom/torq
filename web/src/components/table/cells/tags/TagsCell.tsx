import mixpanel from "mixpanel-browser";
import { MoleculeRegular as NodeIcon, ArrowRoutingRegular as ChannelsIcon } from "@fluentui/react-icons";
import { Link, useLocation } from "react-router-dom";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./tags_cell.module.scss";
import { Tag as TagType } from "pages/tags/tagsTypes";
import Tag, { TagSize } from "components/tags/Tag";
import { ColorVariant, LinkButton, SizeVariant } from "components/buttons/Button";
import classNames from "classnames";
import { userEvents } from "utils/userEvents";

export type TagsCellProps = {
  channelId?: number;
  nodeId: number;
  channelTags: TagType[];
  peerTags: TagType[];
  displayChannelTags: boolean;
  totalCell?: boolean;
};

function EditTag(props: { tag: TagType }) {
  const location = useLocation();
  const { track } = userEvents();
  return (
    <Link
      to={`/update-tag/${props.tag.tagId}`}
      state={{ background: location }}
      className={classNames(cellStyles.action, styles.updateLink)}
      onClick={() => {
        track("Navigate to Update Tag", {
          tagId: props.tag.tagId,
          tagName: props.tag.name,
          tagStyle: props.tag.style,
          tagCategory: props.tag.categoryName,
        });
      }}
    >
      <Tag label={props.tag.name} colorVariant={props.tag.style} sizeVariant={TagSize.tiny} />
    </Link>
  );
}

const TagsCell = (props: TagsCellProps) => {
  const location = useLocation();

  let channelTags = props.channelTags;
  if (!props.displayChannelTags) {
    channelTags = [];
  }
  const allTags: TagType[] = [];

  // peer and channel might have the same tag, de-dupe so we don't show it twice
  for (const item of [...channelTags, ...props.peerTags]) {
    if (!allTags.some((t) => t.tagId === item.tagId)) {
      allTags.push(item);
    }
  }

  return (
    <div className={classNames(cellStyles.cell, styles.tagCell, { [cellStyles.totalCell]: props.totalCell })}>
      {!props.totalCell && allTags.map((tag) => <EditTag tag={tag} key={"tag-" + tag.tagId} />)}

      {!props.totalCell && (
        <>
          {props.displayChannelTags && (
            <LinkButton
              intercomTarget={"tag-cell-add-channel-tag-button"}
              to={`/tag-channel/${props.channelId}`}
              state={{ background: location }}
              icon={<ChannelsIcon />}
              title="Tag Channel"
              buttonSize={SizeVariant.tiny}
              buttonColor={ColorVariant.disabled}
              onClick={() => {
                mixpanel.track("Navigate to Tag Channel", {
                  channelId: props.channelId,
                });
              }}
            />
          )}
          <LinkButton
            intercomTarget={"tag-cell-add-node-tag-button"}
            to={`/tag-node/${props.nodeId}`}
            state={{ background: location }}
            icon={<NodeIcon />}
            title="Tag Peer"
            buttonSize={SizeVariant.tiny}
            buttonColor={ColorVariant.disabled}
            onClick={() => {
              mixpanel.track("Navigate to Tag Node", {
                nodeId: props.nodeId,
              });
            }}
          />
        </>
      )}
    </div>
  );
};

export default TagsCell;
