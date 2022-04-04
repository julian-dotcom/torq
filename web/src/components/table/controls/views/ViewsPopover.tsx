import './views.scss'
import { DragDropContext, Droppable, Draggable } from 'react-beautiful-dnd';
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
  DefaultView, createTableViewAsync, deleteTableViewAsync, updateViewsOrder
} from "../../tableSlice";
import {useState} from "react";
import classNames from "classnames";

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
  <Draggable draggableId={`draggable-view-id-${index}`} index={index}>
      {(provided, snapshot) => (
      <div className={classNames("view-row", {"dragging": snapshot.isDragging})}
         ref={provided.innerRef}
         {...provided.draggableProps}>

        <div className="view-row-drag-handle" {...provided.dragHandleProps}>
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

      </div>)}
    </Draggable>
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

    const droppableContainerId = "views-list-droppable"

  const onDragEnd = (result: any) => {
    const {destination, source, draggableId} = result;

    // Dropped outside of container
    if (!destination || destination.droppableId !== droppableContainerId) {
      return;
    }

    // Position not changed
    if (
      destination.droppableId === source.droppableId &&
      destination.index === source.index
    ) {
      return;
    }

    let newCurrentIndex = selectedView;
    // If moved view is the selected view
    if (source.index === selectedView) {
      newCurrentIndex = destination.index
    }
    // If moved view has a source bellow the selected view
    if ((source.index < selectedView) && (destination.index >= selectedView)) {
      newCurrentIndex = selectedView -1
    }
    // If moved view has a source above the selected view
    if ((source.index > selectedView) && (destination.index <= selectedView)) {
      newCurrentIndex = selectedView +1
    }

    let newViewsOrder: ViewInterface[] = views.slice()
    newViewsOrder.splice(source.index, 1)
    newViewsOrder.splice(destination.index, 0, views[source.index])
    dispatch(updateViewsOrder({views: newViewsOrder, index: newCurrentIndex}))
  }

  let popOverButton = <DefaultButton
      text={views[selectedView].title}
      icon={<TableIcon/>}
      className={""}/>

  const singleView = views.length <= 1
  return (
    <Popover button={popOverButton}>
      <DragDropContext
        onDragEnd={onDragEnd}
      >
        <div className="views-popover-content">
          <Droppable droppableId={droppableContainerId}>
            {provided => (
              <div
                className="view-rows"
                ref={provided.innerRef}
                {...provided.droppableProps}
              >
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
                {provided.placeholder}
              </div>
            )}
          </Droppable>
          <div className="buttons-row">
            <DefaultButton text={"Add table"} icon={<AddIcon/>} onClick={addView} />
          </div>
        </div>
      </DragDropContext>
    </Popover>
  )
}

export default ViewsPopover;
