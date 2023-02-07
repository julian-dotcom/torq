import { MoneySettings20Regular as ChannelPolicyConfiguratorIcon } from "@fluentui/react-icons";
import useTranslations from "services/i18n/useTranslations";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import Socket from "components/forms/socket/Socket";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";
import { SelectWorkflowNodeLinks, SelectWorkflowNodes } from "pages/WorkflowPage/workflowApi";
import { useSelector } from "react-redux";

type ChannelPolicyRunNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function ChannelPolicyRunNode({ ...wrapperProps }: ChannelPolicyRunNodeProps) {
  const { t } = useTranslations();

  const { childLinks } = useSelector(
    SelectWorkflowNodeLinks({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeId: wrapperProps.workflowVersionNodeId,
      stage: wrapperProps.stage,
    })
  );

  const triggered =
    childLinks
      ?.filter((n) => {
        return n.childInput === "triggered";
      })
      ?.map((link) => link.parentWorkflowVersionNodeId) ?? [];

  const triggers = useSelector(
    SelectWorkflowNodes({
      version: wrapperProps.version,
      workflowId: wrapperProps.workflowId,
      nodeIds: triggered,
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
          label={t.trigger}
          selectedNodes={triggers || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"triggered"}
        />
        <Socket
          collapsed={wrapperProps.visibilitySettings.collapsed}
          label={t.inputs}
          selectedNodes={routingPolicySettings || []}
          workflowVersionId={wrapperProps.workflowVersionId}
          workflowVersionNodeId={wrapperProps.workflowVersionNodeId}
          inputName={"routingPolicySettings"}
        />
      </div>
    </WorkflowNodeWrapper>
  );
}
