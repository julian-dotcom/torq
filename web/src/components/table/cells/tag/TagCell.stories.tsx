import { Meta, Story } from "@storybook/react";
import TagCell, { TagCellProps } from "./TagCell";
import { TagColor } from "components/tags/Tag";

export default {
  title: "components/table/cells/TagCell",
  component: TagCell,
} as Meta;

const Template: Story<TagCellProps> = (args) => {
  return <TagCell {...args} />;
};

const defaultArgs: TagCellProps = {
  label: "Drain",
  totalCell: false,
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;

export const Accent1 = Template.bind({});
Accent1.args = {
  ...defaultArgs,
  colorVariant: TagColor.accent1,
  locked: true,
};

export const Total = Template.bind({});
Total.args = {
  ...defaultArgs,
  totalCell: true,
};
