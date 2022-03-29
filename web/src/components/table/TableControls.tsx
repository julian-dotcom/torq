import "./table_controls.scss";
import DefaultButton from "../buttons/Button";
import { useAppDispatch } from "../../store/hooks";
import { toggleNav } from "../navigation/navSlice";
import {
  ColumnTriple20Regular as ColumnsIcon,
  ArrowSortDownLines20Regular as SortIcon,
  Filter20Regular as FilterIcon,
  Navigation20Regular as NavigationIcon,
  ArrowJoin20Regular as GroupIcon,
  Search20Regular as SearchIcon,
  Options20Regular as OptionsIcon,
  Dismiss20Regular as Dismiss,
  ReOrder16Regular as ReOrder,
  AddSquare20Regular,
} from "@fluentui/react-icons";
import { Popover } from "react-tiny-popover";
import { useState } from "react";
import styled from "@emotion/styled";
import Select from "react-select";

import TimeIntervalSelect from "../timeIntervalSelect/TimeIntervalSelect";
import Dropdown from "../formElements/Dropdown";
import { fetchChannelsAsync } from "./tableSlice";

const ControlsContainer = styled.div`
  color: #033048;
  width: 451px;
  background: #ffffff;
  padding: 20px;
  border-radius: 2px;
  box-shadow: 0px 0px 10px rgba(3, 48, 72, 0.1);
  position: fixed;
  overflow: visible;
  top: 0px;
  left: 0px;
  transform: translate(407px, 80px); // Needs to be dyamnic
  z-index: 999;
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
    fontSize: 12,
  }),
  option: (provided: any, state: any) => ({
    ...provided,
    color: "#8198a3",
    background: state.isFocused ? "#D9E0E4" : "#f9fafb",
    "&:hover": {
      boxShadow: "none",
      backgroundColor: "#D9E0E4",
    },
  }),
  menuList: (provided: any, state: any) => ({
    ...provided,
    background: "#f9fafb",
  }),
};

const StyledSelect = styled(Select)`
  background: #ffffff;
  flex: 3;
`;

const StyledSelect2 = styled(Select)`
  flex: 2;
`;

const options = [
  { value: "chocolate", label: "Chocolate" },
  { value: "strawberry", label: "Strawberry" },
  { value: "vanilla", label: "Vanilla" },
];

interface Option {
  value: string;
  label: string;
}
interface SelectProps {
  options: Option[];
}

const StyledSortIcon = styled(ReOrder)({
  marginTop: 10,
});

const StyledDismissIcon = styled(Dismiss)({
  marginTop: 10,
});

const StyledAddSort = styled(AddSquare20Regular)({
  marginTop: 10,
});

const SelectRow = (props: SelectProps) => {
  return (
    <>
      <StyledSortIcon />
      <StyledSelect
        defaultValue={props.options[0]}
        styles={customStyles}
        options={props.options}
      />
      <StyledSelect2 styles={customStyles} options={options} />
      <StyledDismissIcon />
    </>
  );
};

const Container = styled.div`
  display: flex;
  margin-bottom: 10px;
  grid-column-gap: 10px;
`;

const Sort = () => {
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  const [sortOptions, setSortOptions] = useState(options);

  const [sorts, setSorts] = useState([options[0]]);

  const handleAddSort = () => {
    // @ts-ignore
    setSorts((prev) => [...prev, options.pop()]);
  };

  const handleRemoveSort = (id) => {
    setSorts((prev) => [...prev.slice(1)]);
  };
  return (
    <div>
      <Popover
        containerStyle={{ zIndex: "999", background: "white", padding: "20" }}
        reposition={true}
        isOpen={isPopoverOpen}
        positions={["bottom"]}
        content={
          <ControlsContainer>
            {sorts.map((sort) => {
              return (
                <Container>
                  <SelectRow options={options} />
                </Container>
              );
            })}

            <Container>
              <StyledAddSort />
              <button style={{ color: "red" }} onClick={() => handleAddSort()}>
                Add Row
              </button>
            </Container>
          </ControlsContainer>
        }
      >
        <DefaultButton
          isOpen={isPopoverOpen}
          onClick={() => setIsPopoverOpen(!isPopoverOpen)}
          icon={<SortIcon />}
          text={"Sort"}
          className={"collapse-tablet"}
        />
      </Popover>
    </div>
  );
};

function TableControls() {
  const dispatch = useAppDispatch();
  dispatch(fetchChannelsAsync());
  return (
    <div className="table-controls">
      <div className="left-container">
        <div className="upper-container">
          <DefaultButton
            icon={<NavigationIcon />}
            text={"Menu"}
            onClick={() => dispatch(toggleNav())}
            className={"show-nav-btn collapse-tablet"}
          />
          <Dropdown />
          <DefaultButton
            icon={<OptionsIcon />}
            text={""}
            className={"collapse-tablet mobile-options"}
          />
        </div>
        <div className="lower-container">
          <DefaultButton
            icon={<ColumnsIcon />}
            text={"Columns"}
            className={"collapse-tablet"}
          />
          <div>
            <Sort />
          </div>
          <DefaultButton
            icon={<FilterIcon />}
            text={"Filter"}
            className={"collapse-tablet"}
          />
          <DefaultButton
            icon={<GroupIcon />}
            text={"Group"}
            className={"collapse-tablet"}
          />
          <DefaultButton
            icon={<SearchIcon />}
            text={"Search"}
            className={"small-tablet"}
          />
        </div>
      </div>
      <div className="right-container">
        <TimeIntervalSelect />
      </div>
    </div>
  );
}

export default TableControls;
