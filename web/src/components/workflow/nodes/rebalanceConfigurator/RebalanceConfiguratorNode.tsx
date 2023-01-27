import { useState } from "react";
import { ArrowRotateClockwise20Regular as RebalanceConfiguratorIcon, Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes, useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { NumberFormatValues } from "react-number-format";
import { useSelector } from "react-redux";
import { Input, InputSizeVariant, Socket, Form } from "components/forms/forms";

type RebalanceConfiguratorNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export type RebalanceConfiguration = {
  amountMsat: number;
  maximumCostMsat?: number;
  maximumCostMilliMsat?: number;
  maximumConcurrency?: number;
};

export function RebalanceConfiguratorNode({ ...wrapperProps }: RebalanceConfiguratorNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  const [configuration, setConfiguration] = useState<RebalanceConfiguration>({
    amountMsat: 0,
    maximumCostMsat: undefined,
    maximumCostMilliMsat: undefined,
    maximumConcurrency: undefined,
    ...wrapperProps.parameters,
  });

  const [amountSat, setAmountSat] = useState<number | undefined>(
    ((wrapperProps.parameters as RebalanceConfiguration).amountMsat || 0) / 1000
  );
  const [maximumCostSat, setMaximumCostSat] = useState<number | undefined>(
    (wrapperProps.parameters as RebalanceConfiguration).maximumCostMsat?((wrapperProps.parameters as RebalanceConfiguration).maximumCostMsat || 0) / 1000:undefined
  );

  function handleAmountSatChange(e: NumberFormatValues) {
    setAmountSat(e.floatValue);
    setConfiguration((prev) => ({
      ...prev,
      amountMsat: (e.floatValue || 0) * 1000,
    }));
  }

  function handleMaximumCostSatChange(e: NumberFormatValues) {
    setMaximumCostSat(e.floatValue);
    setConfiguration((prev) => ({
      ...prev,
      maximumCostMsat: (e.floatValue || 0) * 1000,
    }));
  }

  function createChangeHandler(key: keyof RebalanceConfiguration) {
    return (e: NumberFormatValues) => {
      setConfiguration((prev) => ({
        ...prev,
        [key]: e.floatValue,
      }));
    };
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: configuration,
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

  const destinationChannelsIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "destinationChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const destinationChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: destinationChannelsIds,
    })
  );

  const sourceChannelIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "sourceChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const sourceChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: sourceChannelIds,
    })
  );

  const avoidChannelsIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "avoidChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const avoidChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: avoidChannelsIds,
    })
  );

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<RebalanceConfiguratorIcon />}
      colorVariant={NodeColorVariant.accent1}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Destinations}
          selectedNodes={destinationChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"destinationChannels"}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Sources}
          selectedNodes={sourceChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"sourceChannels"}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Avoid}
          selectedNodes={avoidChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"avoidChannels"}
        />
        <Input
          formatted={true}
          value={amountSat}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={handleAmountSatChange}
          label={t.amountSat}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={maximumCostSat}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={handleMaximumCostSatChange}
          label={t.maximumCostSat}
          sizeVariant={InputSizeVariant.small}
        />
        <Input
          formatted={true}
          value={configuration.maximumCostMilliMsat}
          thousandSeparator={","}
          suffix={" ppm"}
          onValueChange={createChangeHandler("maximumCostMilliMsat")}
          label={t.maximumCostMilliMsat}
          sizeVariant={InputSizeVariant.small}
        />
        {/*<Input*/}
        {/*  formatted={true}*/}
        {/*  value={parameters.maximumConcurrency}*/}
        {/*  thousandSeparator={","}*/}
        {/*  suffix={" sat"}*/}
        {/*  onValueChange={createChangeHandler("maximumConcurrency")}*/}
        {/*  label={t.maximumConcurrency}*/}
        {/*  sizeVariant={InputSizeVariant.small}*/}
        {/*/>*/}
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
