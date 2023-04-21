import { Scales20Regular as Icon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function BalanceTriggerNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      intercomTarget={"balance-trigger-node-button"}
      colorVariant={NodeColorVariant.primary}
      nodeType={WorkflowNodeType.ChannelBalanceEventTrigger}
      icon={<Icon />}
      title={t.workflowNodes.channelBalanceTrigger}
    />
  );
}
