import { useState } from "react";
import { MoneySettings20Regular as ChannelPolicyConfiguratorIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Input from "components/forms/input/Input";
import { InputSizeVariant } from "components/forms/input/variants";
import Form from "components/forms/form/Form";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { NumberFormatValues } from "react-number-format";
import { useSelector } from "react-redux";

type ChannelPolicyConfiguratorNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

type channelPolicyConfiguratorNode = {
  feeRate: number | undefined;
  baseFee: number | undefined;
  minHTLCAmount: number | undefined;
  maxHTLCAmount: number | undefined;
};

export function ChannelPolicyConfiguratorNode({ ...wrapperProps }: ChannelPolicyConfiguratorNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  const [channelPolicy, setChannelPolicy] = useState<channelPolicyConfiguratorNode>({
    feeRate: undefined,
    baseFee: undefined,
    minHTLCAmount: undefined,
    maxHTLCAmount: undefined,
    ...wrapperProps.parameters,
  });

  function createChangeHandler(key: keyof channelPolicyConfiguratorNode) {
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
      headerIcon={<ChannelPolicyConfiguratorIcon />}
      colorVariant={NodeColorVariant.accent1}
      outputName={"channels"}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.inputs}
          selectedNodes={parentNodes || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"channels"}
        />
        <Input
          formatted={true}
          value={channelPolicy.feeRate}
          thousandSeparator={","}
          suffix={" ppm"}
          onValueChange={createChangeHandler("feeRate")}
          label={t.feeRate}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={channelPolicy.baseFee}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeHandler("baseFee")}
          label={t.baseFee}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={channelPolicy.minHTLCAmount}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeHandler("minHTLCAmount")}
          label={t.minHTLCAmount}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={channelPolicy.maxHTLCAmount}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeHandler("maxHTLCAmount")}
          label={t.maxHTLCAmount}
          sizeVariant={InputSizeVariant.small}
        />
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
