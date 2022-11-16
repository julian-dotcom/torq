import { Meta, Story } from "@storybook/react";
import Select, { SelectProps } from "./Select";
import { InputColorVaraint } from "components/forms/select/variants";
import { InputSizeVariant } from "./variants";
// import { useArgs } from "@storybook/client-api";

export default {
  title: "components/forms/Select",
  component: Select,
} as Meta;

const Template: Story<SelectProps> = (args) => {
  // const [_, updateArgs] = useArgs();
  return (
    <div style={{ height: "400px" }}>
      <Select {...args} />
    </div>
  );
};

const defaultArgs: SelectProps = {
  label: "Label",
  colorVariant: InputColorVaraint.primary,
  sizeVariant: InputSizeVariant.normal,
  options: [
    { value: "1", label: "Option 1" },
    { value: "2", label: "Option 2" },
    { value: "3", label: "Option 3" },
  ],
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;

export const Open = Template.bind({});
Open.args = {
  ...defaultArgs,
  menuIsOpen: true,
};

export const Small = Template.bind({});
Small.args = {
  ...defaultArgs,
  sizeVariant: InputSizeVariant.small,
  menuIsOpen: true,
};

export const Tiny = Template.bind({});
Tiny.args = {
  ...defaultArgs,
  sizeVariant: InputSizeVariant.tiny,
  menuIsOpen: true,
};

export const Accent1 = Template.bind({});
Accent1.args = { ...defaultArgs, colorVariant: InputColorVaraint.accent1 };

export const Accent2 = Template.bind({});
Accent2.args = { ...defaultArgs, colorVariant: InputColorVaraint.accent2 };

export const Accent3 = Template.bind({});
Accent3.args = { ...defaultArgs, colorVariant: InputColorVaraint.accent3 };

export const Warning = Template.bind({});
Warning.args = { ...defaultArgs, warningText: "Warning: Something needs attention" };

export const Error = Template.bind({});
Error.args = { ...defaultArgs, errorText: "Error: Something is wrong" };
