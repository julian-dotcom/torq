import RawSelect, { SelectOptionType } from "../inputs/Select";
import { Props } from "react-select";

type selectProps = {
  label: string;
} & Props;

function Select(props: selectProps) {
  return (
    <div style={{ marginBottom: "15px" }}>
      <div style={{ marginBottom: "8px" }}>
        <span>{props.label}</span>
      </div>
      <RawSelect options={props.options} value={props.value} onChange={props.onChange} />
    </div>
  );
}

export default Select;
export type SelectOption = SelectOptionType;
