import { Money20Regular as ChannelPolicyConfiguratorIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function ChannelPolicyAutoRunNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      intercomTarget={"channel-policy-auto-run-node-button"}
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.ChannelPolicyAutoRun}
      icon={<ChannelPolicyConfiguratorIcon />}
      title={t.workflowNodes.channelPolicyAutoRun}
    />
  );
}
