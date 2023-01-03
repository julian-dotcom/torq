import { ArrowRotateClockwise20Regular as ReBalanceIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function ReBalanceChannelNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.RebalanceParameters}
      icon={<ReBalanceIcon />}
      title={t.workflowNodes.Rebalance}
    />
  );
}
