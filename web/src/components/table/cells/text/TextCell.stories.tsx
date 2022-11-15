import { Meta, Story } from "@storybook/react";
import TextCellMemo, { TextCellProps } from "./TextCell";

export default {
  title: "components/table/cells/TextCell",
  component: TextCellMemo,
} as Meta;

const Template: Story<TextCellProps> = (args) => <TextCellMemo {...args} />;

export const Primary = Template.bind({});
Primary.args = {
  current: "Some value text",
  total: false,
};

export const NoData = Template.bind({});
NoData.args = { current: undefined };

export const Total = Template.bind({});
Total.args = {
  current: "Some total value text",
  total: true,
};
