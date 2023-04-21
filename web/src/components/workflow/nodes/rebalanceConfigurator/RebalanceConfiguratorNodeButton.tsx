import { ArrowRotateClockwise20Regular as RebalanceConfiguratorIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function RebalanceConfiguratorNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      intercomTarget={"rebalance-configurator-node-button"}
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.RebalanceConfigurator}
      icon={<RebalanceConfiguratorIcon />}
      title={t.workflowNodes.rebalanceConfigurator}
    />
  );
}
