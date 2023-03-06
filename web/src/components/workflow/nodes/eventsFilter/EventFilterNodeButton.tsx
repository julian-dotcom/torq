import { ArrowBounce20Regular as EventFilterIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function EventFilterNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.EventFilter}
      icon={<EventFilterIcon />}
      title={t.workflowNodes.eventFilter}
      parameters={`{"$and": []}`}
    />
  );
}
