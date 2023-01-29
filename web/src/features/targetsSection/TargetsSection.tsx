import { useState } from "react";
import mixpanel from "mixpanel-browser";
import {
  ArrowRouting20Regular as ChannelsIcon,
  Molecule20Regular as NodesIcon,
  TargetArrow24Regular as TargetIcon,
} from "@fluentui/react-icons";
import styles from "./targets_section.module.scss";
import TargetsHeader from "components/targets/TargetsHeader";
import Collapse from "features/collapse/Collapse";
import useTranslations from "services/i18n/useTranslations";
import { ChannelNode, TaggedChannels, TaggedNodes } from "pages/tags/tagsTypes";
import {
  useGetNodesChannelsQuery,
  useTagChannelMutation,
  useTagNodeMutation,
  useUntagChannelMutation,
  useUntagNodeMutation,
} from "pages/tags/tagsApi";
import { Select } from "components/forms/forms";
import Target from "components/targets/Target";

export type SelectOptions = {
  label?: string;
  value: number | string;
  type: "node" | "channel" | "group";
};

type TargetsSectionProps = {
  tagId: number;
  tagName: string;
  channels: TaggedChannels[];
  nodes: TaggedNodes[];
};

export default function TargetsSection(props: TargetsSectionProps) {
  const { t } = useTranslations();
  const [collapsedNode, setCollapsedNode] = useState<boolean>(false);
  const [collapsedChannel, setCollapsedChannel] = useState<boolean>(false);
  const [tagChannel] = useTagChannelMutation();
  const [tagNode] = useTagNodeMutation();
  const [untagChannel] = useUntagChannelMutation();
  const [untagNode] = useUntagNodeMutation();

  const { data: channelsNodesResponse } = useGetNodesChannelsQuery<{
    data: ChannelNode;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();

  let channelsOptions: SelectOptions[] = [];
  let nodesOptions: SelectOptions[] = [];
  if (channelsNodesResponse?.channels?.length !== undefined) {
    channelsOptions = channelsNodesResponse.channels.map((channel) => {
      return {
        value: channel.channelId,
        label: `${channel.shortChannelId?.toString()} | ${channel?.alias}`,
        type: "channel",
      };
    });
  }
  if (channelsNodesResponse?.nodes?.length !== undefined) {
    nodesOptions = channelsNodesResponse.nodes.map((node) => {
      return { value: node.nodeId, label: node.alias, type: "node" };
    });
  }

  function addTarget(value: number, type: string) {
    if (type === "channel") {
      const channelsOption = channelsNodesResponse.channels.find((option) => option.channelId === value);
      if (channelsOption !== undefined) {
        tagChannel({
          tagId: props.tagId,
          channelId: value,
        });
      }
    } else {
      const nodesOption = channelsNodesResponse.nodes.find((option) => option.nodeId === value);
      if (nodesOption !== undefined) {
        tagNode({
          tagId: props.tagId,
          nodeId: value,
        });
      }
    }
  }

  return (
    <div className={styles.targetsSection}>
      <div className={styles.target}>
        <div className={styles.targetIcon}>
          <TargetIcon />
        </div>
        <div className={styles.targetText}>Applied to</div>
      </div>

      <div className={styles.addTagWrapper}>
        <Select
          label={t.tagNode}
          onChange={(newValue: unknown) => {
            const selectOptions = newValue as SelectOptions;
            mixpanel.track("Tag", {
              tagId: props.tagId,
              tagName: props.tagName,
              nodeId: selectOptions?.value,
              type: "node",
            });
            addTarget(selectOptions?.value as number, selectOptions?.type as string);
          }}
          options={nodesOptions}
        />
        <div className={styles.collapsGroup}>
          <TargetsHeader
            title={`${props.nodes?.length || 0} ${t.nodes}`}
            icon={<NodesIcon />}
            expanded={!collapsedNode}
            onCollapse={() => setCollapsedNode(!collapsedNode)}
          />
          <Collapse collapsed={collapsedNode} animate={true}>
            <div className={styles.targetsWrapper}>
              {!props.nodes?.length && (
                <div className={styles.noTargets}>
                  <span className={styles.noTargetsText}>{t.tagsModal.noNodesSelected}</span>
                </div>
              )}
              {props.nodes?.map((c) => {
                return (
                  <Target
                    onDeleteTarget={() => {
                      mixpanel.track("Untag", {
                        tagId: props.tagId,
                        tagName: props.tagName,
                        nodeId: c.nodeId,
                        type: "node",
                      });
                      untagNode({ tagId: props.tagId, nodeId: c.nodeId });
                    }}
                    key={"node-target-" + c.nodeId}
                    icon={<NodesIcon />}
                    details={"Channels " + c.channelCount}
                    title={c.name}
                  />
                );
              })}
            </div>
          </Collapse>
        </div>
      </div>

      <div className={styles.addTagWrapper}>
        <Select
          label={t.tagChannel}
          onChange={(newValue: unknown) => {
            const selectOptions = newValue as SelectOptions;
            mixpanel.track("Tag", {
              tagId: props.tagId,
              tagName: props.tagName,
              channelId: selectOptions?.value,
              type: "channel",
            });
            addTarget(selectOptions?.value as number, selectOptions?.type as string);
          }}
          options={channelsOptions}
        />
        <div className={styles.collapsGroup}>
          <TargetsHeader
            title={`${props.channels?.length || 0} ${t.channels}`}
            icon={<ChannelsIcon />}
            expanded={!collapsedChannel}
            onCollapse={() => setCollapsedChannel(!collapsedChannel)}
          />
          <Collapse collapsed={collapsedChannel} animate={true}>
            <div className={styles.targetsWrapper}>
              {!props.channels?.length && (
                <div className={styles.noTargets}>
                  <span className={styles.noTargetsText}>{t.tagsModal.noChannelsSelected}</span>
                </div>
              )}
              {props.channels?.map((c) => {
                return (
                  <Target
                    onDeleteTarget={() => {
                      mixpanel.track("Untag", {
                        tagId: props.tagId,
                        tagName: props.tagName,
                        channelId: c.channelId,
                        type: "channel",
                      });
                      untagChannel({ tagId: props.tagId, channelId: c.channelId });
                    }}
                    key={"channel-target-" + c.channelId}
                    icon={<ChannelsIcon />}
                    details={c.name}
                    title={c.shortChannelId}
                  />
                );
              })}
            </div>
          </Collapse>
        </div>
      </div>
    </div>
  );
}
