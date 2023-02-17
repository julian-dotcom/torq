import { Play20Regular as DataSourceTorqChannelsIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function DataSourceTorqChannelsNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent2}
      nodeType={WorkflowNodeType.DataSourceTorqChannels}
      icon={<DataSourceTorqChannelsIcon />}
      title={t.workflowNodes.torqChannels}
    />
  );
}
