import { useEffect, useState } from "react";
import { Tag20Regular as TagIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { useGetTagsQuery } from "pages/tags/tagsApi";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useSelector } from "react-redux";
import { Tag } from "pages/tags/tagsTypes";
import { InputSizeVariant, RadioChips, Select } from "components/forms/forms";
import Spinny from "features/spinny/Spinny";

type SelectOptions = {
  label?: string;
  value: number | string;
};

type TagProps = Omit<WorkflowNodeProps, "colorVariant">;

export function AddTagNode({ ...wrapperProps }: TagProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  const { data: tagsResponse } = useGetTagsQuery<{
    data: Array<Tag>;
    isLoading: boolean;
    isFetching: boolean;
    isUninitialized: boolean;
    isSuccess: boolean;
  }>();

  let tagsOptions: SelectOptions[] = [];
  if (tagsResponse?.length !== undefined) {
    tagsOptions = tagsResponse.map((tag) => {
      return {
        value: tag?.tagId ? tag?.tagId : 0,
        label: tag.name,
      };
    });
  }

  type SelectedTag = {
    value: number;
    label: string;
  };

  type TagParameters = {
    addedTags: SelectedTag[];
  };
  const applyToChannelId = "channels-" + wrapperProps.workflowVersionNodeId;
  const applyToNodesId = "nodes-" + wrapperProps.workflowVersionNodeId;

  const [applyTo, setApplyTo] = useState(applyToChannelId);
  const [selectedAddedTags, setSelectedAddedtags] = useState<SelectedTag[]>(
    (wrapperProps.parameters as TagParameters).addedTags
  );

  function handleAddedTagChange(newValue: unknown) {
    setSelectedAddedtags(newValue as SelectedTag[]);
  }

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    const appliesTo = applyTo === applyToNodesId ? "nodes" : "channel";
    if (
      ((wrapperProps.parameters as TagParameters).addedTags || [])
        .map((t) => t.value)
        .sort()
        .join("") !==
        (selectedAddedTags || [])
          .map((t) => t.value)
          .sort()
          .join("") ||
      appliesTo !== wrapperProps.parameters?.applyTo
    ) {
      setDirty(true);
    } else {
      setDirty(false);
    }
  }, [applyTo, selectedAddedTags, wrapperProps.parameters]);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setProcessing(true);
    const appliesTo = applyTo === applyToNodesId ? "nodes" : "channel";
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: {
        applyTo: appliesTo,
        addedTags: selectedAddedTags,
      },
    }).finally(() => {
      setProcessing(false);
    });
  }

  const { childLinks } = useSelector(
    SelectWorkflowNodeLinks({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeId: wrapperProps.workflowVersionNodeId,
      stage: wrapperProps.stage,
    })
  );

  const parentNodeIds = childLinks?.map((link) => link.parentWorkflowVersionNodeId) ?? [];
  const parentNodes = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: parentNodeIds,
    })
  );

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      heading={t.workflowNodes.tag}
      headerIcon={<TagIcon />}
      colorVariant={NodeColorVariant.accent3}
      outputName={"channels"}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Targets}
          selectedNodes={parentNodes || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"channels"}
        />
        <RadioChips
          label={t.ApplyTo}
          sizeVariant={InputSizeVariant.small}
          groupName={"node-channels-switch-" + wrapperProps.workflowVersionNodeId}
          options={[
            {
              label: t.channels,
              id: applyToChannelId,
              checked: applyTo === applyToChannelId,
              onChange: () => setApplyTo(applyToChannelId),
            },
            {
              label: t.nodes,
              id: applyToNodesId,
              checked: applyTo === applyToNodesId,
              onChange: () => setApplyTo(applyToNodesId),
            },
          ]}
        />
        <Select
          isMulti={true}
          options={tagsOptions}
          onChange={handleAddedTagChange}
          label={t.workflowNodes.addTag}
          sizeVariant={InputSizeVariant.small}
          value={selectedAddedTags}
        />
        <Button
          type="submit"
          buttonColor={ColorVariant.success}
          buttonSize={SizeVariant.small}
          icon={!processing ? <SaveIcon /> : <Spinny />}
          disabled={!dirty || processing}
        >
          {!processing ? t.save.toString() : t.saving.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
