import React, { useContext, useEffect, useState } from "react";
import {
  MoneySettings20Regular as ChannelPolicyConfiguratorIcon,
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

type ChannelPolicyConfiguratorNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export type ChannelPolicyConfiguration = {
  feeBaseMsat?: number;
  feeRateMilliMsat?: number;
  maxHtlcMsat?: number;
  minHtlcMsat?: number;
  timeLockDelta?: number;
};

export function ChannelPolicyConfiguratorNode({ ...wrapperProps }: ChannelPolicyConfiguratorNodeProps) {
  const { t } = useTranslations();

  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;
  const toastRef = React.useContext(ToastContext);

  const [updateNode] = useUpdateNodeMutation();

  const [channelPolicy, setChannelPolicy] = useState<ChannelPolicyConfiguration>({
    feeBaseMsat: undefined,
    feeRateMilliMsat: undefined,
    maxHtlcMsat: undefined,
    minHtlcMsat: undefined,
    timeLockDelta: undefined,
    ...wrapperProps.parameters,
  });

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    // if the original parameters are different from the current parameters, set dirty to true
    if (JSON.stringify(wrapperProps.parameters) !== JSON.stringify(channelPolicy)) {
      setDirty(true);
    } else {
      setDirty(false);
    }
  }, [channelPolicy, wrapperProps.parameters]);

  const [feeBase, setFeeBase] = useState<number | undefined>(
    (wrapperProps.parameters as ChannelPolicyConfiguration).feeBaseMsat
      ? ((wrapperProps.parameters as ChannelPolicyConfiguration).feeBaseMsat || 0) / 1000
      : undefined
  );
  const [maxHtlc, setMaxHtlc] = useState<number | undefined>(
    (wrapperProps.parameters as ChannelPolicyConfiguration).maxHtlcMsat
      ? ((wrapperProps.parameters as ChannelPolicyConfiguration).maxHtlcMsat || 0) / 1000
      : undefined
  );
  const [minHtlc, setMinHtlc] = useState<number | undefined>(
    (wrapperProps.parameters as ChannelPolicyConfiguration).minHtlcMsat
      ? ((wrapperProps.parameters as ChannelPolicyConfiguration).minHtlcMsat || 0) / 1000
      : undefined
  );

  function createChangeMsatHandler(key: keyof ChannelPolicyConfiguration) {
    return (e: NumberFormatValues) => {
      if (key == "feeBaseMsat") {
        setFeeBase(e.floatValue);
      }
      if (key == "maxHtlcMsat") {
        setMaxHtlc(e.floatValue);
      }
      if (key == "minHtlcMsat") {
        setMinHtlc(e.floatValue);
      }
      if (e.floatValue === undefined) {
        setChannelPolicy((prev) => ({
          ...prev,
          [key]: undefined,
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

    if (editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }

    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: channelPolicy,
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

  const routingPolicySettingIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "routingPolicySettings";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const routingPolicySettings = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: routingPolicySettingIds,
    })
  );

  const channelIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "channels";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const channels = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: channelIds,
    })
  );

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<ChannelPolicyConfiguratorIcon />}
      colorVariant={NodeColorVariant.accent1}
      outputName={"routingPolicySettings"}
    >
      <Form onSubmit={handleSubmit}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.inputs}
          selectedNodes={routingPolicySettings || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"routingPolicySettings"}
          editingDisabled={editingDisabled}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.channels}
          selectedNodes={channels || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"channels"}
          editingDisabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={channelPolicy.feeRateMilliMsat}
          thousandSeparator={","}
          suffix={" ppm"}
          onValueChange={createChangeHandler("feeRateMilliMsat")}
          label={t.updateChannelPolicy.feeRateMilliMsat}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={feeBase}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeMsatHandler("feeBaseMsat")}
          label={t.baseFee}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={minHtlc}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeMsatHandler("minHtlcMsat")}
          label={t.minHTLCAmount}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={maxHtlc}
          thousandSeparator={","}
          suffix={" sat"}
          onValueChange={createChangeMsatHandler("maxHtlcMsat")}
          label={t.maxHTLCAmount}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
        />
        <Input
          formatted={true}
          value={channelPolicy.timeLockDelta}
          thousandSeparator={","}
          onValueChange={createChangeHandler("timeLockDelta")}
          label={t.updateChannelPolicy.timeLockDelta}
          sizeVariant={InputSizeVariant.small}
          disabled={editingDisabled}
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
