import classNames from "classnames";
import { DragDropContext, Droppable, Draggable } from "react-beautiful-dnd";
import {
  Delete16Regular as RemoveIcon,
  Options16Regular as OptionsIcon,
  LineHorizontal1Regular as OptionsExpandedIcon,
  LockClosed16Regular as LockClosedIcon,
  AddCircle16Regular as AddIcon,
  ReOrder16Regular as DragHandle,
} from "@fluentui/react-icons";
import styles from "./columns-section.module.scss";
import Select, { SelectOptionType } from "./ColumnDropDown";
import { ColumnMetaData } from "features/table/types";
import { useState } from "react";
import { useStrictDroppable } from "utils/UseStrictDroppable";
import { useAppDispatch } from "store/hooks";
import {
  addColumn,
  deleteColumn,
  updateColumn,
  updateColumnsOrder,
  ViewSliceStatePages,
} from "features/viewManagement/viewSlice";
import { TableResponses } from "features/viewManagement/types";

const CellOptions: SelectOptionType[] = [
  { label: "Alias", value: "AliasCell" },
  { label: "Number", value: "NumericCell" },
  { label: "Boolean", value: "BooleanCell" },
  { label: "Array", value: "EnumCell" },
  { label: "Bar", value: "BarCell" },
  { label: "Text", value: "TextCell" },
  { label: "Duration", value: "DurationCell" },
  { label: "Date", value: "DateCell" },
  { label: "Long Text", value: "LongTextCell" },
];

const NumericCellOptions: SelectOptionType[] = [
  { label: "Number", value: "NumericCell" },
  { label: "Bar", value: "BarCell" },
  { label: "Text", value: "TextCell" },
  { label: "Long Text", value: "LongTextCell" },
];

const StringCellOptions: SelectOptionType[] = [
  { label: "Text", value: "TextCell" },
  { label: "Long Text", value: "LongTextCell" },
];

const Options = new Map<string, Array<{ label: string; value: string }>>([
  ["number", NumericCellOptions],
  ["string", StringCellOptions],
  ["boolean", [{ label: "Boolean", value: "BooleanCell" }]],
  ["enum", [{ label: "Array", value: "EnumCell" }]],
  ["array", [{ label: "Array", value: "EnumCell" }]],
  ["date", [{ label: "Date", value: "DateCell" }]],
  ["duration", [{ label: "Duration", value: "TextCell" }]],
]);

type ColumnRow<T> = {
  page: ViewSliceStatePages;
  viewIndex: number;
  column: ColumnMetaData<T>;
  index: number;
};

// TODO: Fix bug that causes incorrect rendering of locked cells when they are removed and then added back
function LockedColumnRow<T>(props: ColumnRow<T>) {
  const dispatch = useAppDispatch();

  const selectedOption = CellOptions.filter((option) => {
    if (option.value === props.column.type) {
      return true;
    }
  })[0];

  function handleCellTypeChange(newValue: unknown) {
    dispatch(
      updateColumn({
        page: props.page,
        viewIndex: props.viewIndex,
        columnIndex: props.index,
        columnUpdate: {
          type: (newValue as { value: string; label: string }).value,
        },
      })
    );
  }

  const [expanded, setExpanded] = useState(false);
  return (
    <div
      className={classNames(styles.rowContent, {
        [styles.expanded]: expanded,
      })}
    >
      <div className={classNames(styles.columnRow)}>
        <div className={classNames(styles.rowLeftIcon, styles.lockBtn)}>
          <LockClosedIcon />
        </div>

        <div className={styles.columnName}>
          <div>{props.column.heading}</div>
        </div>

        <div className={styles.rowOptions} onClick={() => setExpanded(!expanded)}>
          {expanded ? <OptionsExpandedIcon /> : <OptionsIcon />}
        </div>
      </div>
      <div className={styles.rowOptionsContainer}>
        <Select
          isDisabled={["date", "duration", "array", "boolean", "enum"].includes(props.column.valueType)}
          options={Options.get(props.column.valueType)}
          value={selectedOption}
          onChange={handleCellTypeChange}
        />
      </div>
    </div>
  );
}

function ColumnRow<T>(props: ColumnRow<T>) {
  const dispatch = useAppDispatch();

  const selectedOption: { value: string; label: string } = CellOptions.filter((option) => {
    if (option.value === props.column.type) {
      return true;
    }
  })[0];

  const [expanded, setExpanded] = useState(false);
  return (
    <Draggable draggableId={`draggable-column-id-${props.column.key.toString()}`} index={props.index}>
      {(provided, snapshot) => (
        <div
          className={classNames(styles.rowContent, {
            dragging: snapshot.isDragging,
            [styles.expanded]: expanded,
          })}
          ref={provided.innerRef}
          {...provided.draggableProps}
        >
          <div className={classNames(styles.columnRow)}>
            <div className={classNames(styles.rowLeftIcon, styles.dragHandle)} {...provided.dragHandleProps}>
              <DragHandle />
            </div>

            <div className={styles.columnName}>
              <div>{props.column.heading}</div>
            </div>

            <div className={styles.rowOptions} onClick={() => setExpanded(!expanded)}>
              {expanded ? <OptionsExpandedIcon /> : <OptionsIcon />}
            </div>

            <div
              className={styles.removeColumn}
              onClick={() => {
                dispatch(deleteColumn({ page: props.page, viewIndex: props.viewIndex, columnIndex: props.index }));
              }}
            >
              <RemoveIcon />
            </div>
          </div>
          <div className={styles.rowOptionsContainer}>
            <Select
              isDisabled={["date", "duration", "array", "boolean", "enum"].includes(props.column.valueType)}
              options={Options.get(props.column.valueType)}
              value={selectedOption}
              onChange={(newValue) => {
                dispatch(
                  updateColumn({
                    page: props.page,
                    viewIndex: props.viewIndex,
                    columnIndex: props.index,
                    columnUpdate: {
                      type: (newValue as { value: string; label: string }).value,
                    },
                  })
                );
              }}
            />
          </div>
        </div>
      )}
    </Draggable>
  );
}

interface unselectedColumnRow {
  name: string;
  onAddColumn: () => void;
}

function UnselectedColumn({ name, onAddColumn }: unselectedColumnRow) {
  return (
    <div
      className={styles.unselectedColumnRow}
      onClick={() => {
        onAddColumn();
      }}
    >
      <div className={styles.unselectedColumnName}>
        <div>{name}</div>
      </div>

      <div className={styles.addColumn}>
        <AddIcon />
      </div>
    </div>
  );
}

type ColumnsSectionProps<T> = {
  page: ViewSliceStatePages;
  viewIndex: number;
  activeColumns: Array<ColumnMetaData<T>>;
  allColumns: Array<ColumnMetaData<T>>;
};

function ColumnsSection<T>(props: ColumnsSectionProps<T>) {
  const dispatch = useAppDispatch();
  const droppableContainerId = "column-list-droppable";

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onDragEnd = (result: any, _: unknown) => {
    const { destination, source } = result;

    // Dropped outside of container
    if (!destination || destination.droppableId !== droppableContainerId) {
      return;
    }

    // Position not changed
    if (destination.droppableId === source.droppableId && destination.index === source.index) {
      return;
    }

    dispatch(
      updateColumnsOrder({
        page: props.page,
        viewIndex: props.viewIndex,
        fromIndex: source.index,
        toIndex: destination.index,
      })
    );
  };

  // Workaround for incorrect handling of React.StrictMode by react-beautiful-dnd
  // https://github.com/atlassian/react-beautiful-dnd/issues/2396#issuecomment-1248018320
  const [strictDropEnabled] = useStrictDroppable(!props.activeColumns);

  return (
    <div>
      <DragDropContext onDragEnd={onDragEnd}>
        <div className={styles.columnsSectionContent}>
          <div className={styles.columnRows}>
            {props.allColumns.map((column, index) => {
              if (column.locked === true) {
                return (
                  <LockedColumnRow
                    viewIndex={props.viewIndex}
                    page={props.page}
                    column={column}
                    key={"selected-" + column.key.toString() + "-" + index}
                    index={index}
                  />
                );
              }
            })}
            {strictDropEnabled && (
              <Droppable droppableId={droppableContainerId}>
                {(provided) => (
                  <div
                    className={classNames(styles.draggableColumns, styles.columnRows)}
                    ref={provided.innerRef}
                    {...provided.droppableProps}
                  >
                    {props.activeColumns.map((column, index) => {
                      if (column.locked !== true) {
                        return (
                          <ColumnRow
                            viewIndex={props.viewIndex}
                            page={props.page}
                            column={column}
                            key={"selected-" + column.key.toString() + "-" + index}
                            index={index}
                          />
                        );
                      }
                    })}
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            )}

            <div className={styles.divider} />

            <div className={styles.unselectedColumnsWrapper}>
              {props.allColumns.map((column, index) => {
                if (props.activeColumns.findIndex((ac) => ac.key === column.key) === -1) {
                  return (
                    <UnselectedColumn
                      name={column.heading}
                      key={"unselected-" + index + "-" + column.key.toString()}
                      onAddColumn={() => {
                        dispatch(
                          addColumn({
                            page: props.page,
                            viewIndex: props.viewIndex,
                            newColumn: column as ColumnMetaData<TableResponses>,
                          })
                        );
                      }}
                    />
                  );
                }
              })}
            </div>
          </div>
        </div>
      </DragDropContext>
    </div>
  );
}

export default ColumnsSection;
