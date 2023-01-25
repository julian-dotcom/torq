import { Filter20Regular as ChannelFilterIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function ChannelFilterNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent1}
      nodeType={WorkflowNodeType.ChannelFilter}
      icon={<ChannelFilterIcon />}
      title={t.workflowNodes.channelFilter}
      parameters={`{"$and": []}`}
    />
  );
}
