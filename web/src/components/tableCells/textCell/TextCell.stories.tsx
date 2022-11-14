import { Meta, Story } from "@storybook/react";
import TextCellMemo from "./TextCell";

export default {
  component: TextCellMemo,
} as Meta;

export const Primary: Story = (args) => <TextCellMemo current={"something"} />;
Primary.args = {
  current: "Some text",
};
