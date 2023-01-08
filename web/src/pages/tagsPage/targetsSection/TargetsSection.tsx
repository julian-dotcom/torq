import { ReactNode, useState } from "react";
import {
  ArrowRouting20Regular as ChannelsIcon,
  Molecule20Regular as NodesIcon,
  TargetArrow20Regular as TargetIcon,
} from "@fluentui/react-icons";
import styles from "./targets_section.module.scss";
import TargetsHeader from "components/targets/TargetsHeader";
import Collapse from "features/collapse/Collapse";
import useTranslations from "services/i18n/useTranslations";
import { ChannelForTag, ChannelNode, NodeForTag } from "pages/tagsPage/tagsTypes";
import { useGetNodesChannelsQuery } from "pages/tagsPage/tagsApi";
import { Select } from "components/forms/forms";
import { ActionMeta } from "react-select";

export type SelectOptions = {
  label?: string;
  value: number | string;
};

type TargetsSectionProps = {
  tagId: number;
  categoryId?: number;
  nodes: ReactNode[];
  channels: ReactNode[];
};

export default function TargetsSection(props: TargetsSectionProps) {
  const { t } = useTranslations();
  const [collapsedNode, setCollapsedNode] = useState<boolean>(true);
  const [collapsedChannel, setCollapsedChannel] = useState<boolean>(true);
  // const [addChannelsGroupsMutation] = useAddChannelsGroupsMutation();

  const { data: channelsNodesResponse } = useGetNodesChannelsQuery<{
    data: ChannelNode;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();

  let channelsNodesOptions: SelectOptions[] = [{ value: 0, label: "Channel or Node" }];
  let channelsOptions: SelectOptions[] = [];
  let nodesOptions: SelectOptions[] = [];
  if (channelsNodesResponse !== undefined) {
    channelsOptions = channelsNodesResponse.nodes.map((nodesOptions: NodeForTag) => {
      return { value: nodesOptions.nodeId, label: nodesOptions.alias, type: "node" };
    });
    nodesOptions = channelsNodesResponse.channels.map((channelsOptions: ChannelForTag) => {
      return { value: channelsOptions.channelId, label: channelsOptions.shortChannelId, type: "channel" };
    });
  }
  channelsNodesOptions = nodesOptions.concat(channelsOptions);

  // function addTarget(value: number, type: string) {
  //   let request: ChannelGroup = {
  //     tagId: props.tagId,
  //   };
  //   addChannelsGroupsMutation(request);
  //   // setTarget(channelsNodesOptions[0].value as number);
  // }

  // function handlePostTagInsert() {
  //   // const channelGoupObj: ChannelGroup = {
  //   //   tagId,
  //   //   nodeId,
  //   // };
  //   // if (channelId && channelId > 0) {
  //   //   channelGoupObj.channelId = channelId;
  //   // }
  //   // if (selectedTagCategory > 0) {
  //   //   channelGoupObj.categoryId = selectedTagCategory;
  //   // }
  //   // addChannelsGroupsMutation(channelGoupObj);
  // }

  return (
    <div className={styles.targetsSection}>
      <div className={styles.target}>
        <TargetIcon className={styles.targetIcon} /> <span className={styles.targetText}>Applied to</span>
      </div>

      <div className={styles.openChannelTableRow}>
        <Select
          label={t.tagsModal.target}
          onChange={(newValue: unknown, _: ActionMeta<unknown>) => {
            const selectOptions = newValue as SelectOptions;
            console.log(selectOptions);
            // addTarget(selectOptions?.value as number, selectOptions?.type as string);
          }}
          options={channelsNodesOptions}
        />
      </div>

      <div className={styles.collapsGroup}>
        <TargetsHeader
          title={`${props.nodes.length} ${t.nodes}`}
          icon={<NodesIcon />}
          expanded={!collapsedNode}
          onCollapse={() => setCollapsedNode(!collapsedNode)}
        />
        <Collapse collapsed={collapsedChannel} animate={true}>
          <div className={styles.targetsWrapper}>
            {!props.nodes.length && (
              <div className={styles.noTargets}>
                <span className={styles.noTargetsText}>{t.tagsModal.noNodesSelected}</span>
              </div>
            )}
            {!!props.channels.length && props.nodes}
          </div>
        </Collapse>
      </div>

      <div className={styles.collapsGroup}>
        <TargetsHeader
          title={`${props.channels.length} ${t.channels}`}
          icon={<ChannelsIcon />}
          expanded={!collapsedChannel}
          onCollapse={() => setCollapsedChannel(!collapsedChannel)}
        />
        <Collapse collapsed={collapsedChannel} animate={true}>
          <div className={styles.targetsWrapper}>
            {!props.channels.length && (
              <div className={styles.noTargets}>
                <span className={styles.noTargetsText}>{t.tagsModal.noChannelsSelected}</span>
              </div>
            )}
            {!!props.channels.length && props.channels}
          </div>
        </Collapse>
      </div>
    </div>
  );
}
