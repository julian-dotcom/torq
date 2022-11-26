import { useMemo } from "react";
import { DragDropContext, Draggable, Droppable } from "react-beautiful-dnd";
import {
  AddSquare20Regular as AddIcon,
  ArrowSortDownLines16Regular as SortDescIcon,
  Delete16Regular as DismissIcon,
  ReOrder16Regular as ReOrderIcon,
} from "@fluentui/react-icons";
import DropDown from "./SortDropDown";
import { ColumnMetaData } from "features/table/types";
import Button, { buttonColor, buttonSize } from "components/buttons/Button";
import styles from "./sort.module.scss";
import classNames from "classnames";
import { ActionMeta } from "react-select";
import { useStrictDroppable } from "utils/UseStrictDroppable";
import View from "../../../viewManagement/View";

export type OrderBy = {
  key: string;
  direction: "asc" | "desc";
};

type OrderByOption = {
  value: string;
  label: string;
};

type SortRowProps<T> = {
  selected: OrderBy;
  options: Array<OrderByOption>;
  index: number;
  view: View<T>;
};

function SortRow<T>(props: SortRowProps<T>) {
  const handleUpdate = (newValue: unknown, _: ActionMeta<unknown>) => {
    props.view.updateSortBy(
      { key: (newValue as OrderByOption).value, direction: props.selected.direction },
      props.index
    );
  };

  const updateDirection = (selected: OrderBy) => {
    const direction = props.selected.direction === "asc" ? "desc" : "asc";
    props.view.updateSortBy({ ...props.selected, direction: direction }, props.index);
  };

  return (
    <Draggable draggableId={`draggable-sort-id-${props.index}`} index={props.index}>
      {(provided, snapshot) => (
        <div
          className={classNames(styles.sortRow, {
            dragging: snapshot.isDragging,
          })}
          ref={provided.innerRef}
          {...provided.draggableProps}
        >
          <div {...provided.dragHandleProps} className={styles.dragHandle}>
            <ReOrderIcon />
          </div>
          <div className={styles.labelWrapper}>
            <DropDown
              onChange={handleUpdate}
              options={props.options}
              getOptionLabel={(option: any): string => option["label"]}
              getOptionValue={(option: any): string => option["value"]}
              value={props.options.find((option) => option.value === props.selected.key)}
            />
          </div>

          <div
            className={classNames(styles.directionWrapper, { [styles.asc]: props.selected.direction === "asc" })}
            onClick={() => updateDirection(props.selected)}
          >
            {<SortDescIcon />}
          </div>
          <div className={styles.dismissIconWrapper}>
            <DismissIcon
              onClick={() => {
                props.view.removeSortBy(props.index);
              }}
            />
          </div>
        </div>
      )}
    </Draggable>
  );
}

type SortSectionProps<T> = {
  columns: Array<ColumnMetaData<T>>;
  view: View<T>;
  defaultSortBy: OrderBy;
};

function SortSection<T>(props: SortSectionProps<T>) {
  const [options, _] = useMemo(() => {
    const options: Array<OrderByOption> = [];
    const selected: Array<OrderByOption> = [];

    props.columns.forEach((column) => {
      options.push({
        value: column.key.toString(),
        label: column.heading,
      });
    });
    return [options, selected];
  }, [props.columns]);

  // const handleAddSort = () => {
  //   const updated: Array<OrderBy> = [
  //     ...props.orderBy,
  //     {
  //       direction: "desc",
  //       key: props.columns[0].key.toString(),
  //     },
  //   ];
  //   props.updateHandler(updated);
  // };

  // const handleUpdateSort = (update: OrderBy, index: number) => {
  //   const updated: Array<OrderBy> = [
  //     ...props.view.sortBy.slice(0, index),
  //     update,
  //     ...props.view.sortBy.slice(index + 1, props.orderBy.length),
  //   ];
  //   props.view.updateAllSortBy(updated);
  // };
  //
  // const handleRemoveSort = (index: number) => {
  //   const updated: OrderBy[] = [
  //     ...props.orderBy.slice(0, index),
  //     ...props.orderBy.slice(index + 1, props.orderBy.length),
  //   ];
  //
  //   props.updateHandler(updated);
  // };

  const droppableContainerId = "sort-list-droppable";

  const onDragEnd = (result: any) => {
    const { destination, source } = result;

    // Dropped outside of container
    if (!destination || destination.droppableId !== droppableContainerId) {
      return;
    }

    // Position not changed
    if (destination.droppableId === source.droppableId && destination.index === source.index) {
      return;
    }

    props.view.moveSortBy(source.index, destination.index);
  };

  // Workaround for incorrect handling of React.StrictMode by react-beautiful-dnd
  // https://github.com/atlassian/react-beautiful-dnd/issues/2396#issuecomment-1248018320
  const [strictDropEnabled] = useStrictDroppable(!props.view.sortBy);

  return (
    <DragDropContext onDragEnd={onDragEnd}>
      <div className={styles.sortPopoverContent}>
        {!props.view.sortBy?.length && <div className={styles.noFilters}>Not sorted</div>}

        {strictDropEnabled && (
          <Droppable droppableId={droppableContainerId}>
            {(provided) => (
              <div className={styles.sortRows} ref={provided.innerRef} {...provided.droppableProps}>
                {props.view.sortBy?.map((item, index) => {
                  return (
                    <SortRow key={item.key + index} selected={item} options={options} index={index} view={props.view} />
                  );
                })}

                {provided.placeholder}
              </div>
            )}
          </Droppable>
        )}

        <div className={styles.buttonsRow}>
          <Button
            buttonColor={buttonColor.ghost}
            buttonSize={buttonSize.small}
            onClick={() => props.view.addSortBy(props.defaultSortBy)}
            text={"Add"}
            icon={<AddIcon />}
          />
        </div>
      </div>
    </DragDropContext>
  );
}

export default SortSection;
