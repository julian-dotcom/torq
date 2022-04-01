import './columns_popover.scss'
import {useState} from "react";
import {
  ColumnTriple20Regular as ColumnsIcon,
  Dismiss20Regular as RemoveIcon,
  Search20Regular as SearchIcon,
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

interface columnRow {
  name: string;
  index: number;
  cellType: string;
  selected: boolean;
  handleRemoveColumn: Function;
  handleUpdateColumn: Function;
}

const CellOptions: SelectOptionType[] = [
  {label: "Number", value: "NumericCell"},
  {label: "Bar", value: "BarCell"},
]

function ColumnRow({name, index, cellType, selected, handleRemoveColumn, handleUpdateColumn}: columnRow) {
  const selectedOption = CellOptions.filter((column) => {
    if (column.value === cellType) {
      return column
    }
  })[0] || {label: 'Name', value: 'AliasCell'}

  return (
    <div className="column-row">
      <div className="row-drag-handle">
       {name !== "Name" ?(<DragHandle/>):""}
      </div>

      <div className="column-name" onClick={() => (index)}>
        <div>{name}</div>
      </div>

      <div className="column-type-select" onClick={() => (index)}>
        {cellType !== 'AliasCell' ? (
          <TorqSelect options={CellOptions} value={selectedOption} onChange={() => {
          }} isDisabled={!selectedOption}/>
        ) : (
          <TorqSelect options={selectedOption} value={selectedOption} isDisabled={true}/>
        )}
      </div>

      {name !== "Name" ? (
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

    const activeColumns = useAppSelector(selectActiveColumns)
    const columns = useAppSelector(selectAllColumns)
    const dispatch = useAppDispatch();

    // const updateColumn = (view: ViewInterface, index: number) => {
    //    const updatedViews = [
    //      ...views.slice(0,index),
    //      {...views[index], ...view},
    //       ...views.slice(index+1, views.length)
    //    ]
    //   dispatch(
    //     updateViews( {views: updatedViews})
    //   )
    // }

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
              name={column.heading}
              key={"selected-"+index}
              index={index}
              handleRemoveColumn={removeColumn}
              handleUpdateColumn={() => {}}
              cellType={column.type as string}
              selected={true}/>
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
              return (column.key !== "alias" ? (<UnselectedColumn
                name={column.heading}
                key={"unselected-" + index}
                index={index}
                handleAddColumn={addColumn}/>) : "")
            })}
          </div>

        </div>

      </div>
    </Popover>
  )
}

export default ColumnsPopover;
