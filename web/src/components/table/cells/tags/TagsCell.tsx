import { ArrowRoutingRegular as ChannelsIcon, MoleculeRegular as NodeIcon } from "@fluentui/react-icons";
import { useLocation } from "react-router-dom";
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

const TagsCell = (props: TagsCellProps) => {
  const location = useLocation();
  const { track } = userEvents();

  return (
    <div className={classNames(cellStyles.cell, styles.tagCell, { [cellStyles.totalCell]: props.totalCell })}>
      {!props.totalCell &&
        (props.peerTags || []).map((tag) => (
          <Tag
            key={"peer-tag-" + tag.tagId}
            label={tag.name}
            colorVariant={tag.style}
            sizeVariant={TagSize.tiny}
            deletable={true}
            tagId={tag.tagId || 1}
            peerId={props.nodeId}
            icon={<NodeIcon />}
          />
        ))}

      {!props.totalCell &&
        props.displayChannelTags &&
        (props.channelTags || []).map((tag) => (
          <Tag
            key={"channel-tag-" + tag.tagId}
            label={tag.name}
            colorVariant={tag.style}
            sizeVariant={TagSize.tiny}
            deletable={true}
            tagId={tag.tagId || 1}
            channelId={props.channelId}
            icon={<ChannelsIcon />}
          />
        ))}

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
                track("Navigate to Tag Channel", {
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
              track("Navigate to Tag Node", {
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
