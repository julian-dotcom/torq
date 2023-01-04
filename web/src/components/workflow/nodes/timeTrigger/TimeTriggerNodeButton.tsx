import { Timer20Regular as TimeTriggerIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function TimeTriggerNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent2}
      nodeType={WorkflowNodeType.TimeTrigger}
      icon={<TimeTriggerIcon />}
      title={t.workflowNodes.timeTrigger}
    />
  );
}
