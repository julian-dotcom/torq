import './views.scss'
import Popover from "../../../popover/Popover";
import DefaultButton from "../../../buttons/Button";
import {
  Table20Regular as TableIcon,
  Dismiss20Regular as RemoveIcon,
  Edit16Regular as EditIcon,
  Save20Regular as SaveIcon,
  AddSquare20Regular as AddIcon,
  Reorder20Regular as DragHandle,
} from "@fluentui/react-icons";
import {useAppDispatch, useAppSelector} from "../../../../store/hooks";
import {
  selectViews,
  updateViews,
  deleteView,
  updateSelectedView,
  selectedViewIndex,
  ViewInterface,
  DefaultView, createTableViewAsync, deleteTableViewAsync
} from "../../tableSlice";
import {useState} from "react";

interface viewRow {
  title: string;
  index: number;
  handleUpdateView: Function;
  handleRemoveView: Function;
  handleSelectView: Function;
  singleView: boolean;
}

function ViewRow({title, index, handleUpdateView, handleRemoveView, handleSelectView, singleView}: viewRow) {

  const [editView, setEditView] = useState(false);
  const [localTitle, setLocalTitle] = useState(title);

  function handleInputChange(e: any) {
    setLocalTitle(e.target.value)
  }

  function handleInputSubmit(e: any) {
    e.preventDefault()
    setEditView(false)
    handleUpdateView({title: localTitle}, index)
  }

  return (
    <div className="view-row">
      <div className="view-row-drag-handle">
        <DragHandle/>
      </div>

      {editView ? (
        <form onSubmit={handleInputSubmit} className={"view-edit torq-input-icon-field"}>
          <input type="text"
            className={""}
            autoFocus={true}
            onChange={handleInputChange}
            value={localTitle}/>
          <button type={"submit"}><SaveIcon/></button>
        </form>
        ):
        <div className="view-select" onClick={() => handleSelectView(index)}>
          <div>{title}</div>
          <div className="edit-view" onClick={() => (setEditView(true))}>
            <EditIcon/>
          </div>
        </div>
      }
      {!singleView ? (<div className="remove-view" onClick={() => (handleRemoveView(index))}>
        <RemoveIcon/>
      </div>) : (
        <div className="remove-view disabled" >
          <RemoveIcon/>
        </div>)}

    </div>
  )
}

function ViewsPopover() {

    const views = useAppSelector(selectViews)
    const selectedView = useAppSelector(selectedViewIndex)
    const dispatch = useAppDispatch();

    const updateView = (view: ViewInterface, index: number) => {
       const updatedViews = [
         ...views.slice(0,index),
         {...views[index], ...view},
          ...views.slice(index+1, views.length)
       ]
      dispatch(updateViews( {views: updatedViews, index: index}))
    }

    const removeView = (index: number) => {
      let confirmed = window.confirm("Are you sure you want to delete this view?")
      if (!confirmed) {
        return
      }
      dispatch(deleteTableViewAsync({view: views[index], index: index}))
    }

    const addView = () => {
      dispatch(createTableViewAsync({view: DefaultView, index: views.length}))
    }

    const selectView = (index: number) => {
      dispatch(
        updateSelectedView({index: index})
      )
    }

  let popOverButton = <DefaultButton
      text={views[selectedView].title}
      icon={<TableIcon/>}
      className={"collapse-tablet"}/>

  const singleView = views.length <= 1 ? true :false
  return (
    <Popover button={popOverButton}>
      <div className="views-popover-content">
        <div className="view-rows">
          {views.map((view, index) => {
            return <ViewRow
              title={view.title}
              key={index}
              index={index}
              handleRemoveView={removeView}
              handleUpdateView={updateView}
              handleSelectView={selectView}
              singleView={singleView} />
          })}
        </div>
        <div className="buttons-row">
          <DefaultButton text={"Add table"} icon={<AddIcon/>} onClick={addView} />
        </div>
      </div>
    </Popover>
  )
}

export default ViewsPopover;
