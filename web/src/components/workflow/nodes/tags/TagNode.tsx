import { useState } from "react";
import { Tag20Regular as TagIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { useGetTagsQuery } from "pages/tags/tagsApi";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "../nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useSelector } from "react-redux";
import { Tag } from "pages/tags/tagsTypes";
import { Select } from "components/forms/forms";

type SelectOptions = {
  label?: string;
  value: number | string;
};

type TagProps = Omit<WorkflowNodeProps, "colorVariant">;

export function TagNode({ ...wrapperProps }: TagProps) {
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
    removedTags: SelectedTag[];
  };

  const [selectedAddedTags, setSelectedAddedtags] = useState<SelectedTag[]>(
    (wrapperProps.parameters as TagParameters).addedTags
  );
  const [selectedRemovedTags, setSelectedRemovedtags] = useState<SelectedTag[]>(
    (wrapperProps.parameters as TagParameters).removedTags
  );

  function handleAddedTagChange(newValue: unknown) {
    setSelectedAddedtags(newValue as SelectedTag[]);
  }
  function handleRemovedTagChange(newValue: unknown) {
    setSelectedRemovedtags(newValue as SelectedTag[]);
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: {
        addedTags: selectedAddedTags,
        removedTags: selectedRemovedTags,
      },
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
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.workflowNodes.targetChannel}
          selectedNodes={parentNodes || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputIndex={1}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.workflowNodes.targetNode}
          selectedNodes={parentNodes || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputIndex={2}
        />
        <Select
          isMulti={true}
          options={tagsOptions}
          onChange={handleAddedTagChange}
          label={t.workflowNodes.addTag}
          value={selectedAddedTags}
        />
        <Select
          isMulti={true}
          options={tagsOptions}
          onChange={handleRemovedTagChange}
          label={t.workflowNodes.removeTag}
          value={selectedRemovedTags}
        />
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
