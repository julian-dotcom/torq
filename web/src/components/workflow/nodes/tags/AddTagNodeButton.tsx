import { Tag20Regular as TagHeaderIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeButtonWrapper from "components/workflow/nodeButtonWrapper/NodeButtonWrapper";
import { WorkflowNodeType } from "pages/WorkflowPage/constants";
import { NodeColorVariant } from "../nodeVariants";

export function AddTagNodeButton() {
  const { t } = useTranslations();

  return (
    <WorkflowNodeButtonWrapper
      colorVariant={NodeColorVariant.accent3}
      nodeType={WorkflowNodeType.AddTag}
      icon={<TagHeaderIcon />}
      title={t.addTag}
      parameters={'{ "applyTo": "channel", "addedTags": [], "removedTags": [] }'}
    />
  );
}
