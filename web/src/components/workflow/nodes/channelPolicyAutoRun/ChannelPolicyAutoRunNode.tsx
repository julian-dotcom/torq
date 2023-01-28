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

type ChannelPolicyAutoRunNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export type ChannelPolicyConfiguration = {
  feeBaseMsat?: number;
  feeRateMilliMsat?: number;
  maxHtlcMsat?: number;
  minHtlcMsat?: number;
  timeLockDelta?: number;
};

export function ChannelPolicyAutoRunNode({ ...wrapperProps }: ChannelPolicyAutoRunNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  const [channelPolicy, setChannelPolicy] = useState<ChannelPolicyConfiguration>({
    feeBaseMsat: undefined,
    feeRateMilliMsat: undefined,
    maxHtlcMsat: undefined,
    minHtlcMsat: undefined,
    timeLockDelta: undefined,
    ...wrapperProps.parameters,
  });

  const [feeBase, setFeeBase] = useState<number | undefined>(
    (wrapperProps.parameters as ChannelPolicyConfiguration).feeBaseMsat?((wrapperProps.parameters as ChannelPolicyConfiguration).feeBaseMsat || 0) / 1000:undefined
  );
  const [maxHtlc, setMaxHtlc] = useState<number | undefined>(
    (wrapperProps.parameters as ChannelPolicyConfiguration).maxHtlcMsat?((wrapperProps.parameters as ChannelPolicyConfiguration).maxHtlcMsat || 0) / 1000:undefined
  );
  const [minHtlc, setMinHtlc] = useState<number | undefined>(
    (wrapperProps.parameters as ChannelPolicyConfiguration).minHtlcMsat?((wrapperProps.parameters as ChannelPolicyConfiguration).minHtlcMsat || 0) / 1000:undefined
  );

  function createChangeMsatHandler(key: keyof ChannelPolicyConfiguration) {
    return (e: NumberFormatValues) => {
      if (key == "feeBaseMsat") {
        setFeeBase(e.floatValue)
      }
      if (key == "maxHtlcMsat") {
        setMaxHtlc(e.floatValue)
      }
      if (key == "minHtlcMsat") {
        setMinHtlc(e.floatValue)
      }
      if (e.floatValue === undefined) {
        setChannelPolicy((prev) => ({
          ...prev,
          [key]: undefined
        }));
      } else {
        setChannelPolicy((prev) => ({
          ...prev,
          [key]: (e.floatValue || 0) * 1000,
        }));
      }
    };
  }

  function createChangeHandler(key: keyof ChannelPolicyConfiguration) {
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
          value={channelPolicy.feeRateMilliMsat}
          thousandSeparator={","}
          suffix={" ppm"}
          onValueChange={createChangeHandler("feeRateMilliMsat")}
          label={t.updateChannelPolicy.feeRateMilliMsat}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={feeBase}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeHandler("feeRateMilliMsat")}
          label={t.updateChannelPolicy.feeRateMilliMsat}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={feeBase}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeMsatHandler("feeBaseMsat")}
          label={t.baseFee}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={minHtlc}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeMsatHandler("minHtlcMsat")}
          label={t.minHTLCAmount}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
        formatted={true}
        value={maxHtlc}
        thousandSeparator={","}
        suffix={" sat"}
        onValueChange={createChangeMsatHandler("maxHtlcMsat")}
        label={t.maxHTLCAmount}
        sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={channelPolicy.timeLockDelta}
          thousandSeparator={","}
          onValueChange={createChangeHandler("timeLockDelta")}
          label={t.updateChannelPolicy.timeLockDelta}
          sizeVariant={InputSizeVariant.small}
        />
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
