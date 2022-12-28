import { Meta, Story } from "@storybook/react";
import Button, { ButtonProps, ColorVariant, SizeVariant } from "./Button";
import { DeleteRegular as Icon } from "@fluentui/react-icons";

export default {
  title: "components/buttons/Button",
  component: Button,
} as Meta;

const Template: Story<ButtonProps> = (args) => {
  return <Button {...args}>{"Button"}</Button>;
};

const defaultArgs = {
  icon: <Icon />,
};

// ----------------------------------------------
// ------------------ Primary -------------------
// ----------------------------------------------
export const Primary = Template.bind({});
Primary.args = {
  ...defaultArgs,
};

export const PrimarySmall = Template.bind({});
PrimarySmall.args = {
  ...defaultArgs,
  buttonSize: SizeVariant.small,
};

export const PrimaryTiny = Template.bind({});
PrimaryTiny.args = {
  ...defaultArgs,
  buttonSize: SizeVariant.tiny,
};

// ----------------------------------------------
// ------------------ Accent 1 ------------------
// ----------------------------------------------
export const Accent1 = Template.bind({});
Accent1.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent1,
};

export const Accent1Small = Template.bind({});
Accent1Small.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent1,
  buttonSize: SizeVariant.small,
};

export const Accent1Tiny = Template.bind({});
Accent1Tiny.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent1,
  buttonSize: SizeVariant.tiny,
};

// ----------------------------------------------
// ------------------ Accent 2 ------------------
// ----------------------------------------------
export const Accent2 = Template.bind({});
Accent2.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent2,
};

export const Accent2Small = Template.bind({});
Accent2Small.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent2,
  buttonSize: SizeVariant.small,
};

export const Accent2Tiny = Template.bind({});
Accent2Tiny.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent2,
  buttonSize: SizeVariant.tiny,
};

// ----------------------------------------------
// ------------------ Accent 3 ------------------
// ----------------------------------------------
export const Accent3 = Template.bind({});
Accent3.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent3,
};

export const Accent3Small = Template.bind({});
Accent3Small.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent3,
  buttonSize: SizeVariant.small,
};

export const Accent3Tiny = Template.bind({});
Accent3Tiny.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.accent3,
  buttonSize: SizeVariant.tiny,
};

// ----------------------------------------------
// ------------------ Warning -------------------
// ----------------------------------------------
export const Warning = Template.bind({});
Warning.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.warning,
};

export const WarningSmall = Template.bind({});
WarningSmall.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.warning,
  buttonSize: SizeVariant.small,
};

export const WarningTiny = Template.bind({});
WarningTiny.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.warning,
  buttonSize: SizeVariant.tiny,
};

// ----------------------------------------------
// ------------------ Error ---------------------
// ----------------------------------------------
export const Error = Template.bind({});
Error.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.error,
};

export const ErrorSmall = Template.bind({});
ErrorSmall.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.error,
  buttonSize: SizeVariant.small,
};

export const ErrorTiny = Template.bind({});
ErrorTiny.args = {
  ...defaultArgs,
  buttonColor: ColorVariant.error,
  buttonSize: SizeVariant.tiny,
};
