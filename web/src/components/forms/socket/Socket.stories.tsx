import { store } from "store/store";
import { Provider } from "react-redux";
import { Story, Meta } from "@storybook/react";
import Socket, { SocketProps } from "./Socket";
import { InputSizeVariant, InputColorVaraint } from "components/forms/variants";
import { useArgs } from "@storybook/client-api";
import WorkflowCanvas from "components/workflow/canvas/WorkflowCanvas";
import WorkflowNodeWrapper from "components/workflow/nodeWrapper/WorkflowNodeWrapper";
import { WorkflowVersionNode } from "pages/WorkflowPage/workflowTypes";
import { NodeColorVariant } from "components/workflow/nodes/nodeVariants";

export default {
  title: "components/forms/Socket",
  component: Socket,
} as Meta;

const Template: Story<SocketProps> = (args) => {
  const [_, updateArgs] = useArgs();

  const nodeData: WorkflowVersionNode = {
    workflowVersionNodeId: 1,
    workflowId: 1,
    version: 1,
    name: "sdafsdf",
    status: 1,
    stage: 1,
    type: 3,
    parameters: {},
    visibilitySettings: {
      collapsed: false,
      xPosition: 0,
      yPosition: 0,
    },
    updatedOn: "2022-12-22T18:01:38.655Z",
    workflowVersionId: 1,
  };

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
