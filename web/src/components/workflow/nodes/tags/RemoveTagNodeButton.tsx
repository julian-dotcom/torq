import { TagDismiss20Regular as TagHeaderIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function RemoveTagNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent3}
      nodeType={WorkflowNodeType.RemoveTag}
      icon={<TagHeaderIcon />}
      title={t.removeTag}
      parameters={'{ "applyTo": "channel", "addedTags": [], "removedTags": [] }'}
    />
  );
}
