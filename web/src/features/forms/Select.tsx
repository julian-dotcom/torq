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
  return <RawSelect label={props.label} options={props.options} value={props.value} onChange={props.onChange} />;
}

export default Select;
export type SelectOption = SelectOptionType;
