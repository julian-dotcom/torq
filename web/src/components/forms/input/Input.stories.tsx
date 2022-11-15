import { Story, Meta } from "@storybook/react";
import { SearchRegular as LeftIcon } from "@fluentui/react-icons";
import Input from "./Input";
import { InputSizeVariant, InputColorVaraint } from "./variants";
import { useArgs } from "@storybook/client-api";

export default {
  title: "components/forms/Input",
  component: Input,
} as Meta;

const Template: Story<typeof Input> = (args) => {
  const [_, updateArgs] = useArgs();

  return <Input {...args} />;
};

const defaultArgs = {
  label: "Label",
  placeholder: "Placeholder",
  sizeVariant: InputSizeVariant.normal,
  colorVariant: InputColorVaraint.primary,
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;

export const Colors = Template.bind({});
Colors.args = { ...defaultArgs, colorVariant: InputColorVaraint.accent1 };

export const Small = Template.bind({});
Small.args = {
  ...defaultArgs,
  checked: true,
  sizeVariant: InputSizeVariant.small,
};

export const Tiny = Template.bind({});
Tiny.args = {
  ...defaultArgs,
  checked: true,
  sizeVariant: InputSizeVariant.tiny,
};

export const Search = Template.bind({});
Search.args = {
  ...defaultArgs,
  placeholder: "Search...",
  leftIcon: <LeftIcon />,
};

export const SearchSmall = Template.bind({});
SearchSmall.args = {
  ...defaultArgs,
  placeholder: "Search...",
  leftIcon: <LeftIcon />,
  sizeVariant: InputSizeVariant.small,
};

export const SearchTiny = Template.bind({});
SearchTiny.args = {
  ...defaultArgs,
  placeholder: "Search...",
  leftIcon: <LeftIcon />,
  sizeVariant: InputSizeVariant.tiny,
};

export const NumberType = Template.bind({});
NumberType.args = {
  ...defaultArgs,
  type: "number",
  defaultValue: 1200000,
};

export const DateType = Template.bind({});
DateType.args = {
  ...defaultArgs,
  type: "date",
};

export const DatetimeType = Template.bind({});
DatetimeType.args = {
  ...defaultArgs,
  type: "datetime-local",
};

export const FormattedInput = Template.bind({});
FormattedInput.args = {
  leftIcon: undefined,
  formatted: true,
  thousandSeparator: ",",
  defaultValue: 1200000,
};
