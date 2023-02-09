import useTranslations from "services/i18n/useTranslations";
import { ChevronDoubleRight16Regular as StageTriggerIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import Note, { NoteType } from "features/note/Note";

type StageTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function StageTriggerNode({ ...wrapperProps }: StageTriggerNodeProps) {
  const { t } = useTranslations();

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      name={t.StageTrigger}
      headerIcon={<StageTriggerIcon />}
      colorVariant={NodeColorVariant.accent2}
      noOptions={true}
      outputName={"channels"}
    >
      <div style={{ flexGrow: 1 }}>
        <Note title={t.note} noteType={NoteType.info}>
          <p>{t.stageTriggerDetails.stageTriggerDescription}</p>
          <p>{t.stageTriggerDetails.stageTriggerDescription2}</p>
        </Note>
      </div>
    </WorkflowNodeWrapper>
  );
}
