import React from "react";
import { Status } from "constants/backend";

export const WorkflowContext = React.createContext<{
  workflowStatus: Status;
}>({
  workflowStatus: Status.Inactive,
});
