import { ReactNode, useEffect, useState } from "react";
import {
  ArrowRouting20Regular as ChannelsIcon,
  Molecule20Regular as NodesIcon,
  TargetArrow24Regular as TargetIcon,
} from "@fluentui/react-icons";
import styles from "./targets_section.module.scss";
import TargetsHeader from "components/targets/TargetsHeader";
import Collapse from "features/collapse/Collapse";
import useTranslations from "services/i18n/useTranslations";
import { ChannelNode, Corridor, CorridorFields } from "pages/tagsPage/tagsTypes";
import {
  useAddChannelsGroupsMutation,
  useDeleteChannelGroupByTagMutation,
  useGetCorridorByReferenceQuery,
  useGetNodesChannelsQuery,
} from "pages/tagsPage/tagsApi";
import { Select } from "components/forms/forms";
import Target from "components/targets/Target";

export type SelectOptions = {
  label?: string;
  value: number | string;
  type: "node" | "channel" | "group";
};

type TargetsSectionProps = {
  tagId: number;
};

export default function TargetsSection(props: TargetsSectionProps) {
  const { t } = useTranslations();
  const [collapsedNode, setCollapsedNode] = useState<boolean>(false);
  const [collapsedChannel, setCollapsedChannel] = useState<boolean>(false);
  const [targetNodes, setTargetNodes] = useState<ReactNode[]>([]);
  const [targetChannels, setTargetChannels] = useState<ReactNode[]>([]);
  const [addChannelsGroupsMutation] = useAddChannelsGroupsMutation();
  const [deleteChannelGroupByTagMutation] = useDeleteChannelGroupByTagMutation();

  function handleDeleteTarget(corridorId: number) {
    deleteChannelGroupByTagMutation(corridorId);
  }

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
        addChannelsGroupsMutation({
          tagId: props.tagId,
          nodeId: channelsOption.nodeId,
          channelId: value,
        });
      }
    } else {
      const nodesOption = channelsNodesResponse.nodes.find((option) => option.nodeId === value);
      if (nodesOption !== undefined) {
        addChannelsGroupsMutation({
          tagId: props.tagId,
          nodeId: value,
        });
      }
    }
  }

  const { data: corridorsResponse } = useGetCorridorByReferenceQuery<{
    data: Corridor;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>(props.tagId);

  useEffect(() => {
    if (corridorsResponse) {
      const listNodes: ReactNode[] = [];
      const listChannels: ReactNode[] = [];
      corridorsResponse.corridors?.map((c: CorridorFields) => {
        if (c.shortChannelId) {
          listChannels.push(
            <Target
              onDeleteTarget={() => handleDeleteTarget(c.corridorId)}
              key={"channel-target-" + c.corridorId}
              icon={<ChannelsIcon />}
              details={c.alias}
              title={c.shortChannelId}
            />
          );
        } else {
          listNodes.push(
            <Target
              onDeleteTarget={() => handleDeleteTarget(c.corridorId)}
              key={"channel-target-" + c.corridorId}
              icon={<NodesIcon />}
              details={"-"}
              title={c.alias}
            />
          );
        }
      });
      setTargetNodes(listNodes);
      setTargetChannels(listChannels);
    }
  }, [corridorsResponse]);

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
            addTarget(selectOptions?.value as number, selectOptions?.type as string);
          }}
          options={nodesOptions}
        />
        <div className={styles.collapsGroup}>
          <TargetsHeader
            title={`${targetNodes.length} ${t.nodes}`}
            icon={<NodesIcon />}
            expanded={!collapsedNode}
            onCollapse={() => setCollapsedNode(!collapsedNode)}
          />
          <Collapse collapsed={collapsedNode} animate={true}>
            <div className={styles.targetsWrapper}>
              {!targetNodes.length && (
                <div className={styles.noTargets}>
                  <span className={styles.noTargetsText}>{t.tagsModal.noNodesSelected}</span>
                </div>
              )}
              {!!targetNodes.length && targetNodes}
            </div>
          </Collapse>
        </div>
      </div>

      <div className={styles.addTagWrapper}>
        <Select
          label={t.tagChannel}
          onChange={(newValue: unknown) => {
            const selectOptions = newValue as SelectOptions;
            addTarget(selectOptions?.value as number, selectOptions?.type as string);
          }}
          options={channelsOptions}
        />
        <div className={styles.collapsGroup}>
          <TargetsHeader
            title={`${targetChannels.length} ${t.channels}`}
            icon={<ChannelsIcon />}
            expanded={!collapsedChannel}
            onCollapse={() => setCollapsedChannel(!collapsedChannel)}
          />
          <Collapse collapsed={collapsedChannel} animate={true}>
            <div className={styles.targetsWrapper}>
              {!targetChannels.length && (
                <div className={styles.noTargets}>
                  <span className={styles.noTargetsText}>{t.tagsModal.noChannelsSelected}</span>
                </div>
              )}
              {!!targetChannels.length && targetChannels}
            </div>
          </Collapse>
        </div>
      </div>
    </div>
  );
}
