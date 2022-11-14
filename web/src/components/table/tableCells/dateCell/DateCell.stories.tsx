import { Meta, Story } from "@storybook/react";
import DateCellMemo, { DateCellProps } from "./DateCell";

export default {
  title: "components/table/tableCells/dateCell/DateCell",
  component: DateCellMemo,
} as Meta;

const Template: Story<DateCellProps> = (args) => <DateCellMemo {...args} />;

export const Primary = Template.bind({});
Primary.args = {
  value: new Date("2022-01-01Z00:21:21"),
};

export const NoData = Template.bind({});
NoData.args = { value: undefined };

// export const Danger = Template.bind({});
// Danger.args = {
//   children: "Danger",
//   variant: "danger",
//   shape: "rounded",
// };

// export const Primary: Story = (args) => <DateCellMemo value={new Date("2022-01-01Z00:21:21")} />;
//
// export const NoData: Story = (args) => <DateCellMemo value={new Date("2022-01-01Z00:21:21")} />;
