import { Save16Regular as SaveIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import { CalendarClock20Regular as CronTriggerIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { Form, Input, InputRow, InputSizeVariant } from "components/forms/forms";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { useUpdateNodeMutation } from "pages/WorkflowPage/workflowApi";
import { NumberFormatValues } from "react-number-format";

type CronTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function CronTriggerNode({ ...wrapperProps }: CronTriggerNodeProps) {
  const { t } = useTranslations();

  const [updateNode] = useUpdateNodeMutation();

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    updateNode({
      workflowVersionNodeId: wrapperProps.workflowVersionNodeId,
      parameters: {
        cronValue: "5 4 * * *",
      },
    });
  }

  function handleCronChange(value: NumberFormatValues) {
    console.log(value);
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
              value={"5 4 * * *"}
              thousandSeparator={true}
              onValueChange={handleCronChange}
              label={t.cron}
              helpText={"Interval specified in Cron format"}
              sizeVariant={InputSizeVariant.small}
            />
          </div>
        </InputRow>
        <Button type="submit" buttonColor={ColorVariant.success} buttonSize={SizeVariant.small} icon={<SaveIcon />}>
          {t.save.toString()}
        </Button>
      </Form>
    </WorkflowNodeWrapper>
  );
}
