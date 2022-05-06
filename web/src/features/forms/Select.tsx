import RawSelect, { SelectOptionType } from "../inputs/Select";

interface selectProps {
  label: string;
  options: SelectOptionType[];
  value: SelectOptionType | undefined;
  onChange: Function;
}

function Select(props: selectProps) {
  return (
    <div style={{ marginBottom: "15px" }}>
      <div style={{ marginBottom: "8px" }}>
        <span>{props.label}</span>
      </div>
      <RawSelect
        options={props.options}
        value={props.value}
        onChange={props.onChange}
      />
    </div>
  );
}

export default Select;
export type SelectOption = SelectOptionType;
