import { Meta, Story } from "@storybook/react";
// import { STORYBOOK_CATEGORIES } from "@storybook/react/categories";
import { useArgs } from "@storybook/client-api";
import Pagination, { PaginationProps } from "./Pagination";

export default {
  title: "components/table/pagination/Pagination",
  component: Pagination,
} as Meta;

const Template: Story<PaginationProps> = (args) => {
  const [_, updateArgs] = useArgs();

  const perPageHandler = (limit: number) => {
    updateArgs({ limit });
  };

  const offsetHandler = (offset: number) => {
    updateArgs({ offset });
  };
  return <Pagination {...args} offsetHandler={offsetHandler} perPageHandler={perPageHandler} />;
};

const defaultArgs = {
  limit: 100,
  offset: 0,
  total: 120000,
};

export const Primary = Template.bind({});
Primary.args = {
  ...defaultArgs,
};

// export const NoData = Template.bind({});
// NoData.args = { current: undefined };
//
// export const Total = Template.bind({});
// Total.args = {
//   current: "Some total value text",
//   total: true,
// };
