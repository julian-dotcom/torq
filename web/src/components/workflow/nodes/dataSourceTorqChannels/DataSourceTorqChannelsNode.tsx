import React, { useContext, useEffect, useState } from "react";
import {
  Play20Regular as DataSourceTorqChannelsIcon,
  Save16Regular as SaveIcon,
} from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { InputSizeVariant, Form, RadioChips } from "components/forms/forms";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import Spinny from "features/spinny/Spinny";

type DataSourceTorqChannelsNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export type TorqChannelsConfiguration = {
  source: string;
};

export function DataSourceTorqChannelsNode({ ...wrapperProps }: DataSourceTorqChannelsNodeProps) {
  const { t } = useTranslations();

  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;
  const toastRef = React.useContext(ToastContext);

  const [updateNode] = useUpdateNodeMutation();

  const [configuration, setConfiguration] = useState<TorqChannelsConfiguration>({
    source: "",
    ...wrapperProps.parameters,
  });

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    setDirty(
      JSON.stringify(wrapperProps.parameters, Object.keys(wrapperProps.parameters).sort()) !==
      JSON.stringify(configuration, Object.keys(configuration).sort()));
  }, [configuration, wrapperProps.parameters]);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (editingDisabled) {
      toastRef?.current?.addToast(t.toast.cannotModifyWorkflowActive, toastCategory.warn);
      return;
    }
    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: configuration,
    }).finally(() => {
      setProcessing(false);
    });
  }

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<DataSourceTorqChannelsIcon />}
      colorVariant={NodeColorVariant.accent2}
      outputName={"channels"}
    >
      <Form onSubmit={handleSubmit}>
        <RadioChips
          label={t.channels}
          sizeVariant={InputSizeVariant.small}
          groupName={"focus-switch-" + wrapperProps.workflowVersionNodeId}
          options={[
            {
              label: t.triggering,
              id: "focus-switch-event-" + wrapperProps.workflowVersionNodeId,
              checked: configuration.source === "event",
              onChange: () => setConfiguration((prev) => ({
                ...prev,
                ["source" as keyof TorqChannelsConfiguration]: "event",
              })),
            },
            {
              label: t.all,
              id: "focus-switch-all-" + wrapperProps.workflowVersionNodeId,
              checked: configuration.source === "all",
              onChange: () => setConfiguration((prev) => ({
                ...prev,
                ["source" as keyof TorqChannelsConfiguration]: "all",
              })),
            },
            {
              label: t.both,
              id: "focus-switch-eventXorAll-" + wrapperProps.workflowVersionNodeId,
              checked: configuration.source === "eventXorAll",
              onChange: () => setConfiguration((prev) => ({
                ...prev,
                ["source" as keyof TorqChannelsConfiguration]: "eventXorAll",
              })),
            },
          ]}
          editingDisabled={editingDisabled}
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
