import useTranslations from "services/i18n/useTranslations";
import { PlugDisconnected20Regular as Icon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import Note, { NoteType } from "features/note/Note";
import { useId } from "react";
import NodeConnector from "components/workflow/nodeWrapper/NodeConnector";

type Props = Omit<WorkflowNodeProps, "colorVariant">;

export function ChannelCloseTriggerNode({ ...wrapperProps }: Props) {
  const { t } = useTranslations();
  const connectorId = useId();

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      name={t.workflowNodes.closeChannelTrigger}
      headerIcon={<Icon />}
      colorVariant={NodeColorVariant.accent2}
    >
      <div style={{ flexGrow: 1 }}>
        <Note title={t.note} noteType={NoteType.info}>
          <p>{t.workflowNodes.channelCloseTriggerDetails}</p>
        </Note>
      </div>
      <NodeConnector
        id={connectorId}
        name={t.trigger}
        outputName={"channels"}
        workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
        workflowVersionId={wrapperProps.workflowVersionId}
      />
    </WorkflowNodeWrapper>
  );
}
