import styled from "@emotion/styled";
import Select, { Props } from "react-select";

export type SelectOptionType = { value: string; label: string };

export const StyledSelect = styled(Select)`
  background: #ffffff;
`;

const customStyles = {
  indicatorSeparator: () => {
    return {};
  },
  control: (provided: any, state: any) => ({
    ...provided,
    borderRadius: 2,
    border: "1px solid transparent",
    boxShadow: "none",
    "&:hover": {
      border: "1px solid #8198a3",
      boxShadow: "none",
    },
  }),
  singleValue: (provided: any) => ({
    ...provided,
    // fontSize: "var(--font-size-small)",
  }),
  option: (provided: any, state: any) => ({
    ...provided,
    color: "#8198a3",
    background: state.isFocused ? "var(--primary-150)" : "#ffffff",
    "&:hover": {
      boxShadow: "none",
      backgroundColor: "var(--primary-100)",
    },
    // fontSize: "var(--font-size-small)",
  }),
  menuList: (provided: any, state: any) => ({
    ...provided,
    background: "#f9fafb",
  }),
  menu: (provided: any, state: any) => ({
    ...provided,
    zIndex: "10",
  }),
};

type TorqSelectProps = Props & {
  styles?: any;
};

export default function TorqSelect(props: Props) {
  return <StyledSelect styles={customStyles} {...props} />;
}
