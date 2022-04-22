import styled from "@emotion/styled";
import Select from "react-select";


export type SelectOptionType = {value: string, label:string}

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
    border: "1px solid #D9E0E4",
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
    background: state.isFocused ? "#D9E0E4" : "#f9fafb",
    "&:hover": {
      boxShadow: "none",
      backgroundColor: "#D9E0E4",
    },
    // fontSize: "var(--font-size-small)",
  }),
  menuList: (provided: any, state: any) => ({
    ...provided,
    background: "#f9fafb",
  }),
};

export default function TorqSelect(props: any) {
  return <StyledSelect styles={customStyles} {...props} />;
}
