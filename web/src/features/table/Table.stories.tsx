import { Meta, Story } from "@storybook/react";
import Table from "./Table";
import { TableProps } from "./types";
import { useArgs } from "@storybook/client-api";

export default {
  title: "features/table/Table",
  component: Table,
} as Meta;

const Template: Story<TableProps<ExampleData>> = (args) => {
  const [_, updateArgs] = useArgs();

  // const onClickHandler = () => {
  //   updateArgs({ checked: !args.checked });
  // };

  return <Table {...args} />;
};

function getExampleRow(id: number) {
  return {
    id: id,
    name: "Example",
    amount: Math.floor(Math.random() * 10000),
    balance: Math.random() * 100,
    date: new Date(),
    duration: 3660,
    active: Math.random() < 0.5,
    // checkbox: Math.random() < 0.5,
  };
}

const exData = Array(20)
  .fill(0)
  .map((_, index) => {
    return getExampleRow(index);
  });

type ExampleData = typeof exData[0];

const defaultArgs: TableProps<ExampleData> = {
  cellRenderer: (row, rowIndex, columnMeta, columnIndex) => {
    return <div>Hello</div>;
  },
  activeColumns: [
    {
      heading: "Name",
      type: "TextCell",
      key: "name",
      valueType: "string",
    },
    {
      heading: "Id",
      type: "NumberCell",
      key: "id",
      valueType: "number",
    },
    {
      heading: "Amount",
      type: "NumberCell",
      key: "amount",
      valueType: "number",
    },
    {
      heading: "Balance",
      type: "BarCell",
      key: "balance",
      valueType: "number",
      max: 100,
    },
    {
      heading: "Date",
      type: "DateCell",
      key: "date",
      valueType: "date",
    },
    {
      heading: "Duration",
      type: "DurationCell",
      key: "duration",
      valueType: "duration",
    },
    {
      heading: "Active",
      type: "BooleanCell",
      key: "active",
      valueType: "boolean",
    },
  ],
  data: exData,
  selectedRowIds: [1, 3, 5],
  isLoading: false,
  // showTotals: false,
  selectable: true,
};

export const Primary = Template.bind({});
Primary.args = defaultArgs;
