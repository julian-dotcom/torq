import { Play20Regular as DataSourceAllChannelsIcon } from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

type DataSourceAllChannelsNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function DataSourceAllChannelsNode({ ...wrapperProps }: DataSourceAllChannelsNodeProps) {
  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<DataSourceAllChannelsIcon />}
      colorVariant={NodeColorVariant.accent2}
      outputName={"channels"}
    >
      <div style={{ flexGrow: 1 }}>
      </div>
    </WorkflowNodeWrapper>
  );
}
