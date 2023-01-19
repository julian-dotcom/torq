import Select, { Props, components } from "react-select";

import { ChevronDown16Regular as ChevronDownIcon } from "@fluentui/react-icons";

export type SelectOptionType = { value: string; label: string };

type TorqSelectProps = {
  selectProps: Props;
  child: boolean;
};

export default function TorqSelect({ selectProps, child }: TorqSelectProps) {
  const customStyles = {
    indicatorSeparator: () => {
      return {};
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    input: (provided: any, _: any) => ({
      ...provided,
      borderRadius: "0px",
      padding: "0",
      margin: "0",
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    valueContainer: (provided: any, _: any) => ({
      ...provided,
      paddingLeft: "8px",
      paddingRight: "8px",
      borderRadius: "0px",
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    dropdownIndicator: (provided: any, _: any) => ({
      ...provided,
      color: "var(--secondary-1-600)",
      // padding: "0",
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    control: (provided: any, _: any) => ({
      ...provided,
      border: "0px solid transparent",
      cursor: "pointer",
      borderRadius: "2px",
      boxShadow: "none",
      backgroundColor: "white",
      minHeight: "34px",
      "&:hover": {
        border: "0px solid transparent",
        // backgroundColor: "var(--secondary-1-100)",
        boxShadow: "none",
        borderRadius: "0px",
      },
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    singleValue: (provided: any) => ({
      ...provided,
      // fontSize: "var(--font-size-small)",
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    option: (provided: any, state: any) => ({
      ...provided,
      color: "var(--content-default)",
      // background: state.isFocused ? "var(--secondary-1-50)" : "#ffffff",
      background: state.isSelected ? (child ? "var(--primary-150)" : "var(--secondary-1-100)") : "#ffffff",
      // background: state.isOptionSelected ? "var(--primary-150)" : "#ffffff",
      "&:hover": {
        boxShadow: "none",
        background: child ? "var(--primary-50)" : "var(--secondary-1-50)",
      },
      // fontSize: "var(--font-size-small)",
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    menuList: (provided: any, _: any) => ({
      ...provided,
      // background: "",
      border: "1px solid var(--secondary-1-150)",
      boxShadow: "none",
      borderRadius: "2px",
      // background: "var(--secondary-1-500)",
    }),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    menu: (provided: any, _: any) => ({
      ...provided,
      boxShadow: "none",
    }),
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const DropdownIndicator = (props: any) => {
    return (
      <components.DropdownIndicator {...props}>
        <ChevronDownIcon />
      </components.DropdownIndicator>
    );
  };
  return <Select components={{ DropdownIndicator }} styles={customStyles} {...selectProps} />;
}
