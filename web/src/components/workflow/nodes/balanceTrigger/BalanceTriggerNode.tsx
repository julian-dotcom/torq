import useTranslations from "services/i18n/useTranslations";
import { Scales20Regular as Icon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import Note, { NoteType } from "features/note/Note";

type BalanceTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function BalanceTriggerNode({ ...wrapperProps }: BalanceTriggerNodeProps) {
  const { t } = useTranslations();

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      name={t.workflowNodes.channelBalanceTrigger}
      headerIcon={<Icon />}
      colorVariant={NodeColorVariant.primary}
    >
      <div style={{ flexGrow: 1 }}>
        <Note title={t.note} noteType={NoteType.info}>
          <p>{t.workflowNodes.balanceTriggerNodeDescription}</p>
        </Note>
      </div>
    </WorkflowNodeWrapper>
  );
}
