import { Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import { CalendarClock20Regular as CronTriggerIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { Form, Input, InputRow, InputSizeVariant } from "components/forms/forms";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import cronstrue from "cronstrue";
import React, { useContext, useEffect, useState } from "react";
import Spinny from "features/spinny/Spinny";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";

type CronTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function CronTriggerNode({ ...wrapperProps }: CronTriggerNodeProps) {
  const { t } = useTranslations();
  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;

  const [updateNode] = useUpdateNodeMutation();
  const [cronValueState, setCronValueState] = React.useState(
    (wrapperProps.parameters as { cronValue: string }).cronValue ?? "0 23 ? * MON-FRI"
  );
  const [isValidState, setIsValidState] = React.useState(true);

  const [dirty, setDirty] = useState(false);
  const [processing, setProcessing] = useState(false);
  useEffect(() => {
    if (((wrapperProps.parameters as { cronValue: string }).cronValue ?? "0 23 ? * MON-FRI") !== cronValueState) {
      setDirty(true);
    } else {
      setDirty(false);
    }
  }, [cronValueState, wrapperProps.parameters]);

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (editingDisabled) {
      return;
    }
    setProcessing(true);
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: {
        cronValue: cronValueState,
      },
    }).finally(() => {
      setProcessing(false);
    });
  }

  function handleCronChange({ currentTarget: { value } }: React.FormEvent<HTMLInputElement>) {
    setCronValueState(value);
    try {
      cronstrue.toString(value);
    } catch (err) {
      setIsValidState(false);
      return;
    }
    setIsValidState(true);
  }

  let cronExplained = "";
  try {
    const cronDesc = cronstrue.toString(cronValueState);
    cronExplained = "Triggers " + cronDesc[0].toLowerCase() + cronDesc.substring(1);
  } catch (err) {
    cronExplained = "Invalid cron value";
  }

  return (
    <WorkflowNodeWrapper {...wrapperProps} headerIcon={<CronTriggerIcon />} colorVariant={NodeColorVariant.primary}>
      <Form onSubmit={handleSubmit} intercomTarget={"cron-trigger-node-form"}>
        <InputRow>
          <div style={{ flexGrow: 1 }}>
            <Input
              disabled={editingDisabled}
              value={cronValueState}
              onChange={handleCronChange}
              placeholder={"0 23 ? * MON-FRI"}
              label={t.cron}
              helpText={"Interval specified in Cron format"}
              sizeVariant={InputSizeVariant.small}
            />
          </div>
        </InputRow>
        <span className="info-box">
          <CronTriggerIcon /> {cronExplained}
        </span>
        <Button
          intercomTarget={"cron-trigger-node-save"}
          type="submit"
          buttonColor={ColorVariant.success}
          buttonSize={SizeVariant.small}
          icon={!processing ? <SaveIcon /> : <Spinny />}
          disabled={!isValidState || !dirty || processing || editingDisabled}
        >
          {!processing ? t.save.toString() : t.saving.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
