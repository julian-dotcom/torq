import RawSelect, { SelectOptionType } from "components/forms/select/Select";
import { Props } from "react-select";

type selectProps = {
  label: string;
} & Props;

export type SelectOptions = {
  label?: string;
  value: number | string;
};

function Select(props: selectProps) {
  return (
    <div style={{ marginBottom: "var(--form-margin-bottom)" }}>
      <div style={{ marginBottom: "var(--form-label-margin-bottom)" }}>
        <span>{props.label}</span>
      </div>
      <RawSelect options={props.options} value={props.value} onChange={props.onChange} />
    </div>
  );
}

export default Select;
export type SelectOption = SelectOptionType;
