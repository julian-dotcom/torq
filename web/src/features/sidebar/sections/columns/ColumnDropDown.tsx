import Select, { Props, components } from "react-select";

import { ChevronDown16Regular as ChevronDownIcon } from "@fluentui/react-icons";

export type SelectOptionType = { value: string; label: string };

const customStyles = {
  indicatorSeparator: () => {
    return {};
  },
  input: (provided: any, state: any) => ({
    ...provided,
    borderRadius: "0px",
    padding: "0",
    margin: "0",
  }),
  valueContainer: (provided: any, state: any) => ({
    ...provided,
    paddingLeft: "8px",
    paddingRight: "8px",
    borderRadius: "0px",
  }),
  dropdownIndicator: (provided: any, state: any) => ({
    ...provided,
    color: "var(--secondary-1-500)",
    // padding: "0",
  }),
  control: (provided: any, state: any) => ({
    ...provided,
    border: "1px solid transparent",
    borderRadius: "2px",
    boxShadow: "none",
    backgroundColor: "white",
    minHeight: "34px",
    "&:hover": {
      border: "1px solid transparent",
      // backgroundColor: "var(--secondary-1-100)",
      boxShadow: "none",
      borderRadius: "0px",
    },
  }),
  singleValue: (provided: any) => ({
    ...provided,
    // fontSize: "var(--font-size-small)",
  }),
  option: (provided: any, state: any) => ({
    ...provided,
    color: "var(--content-default)",
    // background: state.isFocused ? "var(--secondary-1-50)" : "#ffffff",
    background: state.isSelected ? "var(--secondary-1-100)" : "#ffffff",
    // background: state.isOptionSelected ? "var(--primary-150)" : "#ffffff",
    "&:hover": {
      boxShadow: "none",
      backgroundColor: "var(--secondary-1-50)",
    },
    // fontSize: "var(--font-size-small)",
  }),
  menuList: (provided: any, state: any) => ({
    ...provided,
    // background: "",
    border: "1px solid var(--secondary-1-150)",
    boxShadow: "none",
    borderRadius: "2px",
    // background: "var(--secondary-1-500)",
  }),
  menu: (provided: any, state: any) => ({
    ...provided,
    boxShadow: "none",
  }),
  container: (provided: any, state: any) => ({
    ...provided,
    border: "1px solid var(--secondary-1-100)",
    "&:hover": {
      border: "1px solid var(--secondary-1-150)",
      // backgroundColor: "var(--secondary-1-100)",
      boxShadow: "none",
      borderRadius: "2px",
    },
  }),
};

type TorqSelectProps = Props & {
  styles?: any;
};

export default function TorqSelect(props: Props) {
  const DropdownIndicator = (props: any) => {
    return (
      <components.DropdownIndicator {...props}>
        <ChevronDownIcon />
      </components.DropdownIndicator>
    );
  };
  return <Select components={{ DropdownIndicator }} styles={customStyles} {...props} />;
}
