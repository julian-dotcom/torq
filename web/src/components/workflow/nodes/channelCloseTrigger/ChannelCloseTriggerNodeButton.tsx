import { PlugDisconnected20Regular as Icon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function ChannelCloseTriggerNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      intercomTarget={"channel-close-trigger-node-button"}
      colorVariant={NodeColorVariant.primary}
      nodeType={WorkflowNodeType.ChannelCloseEventTrigger}
      icon={<Icon />}
      title={t.workflowNodes.closeChannelTrigger}
    />
  );
}
