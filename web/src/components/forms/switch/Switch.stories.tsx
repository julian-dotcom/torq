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
  checked: false,
  sizeVariant: SwitchSize.normal,
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;

export const Unchecked = Template.bind({});
Unchecked.args = defaultArgs;

export const Checked = Template.bind({});
Checked.args = {
  ...defaultArgs,
  checked: true,
};

export const Small = Template.bind({});
Small.args = {
  ...defaultArgs,
  checked: true,
  sizeVariant: SwitchSize.small,
};

export const Tiny = Template.bind({});
Tiny.args = {
  ...defaultArgs,
  checked: true,
  sizeVariant: SwitchSize.tiny,
};
