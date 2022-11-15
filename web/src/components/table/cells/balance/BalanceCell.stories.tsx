import { Meta, Story } from "@storybook/react";
import BalanceCell, { BalanceCellProps } from "./BalanceCell";

export default {
  title: "components/table/cells/BalanceCell",
  component: BalanceCell,
} as Meta;

const Template: Story<BalanceCellProps> = (args) => <BalanceCell {...args} />;

const defaultArgs: BalanceCellProps = {
  capacity: 100000,
  remote: 40000,
  local: 60000,
  totalCell: false,
};

export const Primary = Template.bind({});
Primary.args = {
  ...defaultArgs,
};

export const Total = Template.bind({});
Total.args = {
  ...defaultArgs,
  totalCell: true,
};
