import { Meta, Story } from "@storybook/react";
import DateCellMemo, { DateCellProps } from "./DateCell";

export default {
  title: "components/table/cells/DateCell",
  component: DateCellMemo,
} as Meta;

const Template: Story<DateCellProps> = (args) => <DateCellMemo {...args} />;

export const Primary = Template.bind({});
Primary.args = {
  value: new Date("2022-01-01Z00:21:21"),
};

export const NoData = Template.bind({});
NoData.args = { value: undefined };

export const Total = Template.bind({});
Total.args = {
  value: new Date("2022-01-01Z00:21:21"),
  totalCell: true,
};
