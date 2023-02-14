import { ArrowForward20Regular as DataSourceAllChannelsIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export function DataSourceAllChannelsNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent2}
      nodeType={WorkflowNodeType.DataSourceAllChannels}
      icon={<DataSourceAllChannelsIcon />}
      title={t.workflowNodes.allChannels}
    />
  );
}
