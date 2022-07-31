import { useEffect, useState } from "react";
import { DragDropContext, Droppable, Draggable } from "react-beautiful-dnd";
import {
  Dismiss20Regular as DismissIcon,
  ReOrder16Regular as ReOrderIcon,
  AddSquare20Regular as AddIcon,
} from "@fluentui/react-icons";
import TorqSelect from "features/inputs/Select";
import DefaultButton from "features/buttons/Button";
import { ColumnMetaData } from "features/forwards/tableSlice";
// TODO: Convert to styled components using imported scss
import styles from "./sort.module.scss";
import classNames from "classnames";
import { ActionMeta } from "react-select";

export interface SortByOptionType {
  value: string;
  label: string;
  direction: string;
}

interface sortRowInterface {
  selected: SortByOptionType;
  options: SortByOptionType[];
  index: number;
  handleUpdateSort: Function;
  handleRemoveSort: Function;
}
interface sortDirectionOption {
  value: string;
  label: string;
}
const sortDirectionOptions: sortDirectionOption[] = [
  { value: "asc", label: "Ascending" },
  { value: "desc", label: "Descending" },
];

const SortRow = ({ selected, options, index, handleUpdateSort, handleRemoveSort }: sortRowInterface) => {
  const handleColumn = (newValue: unknown, actionMeta: ActionMeta<unknown>) => {
    handleUpdateSort(newValue, index);
  };
  const handleDirection = (newValue: unknown, actionMeta: ActionMeta<unknown>) => {
    handleUpdateSort(
      {
        ...selected,
        direction: (newValue as { value: string }).value,
      },
      index
    );
  };

  return (
    <Draggable draggableId={`draggable-sort-id-${index}`} index={index}>
      {(provided, snapshot) => (
        <div
          className={classNames(styles.sortRow, {
            dragging: snapshot.isDragging,
          })}
          ref={provided.innerRef}
          {...provided.draggableProps}
        >
          <div {...provided.dragHandleProps}>
            <ReOrderIcon />
          </div>
          <div style={{ flex: 3 }}>
            <TorqSelect
              onChange={handleColumn}
              options={options}
              getOptionLabel={(option: any): string => option["label"]}
              getOptionValue={(option: any): string => option["value"]}
              value={selected}
            />
          </div>

          <div style={{ flex: 2 }}>
            <TorqSelect
              onChange={handleDirection}
              options={sortDirectionOptions}
              getOptionLabel={(option: any): string => option["label"]}
              getOptionValue={(option: any): string => option["value"]}
              value={sortDirectionOptions.find((dir: sortDirectionOption) => dir.value === selected.direction)}
            />
          </div>
          <DismissIcon
            onClick={() => {
              handleRemoveSort(index);
            }}
          />
        </div>
      )}
    </Draggable>
  );
};

type SortSectionProps = {
  columns: Array<ColumnMetaData>;
  sortBy: Array<SortByOptionType>;
  updateSortByHandler: (sortBy: Array<SortByOptionType>) => void;
};

const SortSection = (props: SortSectionProps) => {
  // const dispatch = useAppDispatch();
  // const columns = useAppSelector(selectAllColumns);
  // const sorts = useAppSelector(selectSortBy);

  let options = props.columns.slice().map((column: { key: string; heading: string; valueType: string }) => {
    return {
      value: column.key,
      label: column.heading,
      direction: "desc",
    };
  });

  const handleAddSort = () => {
    const updated: SortByOptionType[] = [
      ...props.sortBy,
      {
        direction: "desc",
        value: props.columns[0].key,
        label: props.columns[0].heading,
      },
    ];
    props.updateSortByHandler(updated);
  };

  const handleUpdateSort = (update: SortByOptionType, index: number) => {
    const updated: SortByOptionType[] = [
      ...props.sortBy.slice(0, index),
      update,
      ...props.sortBy.slice(index + 1, props.sortBy.length),
    ];
    props.updateSortByHandler(updated);
  };

  const handleRemoveSort = (index: number) => {
    const updated: SortByOptionType[] = [
      ...props.sortBy.slice(0, index),
      ...props.sortBy.slice(index + 1, props.sortBy.length),
    ];
    props.updateSortByHandler(updated);
  };

  const buttonText = (): string => {
    if (props.sortBy.length > 0) {
      return props.sortBy.length + " Sorted";
    }
    return "Sort";
  };

  const droppableContainerId = "sort-list-droppable";

  const onDragEnd = (result: any) => {
    const { destination, source, draggableId } = result;

    // Dropped outside of container
    if (!destination || destination.droppableId !== droppableContainerId) {
      return;
    }

    // Position not changed
    if (destination.droppableId === source.droppableId && destination.index === source.index) {
      return;
    }

    let newSortsOrder = props.sortBy.slice();
    newSortsOrder.splice(source.index, 1);
    newSortsOrder.splice(destination.index, 0, props.sortBy[source.index]);
    props.updateSortByHandler(newSortsOrder);
  };

  return (
    <DragDropContext
      // onDragStart={}
      // onDragUpdate={}
      onDragEnd={onDragEnd}
    >
      <div className={styles.sortPopoverContent}>
        {!props.sortBy.length && <div className={styles.noFilters}>No sorting</div>}

        {!!props.sortBy.length && (
          <Droppable droppableId={droppableContainerId}>
            {(provided) => (
              <div className={styles.sortRows} ref={provided.innerRef} {...provided.droppableProps}>
                {props.sortBy.map((item, index) => {
                  return (
                    <SortRow
                      key={item.value + index}
                      selected={item}
                      options={options}
                      index={index}
                      handleUpdateSort={handleUpdateSort}
                      handleRemoveSort={handleRemoveSort}
                    />
                  );
                })}

                {provided.placeholder}
              </div>
            )}
          </Droppable>
        )}

        <div className={styles.buttonsRow}>
          <DefaultButton onClick={() => handleAddSort()} text={"Add Sort"} icon={<AddIcon />} />
        </div>
      </div>
    </DragDropContext>
  );
};

export default SortSection;
