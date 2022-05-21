import styles from "./columns_popover.module.scss";
import { DragDropContext, Droppable, Draggable } from "react-beautiful-dnd";
import {
  ColumnTriple20Regular as ColumnsIcon,
  Dismiss20Regular as RemoveIcon,
  LockClosed20Regular as LockClosedIcon,
  LockOpen20Regular as LockOpenIcon,
  Add20Regular as AddIcon,
  Reorder20Regular as DragHandle,
} from "@fluentui/react-icons";
import DefaultButton from "../../../buttons/Button";
import Popover from "../../../popover/Popover";
import Select, { SelectOptionType } from "../../../inputs/Select";
import { updateColumns, ColumnMetaData, selectAllColumns, selectActiveColumns } from "../../tableSlice";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
// import Fuse from "fuse.js";
import classNames from "classnames";

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

function ColumnsPopover() {
  const activeColumns = useAppSelector(selectActiveColumns);
  const columns = useAppSelector(selectAllColumns);
  const dispatch = useAppDispatch();

  const updateColumn = (column: ColumnMetaData, index: number) => {
    const updatedColumns = [
      ...activeColumns.slice(0, index),
      column,
      ...activeColumns.slice(index + 1, columns.length),
    ];
    dispatch(updateColumns({ columns: updatedColumns }));
  };

  const removeColumn = (index: number) => {
    const updatedColumns: ColumnMetaData[] = [
      ...activeColumns.slice(0, index),
      ...activeColumns.slice(index + 1, activeColumns.length),
    ];
    dispatch(updateColumns({ columns: updatedColumns }));
  };

  const addColumn = (index: number) => {
    const updatedColumns: ColumnMetaData[] = [...activeColumns, columns[index]];
    dispatch(updateColumns({ columns: updatedColumns }));
  };

  // const fuzeOptions = {
  //   keys: ['heading']
  // }
  // const fuse = new Fuse(columns, fuzeOptions)
  //
  // const handleSearch = (e: any) => {
  //   if (e.target.value) {
  //     return fuse.search(e.target.value).map((item) => item.item)
  //   }
  // }

  const droppableContainerId = "column-list-droppable";
  const draggableColumns = activeColumns.slice(1, activeColumns.length);

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
    dispatch(updateColumns({ columns: [columns[0], ...newColumns] }));
  };

  let popOverButton = (
    <DefaultButton
      text={activeColumns.length + " Columns"}
      icon={<ColumnsIcon />}
      isOpen={true}
      className={"collapse-tablet"}
    />
  );

  return (
    <Popover button={popOverButton} className={"scrollable"}>
      <DragDropContext onDragEnd={onDragEnd}>
        <div className={styles.columnsPopoverContent}>
          <div className={styles.columnRows}>
            <NameColumnRow column={activeColumns.slice(0, 1)[0]} key={"selected-0-name"} index={0} />

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
              {/*<div className="column-search-wrapper">*/}
              {/*  <div className="search-icon">*/}
              {/*    <SearchIcon/>*/}
              {/*  </div>*/}
              {/*  <input type="search"*/}
              {/*    className={"torq-input-field column-search-field"}*/}
              {/*    autoFocus={true}*/}
              {/*    onChange={handleSearch}*/}
              {/*    placeholder={"Search..."}*/}
              {/*    />*/}
              {/*</div>*/}

              <div className={styles.unselectedColumns}>
                {columns.map((column, index) => {
                  if (activeColumns.findIndex((ac) => ac.key === column.key) === -1) {
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
                {activeColumns.length === columns.length ? (
                  <div className={styles.noFilters}>All columns added</div>
                ) : (
                  ""
                )}
              </div>
            </div>
          </div>
        </div>
      </DragDropContext>
    </Popover>
  );
}

export default ColumnsPopover;
