import { ArrowBounce20Regular as ChannelBalanceEventFilterIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function ChannelBalanceEventFilterNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      intercomTarget={"channel-balance-event-filter-node-button"}
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.ChannelBalanceEventFilter}
      icon={<ChannelBalanceEventFilterIcon />}
      title={t.workflowNodes.channelBalanceEventFilter}
      parameters={`{"$and": []}`}
    />
  );
}
