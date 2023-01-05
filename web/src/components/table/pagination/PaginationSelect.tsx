import Select, { Props, components } from "react-select";
import { ChevronDown20Regular as ChevronDownIcon } from "@fluentui/react-icons";

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
    justifyContent: "center",
  }),
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  dropdownIndicator: (provided: any, _: any) => ({
    ...provided,
    color: "var(--content-default)",
    // padding: "0",
    height: "32px",
    width: "32px",
  }),
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  control: (provided: any, _: any) => ({
    ...provided,
    height: "32px",
    minHeight: "30px",
    border: "1px solid transparent",
    borderBottom: "1px solid var(--primary-250)",
    cursor: "pointer",
    backgroundColor: "transparent",
    boxShadow: "none",
    borderRadius: "0px",
    "&:hover": {
      color: "var(--content-default)",
      border: "1px solid var(--primary-250)",
      boxShadow: "none",
      borderRadius: "2px",
    },
  }),
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  singleValue: (provided: any) => ({
    ...provided,
    fontSize: "var(--font-size-default)",
  }),
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
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
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  menuList: (provided: any, _: any) => ({
    ...provided,
    // background: "",
    border: "1px solid var(--primary-150)",
    boxShadow: "none",
    borderRadius: "2px",
  }),
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  menu: (provided: any, _: any) => ({
    ...provided,
    boxShadow: "none",
    zIndex: 10,
  }),
};

export default function TorqSelect(props: Props) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const DropdownIndicator = (props: any) => {
    return (
      <components.DropdownIndicator {...props}>
        <ChevronDownIcon />
      </components.DropdownIndicator>
    );
  };

  return <Select components={{ DropdownIndicator }} styles={customStyles} {...props} />;
}
