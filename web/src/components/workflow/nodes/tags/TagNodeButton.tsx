import { Tag20Regular as TagHeaderIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function TagNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent3}
      nodeType={WorkflowNodeType.Tag}
      icon={<TagHeaderIcon />}
      title={t.workflowNodes.tag}
      // TODO: After merging with master, add the bellow default parameters on drop
      // parameters={{ applyTo: "channels", addedTags: [], removedTags: [] }}
    />
  );
}
