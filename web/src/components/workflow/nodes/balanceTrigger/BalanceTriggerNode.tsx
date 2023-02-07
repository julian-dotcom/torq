import useTranslations from "services/i18n/useTranslations";
import { Scales20Regular as Icon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import Note, { NoteType } from "features/note/Note";
import NodeConnector from "components/workflow/nodeWrapper/NodeConnector";
import {useId} from "react";

type BalanceTriggerNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function BalanceTriggerNode({ ...wrapperProps }: BalanceTriggerNodeProps) {
  const { t } = useTranslations();
  const connectorId = useId();

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      name={t.workflowNodes.channelBalanceTrigger}
      headerIcon={<Icon />}
      colorVariant={NodeColorVariant.accent2}
    >
      <div style={{ flexGrow: 1 }}>
        <Note title={t.note} noteType={NoteType.info}>
          <p>{t.workflowNodes.balanceTriggerNodeDescription}</p>
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
