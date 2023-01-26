import { Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import { CalendarClock20Regular as CronTriggerIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { Form, Input, InputRow, InputSizeVariant } from "components/forms/forms";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import cronstrue from "cronstrue";
import React from "react";

type CronTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function CronTriggerNode({ ...wrapperProps }: CronTriggerNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();
  const [cronValueState, setCronValueState] = React.useState("0 23 ? * MON-FRI");

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: {
        cronValue: cronValueState,
      },
    });
  }

  function handleCronChange({ currentTarget: { value } }: React.FormEvent<HTMLInputElement>) {
    setCronValueState(value);
  }

  let cronExplained = "";
  try {
    cronExplained = cronstrue.toString(cronValueState);
  } catch (err) {
    cronExplained = "Invalid cron value";
  }

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      heading={t.channelPolicyConfiguration}
      headerIcon={<CronTriggerIcon />}
      colorVariant={NodeColorVariant.accent2}
    >
      <Form onSubmit={handleSubmit}>
        <InputRow>
          <div style={{ flexGrow: 1 }}>
            <Input
              value={cronValueState}
              thousandSeparator={true}
              onChange={handleCronChange}
              label={t.cron}
              helpText={"Interval specified in Cron format"}
              sizeVariant={InputSizeVariant.small}
            />
          </div>
        </InputRow>
        <span>{cronExplained}</span>
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
