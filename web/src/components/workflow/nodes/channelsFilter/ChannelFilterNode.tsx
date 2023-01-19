import { useState } from "react";
import { Filter20Regular as FilterIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "../nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useSelector } from "react-redux";
import FilterComponent from "features/sidebar/sections/filter/FilterComponent";
import { AndClause, deserialiseQuery, OrClause } from "features/sidebar/sections/filter/filter";
import { AllChannelsColumns, ChannelsFilterTemplate } from "features/channels/channelsDefaults"

type FilterChannelsNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function ChannelFilterNode({ ...wrapperProps }: FilterChannelsNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();
  wrapperProps.parameters = {"$and":[]}

  const [filterState, setFilterState] = useState(deserialiseQuery(wrapperProps.parameters) as AndClause | OrClause);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: filterState as unknown as Record<string, unknown>,
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

  const handleFilterUpdate = (filter: AndClause | OrClause) => {
    setFilterState(filter);
  };

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      heading={t.channelPolicyConfiguration}
      headerIcon={<FilterIcon />}
      colorVariant={NodeColorVariant.accent1}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.inputs}
          selectedNodes={parentNodes || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputIndex={1}
        />
        <FilterComponent
          filters={filterState}
          columns={AllChannelsColumns}
          defaultFilter={ChannelsFilterTemplate}
          child={false}
          onFilterUpdate={handleFilterUpdate}
        />
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
