import React, { useContext, useEffect, useState } from "react";
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
import { InputSizeVariant, RadioChips, Select } from "components/forms/forms";
import Spinny from "features/spinny/Spinny";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";

type SelectOptions = {
  label?: string;
  value: number | string;
};

type TagProps = Omit<WorkflowNodeProps, "colorVariant">;

export function RemoveTagNode({ ...wrapperProps }: TagProps) {
  const { t } = useTranslations();

  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;
  const toastRef = React.useContext(ToastContext);

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
    removedTags: SelectedTag[];
  };
  const applyToChannelId = "channels-" + wrapperProps.workflowVersionNodeId;
  const applyToNodesId = "nodes-" + wrapperProps.workflowVersionNodeId;

  const [appliesTo, setAppliesTo] = useState(wrapperProps.parameters.applyTo || "channel");
  const [selectedRemovedTags, setSelectedRemovedtags] = useState<SelectedTag[]>(
    (wrapperProps.parameters as TagParameters).removedTags
  );

  function handleRemovedTagChange(newValue: unknown) {
    setSelectedRemovedtags(newValue as SelectedTag[]);
  }

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    if (
      ((wrapperProps.parameters as TagParameters).removedTags || [])
        .map((t) => t.value)
        .sort()
        .join("") !==
        (selectedRemovedTags || [])
          .map((t) => t.value)
          .sort()
          .join("") ||
      appliesTo !== wrapperProps.parameters?.applyTo
    ) {
      setDirty(true);
    } else {
      setDirty(false);
    }
  }, [appliesTo, selectedRemovedTags, wrapperProps.parameters]);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: {
        applyTo: appliesTo,
        removedTags: selectedRemovedTags,
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
          editingDisabled={editingDisabled}
        />
        <RadioChips
          label={t.ApplyTo}
          sizeVariant={InputSizeVariant.small}
          groupName={"node-channels-switch-" + wrapperProps.workflowVersionNodeId}
          options={[
            {
              label: t.channels,
              id: applyToChannelId,
              checked: appliesTo === "channel",
              onChange: () => setAppliesTo("channel"),
            },
            {
              label: t.nodes,
              id: applyToNodesId,
              checked: appliesTo === "nodes",
              onChange: () => setAppliesTo("nodes"),
            },
          ]}
          editingDisabled={editingDisabled}
        />
        <Select
          isMulti={true}
          options={tagsOptions}
          onChange={handleRemovedTagChange}
          label={t.workflowNodes.removeTag}
          sizeVariant={InputSizeVariant.small}
          value={selectedRemovedTags}
          isDisabled={editingDisabled}
        />
        <Button
          type="submit"
          buttonColor={ColorVariant.success}
          buttonSize={SizeVariant.small}
          icon={!processing ? <SaveIcon /> : <Spinny />}
          disabled={!dirty || processing || editingDisabled}
        >
          {!processing ? t.save.toString() : t.saving.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
