//@ts-nocheck

import DefaultButton from "../../buttons/Button";
import classNames from "classnames";
import {
  Filter20Regular as FilterIcon,
  Dismiss20Regular as RemoveIcon,
  AddSquare20Regular as AddFilterIcon
} from "@fluentui/react-icons";
import { useState } from "react";
import TorqSelect from "../../inputs/Select";

import "./filter_popover.scoped.scss";
import { useAppSelector } from "../../../store/hooks";
import selectColumns from "../tableSlice";
import { FilterFunctions } from "../filter";

import NumberFormat from "react-number-format";

const combinerOptions = [
  { value: "and", label: "And" }
  // { value: "or", label: "Or" },
];

// {
//   filterCategory: 'number',
//   filterName: 'gte',
//   key: "amount_out",
//   parameter: 5000000
// }

const ffLabels = {
  eq: "=",
  neq: "≠",
  gt: ">",
  gte: ">=",
  lt: "<",
  lte: "<=",
  includes: "Include",
  notInclude: "Not include"
};

function getFilterFunctions(filterCategory: "number" | "string") {
  return Object.keys(FilterFunctions[filterCategory]).map((key: string) => {
    // @ts-ignore
    return { value: key, label: ffLabels[key] || key };
  });
}

const FilterRow = (props: { first: boolean }) => {
  let columnsMeta = useAppSelector(selectColumns) || [];

  let columnOptions = columnsMeta.slice().map(column => {
    return { value: column.key, label: column.heading };
  });

  columnOptions.sort((a, b) => {
    if (a.label < b.label) {
      return -1;
    }
    if (a.label > b.label) {
      return 1;
    }
    return 0;
  });

  let functionOptions = getFilterFunctions("number");

  return (
    <div className={classNames("filter-row", { first: props.first })}>
      {props.first && <div className="filter-combiner-container">Where</div>}
      {!props.first && (
        <div className="combiner-container">
          <TorqSelect
            options={combinerOptions}
            defaultValue={combinerOptions[0]}
          />
        </div>
      )}
      <div className="filter-key-container">
        <TorqSelect options={columnOptions} defaultValue={columnOptions[0]} />
      </div>
      <div className="filter-function-container">
        <TorqSelect
          options={functionOptions}
          defaultValue={functionOptions[0]}
        />
      </div>
      <div className="filter-parameter-container">
        <NumberFormat
          className={"torq-input-field"}
          thousandSeparator=","
          defaultValue={1000}
        />
      </div>
      <div className="remove-filter">
        <RemoveIcon />
      </div>
    </div>
  );
};

const FilterButton = () => {
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  return (
    <div
      onClick={() => setIsPopoverOpen(!isPopoverOpen)}
      className={classNames("torq-popover-button-wrapper")}
    >
      <DefaultButton
        text={"Filter"}
        icon={<FilterIcon />}
        className={"collapse-tablet"}
      />
      <div
        className={classNames("popover-wrapper", {
          "popover-open": isPopoverOpen
        })}
        onClick={e => {
          e.stopPropagation();
        }}
      >
        <div className="filter-rows">
          <FilterRow first={true} />
          <FilterRow first={false} />
          <FilterRow first={false} />
        </div>
        <div className="buttons-row">
          <DefaultButton text={"Add filter"} icon={<AddFilterIcon />} />
          <DefaultButton text={"Add filter"} icon={<AddFilterIcon />} />
        </div>
      </div>
    </div>
  );
};

export default FilterButton;
