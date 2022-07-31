import classNames from "classnames";
import { DragDropContext, Droppable, Draggable } from "react-beautiful-dnd";
import {
  Dismiss20Regular as RemoveIcon,
  LockClosed20Regular as LockClosedIcon,
  LockOpen20Regular as LockOpenIcon,
  Add20Regular as AddIcon,
  Reorder20Regular as DragHandle,
} from "@fluentui/react-icons";
import styles from "./columns-section.module.scss";
import Select, { SelectOptionType } from "features/inputs/Select";
import { ColumnMetaData } from "features/forwards/forwardsSlice";

interface columnRow {
  column: ColumnMetaData;
  index: number;
  handleRemoveColumn: Function;
  handleUpdateColumn: Function;
}

const CellOptions: SelectOptionType[] = [
  { label: "Number", value: "NumericCell" },
  { label: "Bar", value: "BarCell" },
  { label: "Text", value: "TextCell" },
];

function NameColumnRow({ column, index }: { column: ColumnMetaData; index: number }) {
  return (
    <div className={classNames(styles.columnRow, index)}>
      <div className={classNames(styles.rowLeftIcon, styles.lockBtn)}>
        {column.locked ? <LockClosedIcon /> : <LockOpenIcon />}
      </div>

      <div className={styles.columnName}>
        <div>{column.heading}</div>
      </div>
      <Select value={{ label: "Name", value: "AliasCell" }} isDisabled={true} />
    </div>
  );
}

function ColumnRow({ column, index, handleRemoveColumn, handleUpdateColumn }: columnRow) {
  const selectedOption = CellOptions.filter((option) => {
    if (option.value === column.type) {
      return option;
    }
  })[0] || { label: "Name", value: "AliasCell" };

  return (
    <Draggable draggableId={`draggable-column-id-${index}`} index={index}>
      {(provided, snapshot) => (
        <div
          className={classNames(styles.columnRow, {
            dragging: snapshot.isDragging,
          })}
          ref={provided.innerRef}
          {...provided.draggableProps}
        >
          <div className={classNames(styles.rowLeftIcon, styles.dragHandle)} {...provided.dragHandleProps}>
            <DragHandle />
          </div>

          <div className={styles.columnName}>
            <div>{column.heading}</div>
          </div>

          <div className={styles.columnTypeSelect}>
            <Select
              isDisabled={column.valueType === "string"}
              options={CellOptions}
              value={selectedOption}
              onChange={(o, actionMeta) => {
                handleUpdateColumn(
                  {
                    ...column,
                    type: actionMeta.option,
                  },
                  index + 1
                );
              }}
            />
          </div>

          <div
            className={styles.removeColumn}
            onClick={() => {
              handleRemoveColumn(index + 1);
            }}
          >
            <RemoveIcon />
          </div>
        </div>
      )}
    </Draggable>
  );
}

interface unselectedColumnRow {
  name: string;
  index: number;
  handleAddColumn: Function;
}

function UnselectedColumn({ name, index, handleAddColumn }: unselectedColumnRow) {
  return (
    <div
      className={styles.unselectedColumnRow}
      onClick={() => {
        handleAddColumn(index);
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

type ColumnsSectionProps = {
  activeColumns: ColumnMetaData[];
  columns: ColumnMetaData[];
  handleUpdateColumn: Function;
};

function ColumnsSection(props: ColumnsSectionProps) {
  const updateColumn = (column: ColumnMetaData, index: number) => {
    const updatedColumns = [
      ...props.activeColumns.slice(0, index),
      column,
      ...props.activeColumns.slice(index + 1, props.columns.length),
    ];
    props.handleUpdateColumn(updatedColumns);
  };

  const removeColumn = (index: number) => {
    const updatedColumns: ColumnMetaData[] = [
      ...props.activeColumns.slice(0, index),
      ...props.activeColumns.slice(index + 1, props.activeColumns.length),
    ];
    props.handleUpdateColumn(updatedColumns);
  };

  const addColumn = (index: number) => {
    const updatedColumns: ColumnMetaData[] = [...props.activeColumns, props.columns[index]];
    props.handleUpdateColumn(updatedColumns);
  };

  const droppableContainerId = "column-list-droppable";
  const draggableColumns = props.activeColumns.slice(1, props.activeColumns.length);

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

    let newColumns: ColumnMetaData[] = draggableColumns.slice();
    newColumns.splice(source.index, 1);
    newColumns.splice(destination.index, 0, draggableColumns[source.index]);
    props.handleUpdateColumn([props.columns[0], ...newColumns]);
  };

  return (
    <div>
      <DragDropContext onDragEnd={onDragEnd}>
        <div className={styles.columnsPopoverContent}>
          <div className={styles.columnRows}>
            <NameColumnRow column={props.activeColumns.slice(0, 1)[0]} key={"selected-0-name"} index={0} />

            <div className={styles.divider} />

            <Droppable droppableId={droppableContainerId}>
              {(provided) => (
                <div
                  className={classNames(styles.draggableColumns, styles.columnRows)}
                  ref={provided.innerRef}
                  {...provided.droppableProps}
                >
                  {draggableColumns.map((column, index) => {
                    return (
                      <ColumnRow
                        column={column}
                        key={"selected-" + index}
                        index={index}
                        handleRemoveColumn={removeColumn}
                        handleUpdateColumn={updateColumn}
                      />
                    );
                  })}
                  {provided.placeholder}
                </div>
              )}
            </Droppable>

            <div className={styles.divider} />

            <div className={styles.unselectedColumnsWrapper}>
              <div className={styles.unselectedColumns}>
                {props.columns.map((column, index) => {
                  if (props.activeColumns.findIndex((ac) => ac.key === column.key) === -1) {
                    return (
                      <UnselectedColumn
                        name={column.heading}
                        key={"unselected-" + index}
                        index={index}
                        handleAddColumn={addColumn}
                      />
                    );
                  }
                })}
                {props.activeColumns.length === props.columns.length ? (
                  <div className={styles.noFilters}>All columns added</div>
                ) : (
                  ""
                )}
              </div>
            </div>
          </div>
        </div>
      </DragDropContext>
    </div>
  );
}

export default ColumnsSection;
