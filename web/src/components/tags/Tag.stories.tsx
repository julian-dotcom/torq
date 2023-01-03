import { Meta, Story } from "@storybook/react";
import Tag, { TagColor, TagProps } from "./Tag";

export default {
  title: "components/tags/Tag",
  component: Tag,
} as Meta;

const Template: Story<TagProps> = (args) => {
  return <Tag {...args} />;
};

const defaultArgs: TagProps = {
  label: "Drain",
};
const argTypes = {
  locked: {
    control: {
      type: "boolean",
    },
  },
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;
Primary.argTypes = argTypes;

export const Locked = Template.bind({});
Locked.args = {
  ...defaultArgs,
  locked: true,
};

export const Success = Template.bind({});
Success.args = {
  ...defaultArgs,
  locked: true,
  colorVariant: TagColor.success,
};

export const Warning = Template.bind({});
Warning.args = {
  ...defaultArgs,
  locked: true,
  colorVariant: TagColor.warning,
};

export const Error = Template.bind({});
Error.args = {
  ...defaultArgs,
  locked: true,
  colorVariant: TagColor.error,
};

export const Accent1 = Template.bind({});
Accent1.args = {
  ...defaultArgs,
  locked: true,
  colorVariant: TagColor.accent1,
};

export const Accent2 = Template.bind({});
Accent2.args = {
  ...defaultArgs,
  locked: true,
  colorVariant: TagColor.accent2,
};

export const Accent3 = Template.bind({});
Accent3.args = {
  ...defaultArgs,
  locked: true,
  colorVariant: TagColor.accent3,
};

export const Custom = Template.bind({});
Custom.args = {
  ...defaultArgs,
  locked: true,
  customTextColor: "#fff",
  customBackgroundColor: "blue",
};

// export const Small = Template.bind({});
// Small.args = {
//   ...defaultArgs,
//   sizeVariant: SwitchSize.small,
// };
//
// export const Tiny = Template.bind({});
// Tiny.args = {
//   ...defaultArgs,
//   sizeVariant: SwitchSize.tiny,
// };
