import { Meta, Story } from "@storybook/react";
import CheckboxCell, { CheckboxCellProps } from "./CheckboxCell";
import { useArgs } from "@storybook/client-api";

export default {
  title: "components/table/cells/CheckboxCell",
  component: CheckboxCell,
} as Meta;

const Template: Story<CheckboxCellProps> = (args) => {
  const [_, updateArgs] = useArgs();

  const onClickHandler = () => {
    updateArgs({ checked: !args.checked });
  };

  return <CheckboxCell {...args} onClick={onClickHandler} />;
};

const defaultArgs: CheckboxCellProps = {
  checked: false,
  wrapperClassNames: undefined,
  totalCell: false,
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;

export const Checked = Template.bind({});
Checked.args = {
  ...defaultArgs,
  checked: true,
};

export const Total = Template.bind({});
Total.args = {
  ...defaultArgs,
  totalCell: true,
};
