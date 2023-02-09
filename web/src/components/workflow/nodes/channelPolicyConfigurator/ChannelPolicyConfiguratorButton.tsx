import { MoneySettings20Regular as ChannelPolicyConfiguratorIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function ChannelPolicyConfiguratorNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.ChannelPolicyConfigurator}
      icon={<ChannelPolicyConfiguratorIcon />}
      title={t.workflowNodes.channelPolicyConfigurator}
      disabled={true}
    />
  );
}
