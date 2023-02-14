import {
  ArrowForward20Regular as DataSourceEventChannelsIcon,
} from "@fluentui/react-icons";
import WorkflowNodeWrapper, { WorkflowNodeProps } from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

type DataSourceEventChannelsNodeProps = Omit<WorkflowNodeProps, "colorVariant">;

export function DataSourceEventChannelsNode({ ...wrapperProps }: DataSourceEventChannelsNodeProps) {
  return (
    <WorkflowNodeWrapper
      {...wrapperProps}
      headerIcon={<DataSourceEventChannelsIcon />}
      colorVariant={NodeColorVariant.accent2}
      outputName={"channels"}
    >
      <div style={{ flexGrow: 1 }}>
      </div>
    </WorkflowNodeWrapper>
  );
}
