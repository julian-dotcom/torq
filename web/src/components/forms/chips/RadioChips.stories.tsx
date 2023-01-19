import { Story, Meta } from "@storybook/react";
// import { SearchRegular as LeftIcon } from "@fluentui/react-icons";
import RadioChips, { RadioChipsProps } from "./RadioChips";
import { InputSizeVariant, InputColorVaraint } from "components/forms/variants";

export default {
  title: "components/forms/RadioChips",
  component: RadioChips,
} as Meta;

const Template: Story<RadioChipsProps> = (args) => {
  return <RadioChips {...args} />;
};

const defaultArgs = {
  label: "Label",
  groupName: "example",
  options: [
    { id: "option1", label: "Option 1" },
    { id: "option2", label: "Option 2 (selected)", checked: true },
  ],
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
  // warningText: "Warning, this value is dangerous",
};

export const Error = Template.bind({});
Error.args = {
  ...defaultArgs,
  colorVariant: InputColorVaraint.primary,
  // errorText: "Error: Something went wrong. Change the input value.",
};

export const Small = Template.bind({});
Small.args = {
  ...defaultArgs,
  sizeVariant: InputSizeVariant.small,
};

export const Tiny = Template.bind({});
Tiny.args = {
  ...defaultArgs,
  sizeVariant: InputSizeVariant.tiny,
};
