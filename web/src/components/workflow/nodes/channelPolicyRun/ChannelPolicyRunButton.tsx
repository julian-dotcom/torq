import { MoneySettings20Regular as ChannelPolicyConfiguratorIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function ChannelPolicyRunNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.ChannelPolicyRun}
      icon={<ChannelPolicyConfiguratorIcon />}
      title={t.workflowNodes.channelPolicyRun}
    />
  );
}
