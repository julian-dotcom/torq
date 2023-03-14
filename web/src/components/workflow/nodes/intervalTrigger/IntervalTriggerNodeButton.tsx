import { Timer20Regular as IntervalTriggerIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function IntervalTriggerNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.primary}
      nodeType={WorkflowNodeType.IntervalTrigger}
      icon={<IntervalTriggerIcon />}
      title={t.workflowNodes.intervalTrigger}
      parameters={'{"seconds": 3600, "timeUnit": "seconds"}'}
    />
  );
}
