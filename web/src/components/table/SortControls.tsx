//@ts-nocheck

import { useEffect, useState } from "react";
import styled from "@emotion/styled";
import { Popover } from "react-tiny-popover";
import {
  ArrowSortDownLines20Regular as SortIcon,
  Dismiss20Regular as Dismiss,
  ReOrder16Regular as ReOrder,
  AddSquare20Regular
} from "@fluentui/react-icons";
import * as _ from "lodash";

import TorqSelect from "../inputs/Select";
import DefaultButton from "../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../store/hooks";
import { selectColumns, selectSorts, updateSort, updateSortOptions, selectSortByOptions  } from "./tableSlice";

const ControlsContainer = styled.div({
  color: "#033048",
  width: 451,
  background: "#ffffff",
  padding: 20,
  borderRadius: 2,
  boxShadow: "0px 0px 10px rgba(3, 48, 72, 0.1)",
  position: "fixed",
  overflow: "visible",
  top: 0,
  left: 0,
  transform: "translate(407px, 80px)",
  zIndex: 999
});

const StyledSortIcon = styled(ReOrder)({
  marginTop: 10
});

const StyledDismissIcon = styled(Dismiss)({
  marginTop: 10
});

const StyledAddSort = styled(AddSquare20Regular)({
  marginTop: 10
});

const Container = styled.div`
  display: flex;
  margin-bottom: 10px;
  grid-column-gap: 10px;
`;

const sortOptions = [
  { value: "", label: "Ascending" },
  { value: "-", label: "Descending" }
];

const AddSortButton = styled.button({
  marginTop: 8,
  color: "#033048"
});

function getDifference(array1: any[], array2: { key: any }[]) {
  return array1.filter((object1: { key: any }) => {
    return !array2.some((object2: { key: any }) => {
      return object1.key === object2.key;
    });
  });
}

const SortRow = (props: any) => {
  const [sortKey, setSortKey] = useState(props.options[0].key);
  const [sortBy, setSortBy] = useState("");

  const { handleSorts } = props;
  useEffect(() => {
    handleSorts({ key: sortKey, sortBy: sortBy });
  }, [sortBy, sortKey]);
  return (
    <>
      <StyledSortIcon />
      <div style={{ flex: 3 }}>
        <TorqSelect
          onChange={(v: any) => setSortKey(v.key)}
          defaultValue={props.options[0]}
          options={props.options}
          getOptionLabel={(option: { [x: string]: any }) =>
            option[props.labelValue]
          }
          getOptionValue={(option: { [x: string]: any }) =>
            option[props.keyValue]
          }
        />
      </div>

      <div style={{ flex: 2 }}>
        <TorqSelect
          onChange={(v: any) => setSortBy(v.value)}
          defaultValue={sortOptions[0]}
          options={sortOptions}
        />
      </div>
      <StyledDismissIcon onClick={() => props.handleDelete(props.id)} />
    </>
  );
};

const SortControls = (props: any) => {
  const dispatch = useAppDispatch();
  const columns = useAppSelector(selectColumns) || [];
  const sorts = useAppSelector(selectSorts) || [];
  const sortOptions = useAppSelector(selectSortByOptions) || [];

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  // const [sortOptions, setSortOptions] = useState(columns);
  // const [sorts, setSorts] = useState<any[]>([]);
  const [sortBy, setSortBy] = useState<any[]>([]);

  const handleAddSort = () => {
    const newOptions = getDifference(columns, sorts);
    // setSortOptions(newOptions);
    //@ts-ignore
    // setSorts(prev => {
    //   return [...prev, newOptions[0]];
    // });

    dispatch(updateSortOptions([[...sorts, newOptions[0]], [newOptions]]))
  };

  const handleSorts = (i: any) => {
    const newArr = [...sortBy];
    let objIndex = newArr.findIndex(obj => obj.key === i.key);

    if (objIndex >= 0) {
      //@ts-ignore
      newArr[objIndex].sortBy = i.sortBy;
    } else {
      //@ts-ignore
      newArr.push(i);
    }



    const ba: string[] = newArr.map(i => `${i.sortBy}${i.key}`);


    setSortBy(newArr);

    dispatch(updateSort(ba));
  };

  const handleRemoveSort = (key: any) => {
    const newArray = sorts.filter(function(element: { key: string }) {
      return element.key !== key;
    });
    // setSorts(newArray);
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
            {sorts?.map(sort => {
              return (
                <Container key={sort.key}>
                  <SortRow
                    handleSorts={handleSorts}
                    handleDelete={handleRemoveSort}
                    keyValue="key"
                    labelValue="heading"
                    options={sortOptions}
                    id={sort.key}
                  />
                </Container>
              );
            })}

            {columns.length === sorts.length ? null : (
              <Container>
                <StyledAddSort />
                <AddSortButton onClick={() => handleAddSort()}>
                  Add Row
                </AddSortButton>
              </Container>
            )}
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

export default SortControls;
