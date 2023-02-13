import { ArrowRotateClockwise20Regular as RebalanceAutoRunIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function RebalanceAutoRunNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.RebalanceAutoRun}
      icon={<RebalanceAutoRunIcon />}
      title={t.workflowNodes.rebalanceAutoRun}
    />
  );
}
