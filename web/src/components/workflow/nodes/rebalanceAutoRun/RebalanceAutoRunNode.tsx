import React, { useContext, useEffect, useState } from "react";
import {
  ArrowRotateClockwise20Regular as RebalanceConfiguratorIcon,
  Save16Regular as SaveIcon,
} from "@fluentui/react-icons";
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
import Spinny from "features/spinny/Spinny";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";

type RebalanceAutoRunNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export type RebalanceConfiguration = {
  amountMsat?: number;
  maximumCostMsat?: number;
  maximumCostMilliMsat?: number;
  maximumConcurrency?: number;
};

export function RebalanceAutoRunNode({ ...wrapperProps }: RebalanceAutoRunNodeProps) {
  const { t } = useTranslations();
  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;
  const toastRef = React.useContext(ToastContext);

  const [updateNode] = useUpdateNodeMutation();

  const [rebalance, setRebalance] = useState<RebalanceConfiguration>({
    amountMsat: undefined,
    maximumCostMsat: undefined,
    maximumCostMilliMsat: undefined,
    maximumConcurrency: undefined,
    ...wrapperProps.parameters,
  });

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    setDirty(
      JSON.stringify(wrapperProps.parameters, Object.keys(wrapperProps.parameters).sort()) !==
      JSON.stringify(rebalance, Object.keys(rebalance).sort()));
  }, [rebalance, wrapperProps.parameters]);

  const [amountSat, setAmountSat] = useState<number | undefined>(
    (wrapperProps.parameters as RebalanceConfiguration).amountMsat
      ? ((wrapperProps.parameters as RebalanceConfiguration).amountMsat || 0) / 1000
      : undefined
  );
  const [maximumCostSat, setMaximumCostSat] = useState<number | undefined>(
    (wrapperProps.parameters as RebalanceConfiguration).maximumCostMsat
      ? ((wrapperProps.parameters as RebalanceConfiguration).maximumCostMsat || 0) / 1000
      : undefined
  );

  function createChangeMsatHandler(key: keyof RebalanceConfiguration) {
    return (e: NumberFormatValues) => {
      if (key == "amountMsat") {
        setAmountSat(e.floatValue);
      }
      if (key == "maximumCostMsat") {
        setMaximumCostSat(e.floatValue);
      }
      if (e.floatValue === undefined) {
        setRebalance((prev) => ({
          ...prev,
          [key]: undefined,
        }));
      } else {
        setRebalance((prev) => ({
          ...prev,
          [key]: (e.floatValue || 0) * 1000,
        }));
      }
    };
  }

  function createChangeHandler(key: keyof RebalanceConfiguration) {
    return (e: NumberFormatValues) => {
      setRebalance((prev) => ({
        ...prev,
        [key]: e.floatValue,
      }));
    };
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: rebalance,
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

  const incomingChannelIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "incomingChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const incomingChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: incomingChannelIds,
    })
  );

  const outgoingChannelIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "outgoingChannels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const outgoingChannels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: outgoingChannelIds,
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
      outputName={"channels"}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Destinations}
          selectedNodes={incomingChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"incomingChannels"}
          editingDisabled={editingDisabled}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Sources}
          selectedNodes={outgoingChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"outgoingChannels"}
          editingDisabled={editingDisabled}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.Avoid}
          selectedNodes={avoidChannels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"avoidChannels"}
          editingDisabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={amountSat}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeMsatHandler("amountMsat")}
          label={t.amountSat}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={maximumCostSat}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeMsatHandler("maximumCostMsat")}
          label={t.maximumCostSat}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={rebalance.maximumCostMilliMsat}
          thousandSeparator={","}
          suffix={" ppm"}
          onValueChange={createChangeHandler("maximumCostMilliMsat")}
          label={t.maximumCostMilliMsat}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
        />
        {/*<Input*/}
        {/*  formatted={true}*/}
        {/*  value={rebalance.maximumConcurrency}*/}
        {/*  thousandSeparator={","}*/}
        {/*  onValueChange={createChangeHandler("maximumConcurrency")}*/}
        {/*  label={t.maximumConcurrency}*/}
        {/*  sizeVariant={InputSizeVariant.small}*/}
        {/*/>*/}
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
