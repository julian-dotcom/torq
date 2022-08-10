import Select, { Props, components } from "react-select";
import { ChevronDown20Regular as ChevronDownIcon } from "@fluentui/react-icons";

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
    justifyContent: "center",
  }),
  dropdownIndicator: (provided: any, state: any) => ({
    ...provided,
    color: "var(--content-default)",
    // padding: "0",
    height: "32px",
    width: "32px",
  }),
  control: (provided: any, state: any) => ({
    ...provided,
    height: "32px",
    minHeight: "30px",
    border: "1px solid transparent",
    borderBottom: "1px solid var(--primary-250)",
    cursor: "pointer",
    backgroundColor: "transparent",
    boxShadow: "none",
    "&:hover": {
      color: "var(--content-default)",
      border: "1px solid var(--primary-250)",
      boxShadow: "none",
      borderRadius: "2px",
    },
  }),
  singleValue: (provided: any) => ({
    ...provided,
    fontSize: "var(--font-size-default)",
  }),
  option: (provided: any, state: any) => ({
    ...provided,
    color: "var(--content-default)",
    background: state.isSelected ? "var(--primary-150)" : "#ffffff",
    "&:hover": {
      boxShadow: "none",
      background: "var(--primary-50)",
    },
    fontSize: "var(--font-size-default)",
  }),
  menuList: (provided: any, state: any) => ({
    ...provided,
    // background: "",
    border: "1px solid var(--primary-150)",
    boxShadow: "none",
    borderRadius: "2px",
  }),
  menu: (provided: any, state: any) => ({
    ...provided,
    boxShadow: "none",
    zIndex: 10,
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
