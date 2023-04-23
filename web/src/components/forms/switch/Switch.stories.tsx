import { Meta, Story } from "@storybook/react";
import Switch, { SwitchProps, SwitchSize } from "./Switch";
import { useArgs } from "@storybook/client-api";

export default {
  title: "components/forms/Switch",
  component: Switch,
} as Meta;

const Template: Story<SwitchProps> = (args) => {
  const [_, updateArgs] = useArgs();

  const onClickHandler = () => {
    updateArgs({ checked: !args.checked });
  };

  return <Switch {...args} onClick={onClickHandler} />;
};

const defaultArgs: SwitchProps = {
  label: "Add Torq",
  intercomTarget: "switch",
  sizeVariant: SwitchSize.normal,
};
const argTypes = {
  checked: {
    control: {
      type: "boolean",
    },
  },
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;
Primary.argTypes = argTypes;

export const Unchecked = Template.bind({});
Unchecked.args = defaultArgs;

export const Checked = Template.bind({});
Checked.args = {
  ...defaultArgs,
  defaultChecked: true,
};

export const Small = Template.bind({});
Small.args = {
  ...defaultArgs,
  sizeVariant: SwitchSize.small,
};

export const Tiny = Template.bind({});
Tiny.args = {
  ...defaultArgs,
  sizeVariant: SwitchSize.tiny,
};
