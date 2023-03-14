import { CalendarClock20Regular as CronTriggerIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function CronTriggerNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.primary}
      nodeType={WorkflowNodeType.CronTrigger}
      icon={<CronTriggerIcon />}
      title={t.workflowNodes.cronTrigger}
      parameters={'{ "cronValue": "0 23 ? * MON-FRI"}'}
    />
  );
}
