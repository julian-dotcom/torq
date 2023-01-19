import classNames from "classnames";
import { Delete16Regular as RemoveIcon } from "@fluentui/react-icons";
import { FilterCategoryType, FilterClause, FilterFunctions, FilterInterface, FilterParameterType } from "./filter";
import { Clause, FilterClause, FilterInterface, FilterParameterType } from "./filter";
import styles from "./filter-section.module.scss";
import { useState } from "react";
import { format } from "d3";
import { Input, InputColorVaraint, InputSizeVariant, Select } from "components/forms/forms";
import { TagResponse } from "pages/tags/tagsTypes";
import { useGetTagsQuery } from "pages/tags/tagsApi";
import clone from "clone";

const formatter = format(",.0f");

function formatParameter(value: number) {
  return value % 1 != 0 ? value : formatter(value);
}

const ffLabels = new Map<string, string>([
  ["eq", "is equal to"],
  ["neq", "is not equal to"],
  ["gt", "is greater than"],
  ["gte", "is greater than or equal to"],
  ["lt", "is less than"],
  ["lte", "is less than or equal to"],
  ["like", "contain"],
  ["notLike", "don't contain"],
  ["any", "contain"],
  ["notAny", "don't contain"],
]);

function getFilterFunctions(filterCategory: FilterCategoryType) {
  const filterFuncs = FilterFunctions.get(filterCategory)?.entries();
  if (!filterFuncs) {
    throw new Error("Filter category not found in list of filters");
  }
  return Array.from(filterFuncs, ([key, _]) => ({ value: key, label: ffLabels.get(key) ?? "Label not found" }));
}

interface filterRowInterface {
  index: number;
  child: boolean;
  filterClause: FilterClause;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  filterOptions: Array<{ label: string; value: any; valueType?: FilterCategoryType; selectOptions?: Array<any> }>;
  onUpdateFilter: (filters: Clause) => void;
  onRemoveFilter: (index: number) => void;
  handleCombinerChange: () => void;
  combiner?: string;
}

function FilterRow({
  index,
  child,
  filterClause,
  filterOptions,
  onUpdateFilter,
  onRemoveFilter,
  handleCombinerChange,
  combiner,
}: filterRowInterface) {
  const rowValues = filterClause.filter;
  const [rowExpanded, setRowExpanded] = useState(false);
  const { data: getTags } = useGetTagsQuery();

  const tags = (getTags || []).map((tag: TagResponse) => ({ label: tag.name, value: tag.tagId })) || [];

  const functionOptions = getFilterFunctions(rowValues.category);

  const keyOption = filterOptions.find((item) => item.value === rowValues.key);
  const funcOption = functionOptions.find((item) => item.value === rowValues.funcName);

  const selectData = {
    func: funcOption,
    key: keyOption,
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleKeyChange = (item: any) => {
    const newRow = { ...rowValues };
    newRow.key = item.value;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const newCategory = filterOptions.find((item: any) => item.value === newRow.key)?.valueType || "number";
    switch (newCategory) {
      case "number":
        newRow.parameter = 0;
        newRow.funcName = "gte";
        break;
      case "boolean":
        if (newCategory !== newRow.category) {
          newRow.parameter = true;
        }
        newRow.funcName = "eq";
        break;
      case "date": {
        newRow.parameter = new Date().toISOString().slice(0, 10) + "T00:00:00";
        newRow.funcName = "gte";
        break;
      }
      case "array":
        newRow.parameter = "";
        newRow.funcName = "eq";
        break;
      case "tag":
        newRow.parameter = [];
        newRow.funcName = "any";
        break;
      default:
        newRow.parameter = "";
        newRow.funcName = "like";
    }
    newRow.category = newCategory;
    const updatedFilterClause = clone(filterClause);
    updatedFilterClause.filter = newRow;
    onUpdateFilter(updatedFilterClause);
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleFunctionChange = (item: any) => {
    const updatedFilterClause = clone(filterClause);
    updatedFilterClause.filter = { ...rowValues, funcName: item.value };
    onUpdateFilter(updatedFilterClause);
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const handleParamChange = (e: any) => {
    const newRow = { ...rowValues };
    switch (newRow.category) {
      case "number":
        newRow.parameter = e.floatValue || 0;
        break;
      case "boolean":
        newRow.parameter = e.value;
        break;
      case "tag":
        newRow.parameter = (e as Array<{ value: number }>).map((item) => item.value);
        break;
      case "array":
        newRow.parameter = String(e.value);
        break;
      case "enum":
        newRow.parameter = String(e.value);
        break;
      default:
        newRow.parameter = e.target.value ? e.target.value : "";
    }
    const updatedFilterClause = clone(filterClause);
    updatedFilterClause.filter = newRow;
    onUpdateFilter(updatedFilterClause);
  };

  const label = filterOptions.find((item) => item.value === rowValues.key)?.label;
  const options = filterOptions.find((item) => item.value === rowValues.key)?.selectOptions || [];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any

  const getParameter = () => {
    switch (rowValues.category) {
      case "number":
        return formatParameter(rowValues.parameter as number);
      case "duration":
        return formatParameter(rowValues.parameter as number);
      case "boolean":
        return !rowValues.parameter ? "False" : "True";
      case "array":
        return options?.find((item) => {
          return item.value === rowValues.parameter ? item : "";
        })?.label;
      case "tag":
        if (rowValues?.parameter !== undefined) {
          const parameter = rowValues.parameter as Array<number>;
          return (parameter || []).map((tagId) => (tags || []).find((t) => t.value === tagId)?.label).join(", ");
        }
        return "";
      default:
        return rowValues.parameter;
    }
  };

  return (
    <div className={classNames(styles.filter, { [styles.first]: !index })}>
      <div className={styles.filterKeyContainer}>
        {index !== 0 && (
          <div
            className={styles.combinerContainer}
            onClick={() => {
              handleCombinerChange();
            }}
          >
            {combiner}
          </div>
        )}
        <div className={styles.filterKeyLabel} onClick={() => setRowExpanded(!rowExpanded)}>
          <span>{label}</span>
          <span className={styles.filterFunctionLabel}>&nbsp;{funcOption?.label}&nbsp;</span>
          <span className={styles.parameterLabel}>{getParameter()}</span>
        </div>
        <div className={classNames(styles.removeFilter, styles.desktopRemove)} onClick={() => onRemoveFilter(index)}>
          <RemoveIcon />
        </div>
      </div>

      <div className={classNames(styles.filterOptions, { [styles.expanded]: rowExpanded })}>
        <Select
          options={filterOptions}
          value={selectData.key}
          onChange={handleKeyChange}
          sizeVariant={InputSizeVariant.small}
          colorVariant={child ? InputColorVaraint.primary : InputColorVaraint.accent1}
        />

        <div className="filter-function-container">
          <Select
            options={functionOptions}
            value={selectData.func}
            onChange={handleFunctionChange}
            sizeVariant={InputSizeVariant.small}
            colorVariant={child ? InputColorVaraint.primary : InputColorVaraint.accent1}
          />
        </div>

        <div className={styles.filterParameterContainer}>
          <FilterInputField onChange={handleParamChange} rowValues={rowValues} options={options} child={child} />
        </div>
      </div>
    </div>
  );
}

export default FilterRow;

function FilterInputField(props: {
  onChange: (e: any) => void; // eslint-disable-line @typescript-eslint/no-explicit-any
  rowValues: FilterInterface;
  child: boolean;
  options: Array<any>; // eslint-disable-line @typescript-eslint/no-explicit-any
}) {
  const { data: getTags } = useGetTagsQuery();
  const tags = (getTags || []).map((tag: TagResponse) => ({ label: tag.name, value: tag.tagId })) || [];
  switch (props.rowValues.category) {
    case "number":
      return (
        <Input
          formatted={true}
          className={classNames(styles.filterInputField, styles.small)}
          thousandSeparator=","
          defaultValue={props.rowValues.parameter as keyof FilterParameterType}
          onValueChange={(e) => {
            props.onChange(e);
          }}
        />
      );
    case "boolean":
      return (
        <Select
          options={[
            { label: "True", value: true },
            { label: "False", value: false },
          ]}
          value={props.rowValues.parameter}
          onChange={props.onChange}
          colorVariant={props.child ? InputColorVaraint.primary : InputColorVaraint.accent1}
          sizeVariant={InputSizeVariant.small}
        />
      );
    case "array": {
      const label = props.options?.find((item) => {
        return item.value === props.rowValues.parameter ? item : "";
      })?.label;
      return (
        <Select
          options={props.options}
          value={{ value: props.rowValues.parameter, label: label }}
          onChange={props.onChange}
          colorVariant={props.child ? InputColorVaraint.primary : InputColorVaraint.accent1}
          sizeVariant={InputSizeVariant.small}
        />
      );
    }
    case "tag": {
      return (
        <Select
          isMulti={true}
          options={tags}
          value={((props.rowValues.parameter as Array<number>) || []).map((tagId) => {
            return { label: tags.find((o) => o.value === tagId)?.label, value: tagId };
          })}
          onChange={props.onChange}
          colorVariant={props.child ? InputColorVaraint.primary : InputColorVaraint.accent1}
          sizeVariant={InputSizeVariant.small}
        />
      );
    }
    case "enum": {
      console.log(props.rowValues.parameter);
      return (
        <Select
          options={props.options}
          value={props.rowValues.parameter}
          onChange={props.onChange}
          colorVariant={props.child ? InputColorVaraint.primary : InputColorVaraint.accent1}
          sizeVariant={InputSizeVariant.small}
        />
      );
    }
    case "date":
      return (
        <input
          type="datetime-local"
          className={"torq-input-field"}
          value={props.rowValues.parameter as string}
          onChange={props.onChange}
        />
      );
    default:
      return (
        <input
          type="text"
          className={"torq-input-field"}
          value={props.rowValues.parameter as keyof FilterParameterType}
          onChange={props.onChange}
        />
      );
  }
}
