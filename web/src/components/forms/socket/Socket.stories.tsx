import { store } from "store/store";
import { Provider } from "react-redux";
import { Story, Meta } from "@storybook/react";
import Socket, { SocketProps } from "./Socket";
import { InputSizeVariant, InputColorVaraint } from "components/forms/variants";
import { useArgs } from "@storybook/client-api";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import WorkflowNodeWrapper from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { WorkflowNode } from "pages/WorkflowPage/workflowTypes";
import { useState } from "react";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export default {
  title: "components/forms/Socket",
  component: Socket,
} as Meta;

const Template: Story<SocketProps> = (args) => {
  const [_, updateArgs] = useArgs();

  const nodeData: WorkflowNode = {
    workflowVersionNodeId: 1,
    name: "sdafsdf",
    status: 1,
    type: 3,
    parameters: {},
    visibilitySettings: {
      collapsed: false,
      xPosition: 100,
      yPosition: 100,
    },
    updatedOn: "2022-12-22T18:01:38.655Z",
    parentNodes: {},
    childNodes: {},
    LinkDetails: {},
    workflowVersionId: 1,
  };
  const [positionsState, setPositionsState] = useState({ 1: { x: 100, y: 100 } });
  function handlePositionChange(stage: number, position: { x: number; y: number }) {
    setPositionsState({ ...positionsState, [stage]: position });
  }

  return (
    <Provider store={store}>
      <WorkflowCanvas active={true} workflowVersionId={1} stageNumber={1}>
        <WorkflowNodeWrapper id={"test"} heading={"test"} {...nodeData} colorVariant={NodeColorVariant.accent2}>
          <Socket {...args} />
        </WorkflowNodeWrapper>
      </WorkflowCanvas>
    </Provider>
  );
};

const defaultArgs = {
  label: "Destinations",
  placeholder: "Placeholder",
  sizeVariant: InputSizeVariant.normal,
  colorVariant: InputColorVaraint.primary,
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;

export const Accent1 = Template.bind({});
Accent1.args = { ...defaultArgs, colorVariant: InputColorVaraint.accent1 };

export const Accent2 = Template.bind({});
Accent2.args = { ...defaultArgs, colorVariant: InputColorVaraint.accent2 };

export const Accent3 = Template.bind({});
Accent3.args = { ...defaultArgs, colorVariant: InputColorVaraint.accent3 };

export const Warning = Template.bind({});
Warning.args = {
  ...defaultArgs,
  colorVariant: InputColorVaraint.primary,
  warningText: "Warning, this value is dangerous",
};

export const Error = Template.bind({});
Error.args = {
  ...defaultArgs,
  colorVariant: InputColorVaraint.primary,
  errorText: "Error: Something went wrong. Change the input value.",
};
