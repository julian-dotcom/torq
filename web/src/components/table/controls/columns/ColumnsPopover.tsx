import './columns_popover.scss'
import {useState} from "react";
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
import TorqSelect, {SelectOptionType} from "../../../inputs/Select";
import {
  updateColumns,
  ColumnMetaData,
  selectAllColumns,
  selectActiveColumns
} from "../../tableSlice";
import {useAppDispatch, useAppSelector} from "../../../../store/hooks";
import Fuse from "fuse.js";
import classNames from "classnames";

interface columnRow {
  column: ColumnMetaData;
  index: number;
  handleRemoveColumn: Function;
  handleUpdateColumn: Function;
}

const CellOptions: SelectOptionType[] = [
  {label: "Number", value: "NumericCell"},
  {label: "Bar", value: "BarCell"},
]

function ColumnRow({column, index, handleRemoveColumn, handleUpdateColumn}: columnRow) {

  const selectedOption = CellOptions.filter((option) => {
    if (option.value === column.type) {
      return option
    }
  })[0] || {label: 'Name', value: 'AliasCell'}

  return (
    <div className="column-row">
      {column.key === "alias" ? (
        <div className={classNames("row-left-icon lock-btn")}>
          {column.locked ? <LockClosedIcon/> : <LockOpenIcon/>}
        </div>
      ) : (
        <div className={classNames("row-left-icon  drag-handle")}>
          <DragHandle/>
        </div>
      )}

      <div className="column-name">
        <div>{column.heading}</div>
      </div>

      <div className="column-type-select">
        {column.type !== 'AliasCell' ? (
          <TorqSelect options={CellOptions} value={selectedOption} onChange={(o: SelectOptionType) => {
            handleUpdateColumn({
              ...column,
              type: o.value,
            }, index)
          }} isDisabled={!selectedOption}/>
        ) : (
          <TorqSelect options={selectedOption} value={selectedOption} isDisabled={true}/>
        )}
      </div>

      {column.key !== "alias" ? (
        <div className="remove-column" onClick={() => {handleRemoveColumn(index)}}>
          <RemoveIcon/>
        </div>
      ) : ""}

    </div>
  )
}

interface unselectedColumnRow {
  name: string;
  index: number;
  handleAddColumn: Function;
}

function UnselectedColumn({name, index, handleAddColumn}: unselectedColumnRow) {

  return (
    <div className="unselected-column-row" onClick={() => {handleAddColumn(index)}}>

      <div className="unselected-column-name">
        <div>{name}</div>
      </div>

      <div className="add-column" >
        <AddIcon/>
      </div>

    </div>
  )
}

function ColumnsPopover() {

    const activeColumns = useAppSelector(selectActiveColumns);
    const columns = useAppSelector(selectAllColumns);
    const dispatch = useAppDispatch();

    const updateColumn = (column: ColumnMetaData, index: number) => {
       const updatedColumns = [
         ...activeColumns.slice(0,index),
         column,
          ...activeColumns.slice(index+1, columns.length)
       ]
      dispatch(
        updateColumns( {columns: updatedColumns})
      )
    }

    const removeColumn = (index: number) => {
      const updatedColumns: ColumnMetaData[] = [
        ...activeColumns.slice(0,index),
        ...activeColumns.slice(index+1, activeColumns.length)
      ]
      dispatch(updateColumns({columns: updatedColumns}))
    }

    const addColumn = (index: number) => {
      const updatedColumns: ColumnMetaData[] = [
        ...activeColumns,
        columns[index],
      ]
      dispatch(updateColumns({columns: updatedColumns}))
    }

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

  let popOverButton = <DefaultButton
      text={activeColumns.length + " Columns"}
      icon={<ColumnsIcon/>}
      isOpen={true}
      className={'collapse-tablet'}/>

  return (
    <Popover button={popOverButton} className={"scrollable"}>
      <div className="columns-popover-content">

        <div className="column-rows">
          {activeColumns.map((column, index) => {
            return <ColumnRow
              column={column}
              key={"selected-"+index}
              index={index}
              handleRemoveColumn={removeColumn}
              handleUpdateColumn={updateColumn}/>
          })}
        </div>

        <div className="divider"/>

        <div className={"unselected-columns-wrapper"}>

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

          <div className="unselected-columns">
            {columns.map((column, index) => {
              if (activeColumns.findIndex((ac) => ac.key === column.key) === -1) {
                return (<UnselectedColumn
                  name={column.heading}
                  key={"unselected-" + index}
                  index={index}
                  handleAddColumn={addColumn}/>)
              }
            })}
            {activeColumns.length === columns.length ? (<div className={"no-filters"}>All columns added</div>) : ""}
          </div>

        </div>

      </div>
    </Popover>
  )
}

export default ColumnsPopover;
