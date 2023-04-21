import { PlugConnected20Regular as Icon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function ChannelOpenTriggerNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      intercomTarget={"channel-open-trigger-node-button"}
      colorVariant={NodeColorVariant.primary}
      nodeType={WorkflowNodeType.ChannelOpenEventTrigger}
      icon={<Icon />}
      title={t.workflowNodes.openChannelTrigger}
    />
  );
}
