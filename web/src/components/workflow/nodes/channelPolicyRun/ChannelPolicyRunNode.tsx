import { MoneySettings20Regular as ChannelPolicyConfiguratorIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes } from "pages/WorkflowPage/workflowApi";
import { useSelector } from "react-redux";
import { WorkflowContext } from "components/workflow/WorkflowContext";
import { Status } from "constants/backend";
import { useContext } from "react";

type ChannelPolicyRunNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function ChannelPolicyRunNode({ ...wrapperProps }: ChannelPolicyRunNodeProps) {
  const { t } = useTranslations();

  const { workflowStatus } = useContext(WorkflowContext);
  const editingDisabled = workflowStatus === Status.Active;

  const { childLinks } = useSelector(
    SelectWorkflowNodeLinks({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeId: wrapperProps.workflowVersionNodeId,
      stage: wrapperProps.stage,
    })
  );

  const routingPolicySettingIds =
    childLinks
      ?.filter((n) => {
        return n.childInput === "routingPolicySettings";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const routingPolicySettings = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: routingPolicySettingIds,
    })
  );

  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<ChannelPolicyConfiguratorIcon />}
      colorVariant={NodeColorVariant.accent1}
    >
      <div style={{ flexGrow: 1 }}>
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.inputs}
          selectedNodes={routingPolicySettings || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"routingPolicySettings"}
          editingDisabled={editingDisabled}
        />
      </div>
    </WorkflowNodeWrapper>
  );
}
