import { useEffect, useState } from "react";
import styled from "@emotion/styled";
import {
  ArrowSortDownLines20Regular as SortIcon,
  Dismiss20Regular as Dismiss,
  ReOrder16Regular as ReOrder,
  AddSquare20Regular,
} from "@fluentui/react-icons";

import TorqSelect from "../../../inputs/Select";
import DefaultButton from "../../../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { selectAllColumns, selectSorts, updateSort, updateSortOptions, selectSortByOptions, } from "../../tableSlice";
import Popover from '../../../popover/Popover';

const ControlsContainer = styled.div({
  width: 451,
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
  const columns = useAppSelector(selectAllColumns) || [];
  const sorts: any[] = useAppSelector(selectSorts) || [];
  const sortOptions = useAppSelector(selectSortByOptions) || [];

  const [sortBy, setSortBy] = useState<any[]>([]);

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  const handleAddSort = () => {
    dispatch(updateSortOptions([[...sorts, columns[sorts.length+1]], columns]))
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
    setSortBy(newArr)
    const ba: string[] = newArr.map(i => `${i.sortBy}${i.key}`);
    dispatch(updateSort(ba));
  };

  const handleRemoveSort = (key: any) => {
    const newArray = sorts.filter(function(element: { key: string }) {
      return element.key !== key;
    });

    dispatch(updateSortOptions([[...newArray], columns]))
  };

  let popOverButton = <DefaultButton
          isOpen={isPopoverOpen}
          onClick={() => setIsPopoverOpen(!isPopoverOpen)}
          icon={<SortIcon />}
          text={"Sort"}
          className={"collapse-tablet"}
        />

  return (
    <div>
      <Popover button={popOverButton}>
        <ControlsContainer>
            {!sorts.length && <div className={"no-filters"}>No sorting</div>}
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
                  Add Sort
                </AddSortButton>
              </Container>
            )}
          </ControlsContainer>

      </Popover>
    </div>
  );
};

export default SortControls;
