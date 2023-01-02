import { useState } from "react";
import { MoneySettings20Regular as ChannelPolicyIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "../nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { NumberFormatValues } from "react-number-format";
import { useSelector } from "react-redux";

type FilterChannelsNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

type channelPolicyConfigurationNode = {
  feeRate: number | undefined;
  baseFee: number | undefined;
  minHTLCAmount: number | undefined;
  maxHTLCAmount: number | undefined;
};

export function ChannelFilterNode<T>({ ...wrapperProps }: FilterChannelsNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  const [channelPolicy, setChannelPolicy] = useState<channelPolicyConfigurationNode>({
    feeRate: undefined,
    baseFee: undefined,
    minHTLCAmount: undefined,
    maxHTLCAmount: undefined,
    ...wrapperProps.parameters,
  });

  function createChangeHandler(key: keyof channelPolicyConfigurationNode) {
    return (e: NumberFormatValues) => {
      setChannelPolicy((prev) => ({
        ...prev,
        [key]: e.floatValue,
      }));
    };
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: channelPolicy,
    });
  }

  const nodeLinks = useSelector(
    SelectWorkflowNodeLinks({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeId: wrapperProps.workflowVersionNodeId,
      childLinks: true,
    })
  );

  const parentNodeIds = nodeLinks?.map((link) => link.parentWorkflowVersionNodeId) ?? [];
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
      heading={t.channelPolicyConfiguration}
      headerIcon={<ChannelPolicyIcon />}
      colorVariant={NodeColorVariant.accent1}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          label={t.inputs}
          selectedNodes={parentNodes || []}
          id={"sss"}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputIndex={1}
        />
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
