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
import View from "features/viewManagement/View";

const CellOptions: SelectOptionType[] = [
  { label: "Number", value: "NumericCell" },
  { label: "Boolean", value: "BooleanCell" },
  { label: "Array", value: "EnumCell" },
  { label: "Bar", value: "BarCell" },
  { label: "Text", value: "TextCell" },
  { label: "Duration", value: "TextCell" },
];

const NumericCellOptions: SelectOptionType[] = [
  { label: "Number", value: "NumericCell" },
  { label: "Bar", value: "BarCell" },
  { label: "Text", value: "TextCell" },
];

type ColumnRow<T> = {
  view: View<T>;
  index: number;
};

function ColumnRow<T>(props: ColumnRow<T>) {
  const column = props.view.getColumn(props.index);
  const selectedOption = CellOptions.filter((option) => {
    if (option.value === column.type) {
      return option;
    }
  })[0];

  const [expanded, setExpanded] = useState(false);

  return (
    <Draggable draggableId={`draggable-column-id-${column.key.toString()}`} index={props.index}>
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
            {column.locked ? (
              <div className={classNames(styles.rowLeftIcon, styles.lockBtn)}>
                <LockClosedIcon />
              </div>
            ) : (
              <div className={classNames(styles.rowLeftIcon, styles.dragHandle)} {...provided.dragHandleProps}>
                <DragHandle />
              </div>
            )}

            <div className={styles.columnName}>
              <div>{column.heading}</div>
            </div>

            <div className={styles.rowOptions} onClick={() => setExpanded(!expanded)}>
              {expanded ? <OptionsExpandedIcon /> : <OptionsIcon />}
            </div>

            <div
              className={styles.removeColumn}
              onClick={() => {
                props.view.removeColumn(props.index);
              }}
            >
              <RemoveIcon />
            </div>
          </div>
          <div className={styles.rowOptionsContainer}>
            <Select
              isDisabled={["date", "array", "string", "boolean", "enum"].includes(column.valueType)}
              options={NumericCellOptions}
              value={selectedOption}
              onChange={(newValue) => {
                props.view.updateColumn(
                  {
                    ...column,
                    type: (newValue as { value: string; label: string }).value,
                  },
                  props.index
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
  view: View<T>;
  columns: Array<ColumnMetaData<T>>;
};

function ColumnsSection<T>(props: ColumnsSectionProps<T>) {
  const droppableContainerId = "column-list-droppable";

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

    props.view.moveColumn(source.index, destination.index);
  };

  // Workaround for incorrect handling of React.StrictMode by react-beautiful-dnd
  // https://github.com/atlassian/react-beautiful-dnd/issues/2396#issuecomment-1248018320
  const [strictDropEnabled] = useStrictDroppable(!props.view.columns);

  return (
    <div>
      <DragDropContext onDragEnd={onDragEnd}>
        <div className={styles.columnsSectionContent}>
          <div className={styles.columnRows}>
            {props.view.getLocedColumns().map((column, index) => {
              return <ColumnRow view={props.view} key={"selected-" + column.key.toString() + "-" + index} index={0} />;
            })}
            {strictDropEnabled && (
              <Droppable droppableId={droppableContainerId}>
                {(provided) => (
                  <div
                    className={classNames(styles.draggableColumns, styles.columnRows)}
                    ref={provided.innerRef}
                    {...provided.droppableProps}
                  >
                    {props.view.getMovableColumns().map((column, index) => {
                      if (column) {
                        return (
                          <ColumnRow
                            view={props.view}
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
              {props.columns.map((column, index) => {
                if (props.view.columns.findIndex((ac) => ac.key === column.key) === -1) {
                  return (
                    <UnselectedColumn
                      name={column.heading}
                      key={"unselected-" + index + "-" + column.key.toString()}
                      onAddColumn={() => {
                        props.view.addColumn(column);
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
