import { useState } from "react";
import classNames from "classnames";
import { Draggable } from "react-beautiful-dnd";
import {
  Dismiss20Regular as RemoveIcon,
  Edit16Regular as EditIcon,
  Save20Regular as SaveIcon,
  Reorder20Regular as DragHandle,
} from "@fluentui/react-icons";
import { ViewInterface, ViewRow } from "./types";
import styles from "./views.module.scss";
import { useCreateTableViewMutation, useDeleteTableViewMutation, useUpdateTableViewMutation } from "./viewsApiSlice";
import { ViewInterfaceResponse } from "./types";

export default function ViewRow(props: ViewRow<ViewInterfaceResponse>) {
  const [updateTableView] = useUpdateTableViewMutation();
  const [deleteTableView] = useDeleteTableViewMutation();

  const [editView, setEditView] = useState(false);
  const [localTitle, setLocalTitle] = useState(props.view.title);

  function handleInputChange(e: any) {
    setLocalTitle(e.target.value);
  }

  return (
    <Draggable draggableId={`draggable-view-id-${props.index}`} index={props.index}>
      {(provided, snapshot) => (
        <div
          className={classNames(styles.viewRow, { dragging: snapshot.isDragging })}
          ref={provided.innerRef}
          {...provided.draggableProps}
        >
          <div className={styles.viewRowDragHandle} {...provided.dragHandleProps}>
            <DragHandle />
          </div>

          {editView ? (
            <form
              onSubmit={() => updateTableView({ ...props.view, title: localTitle })}
              className={classNames(styles.viewEdit, "torq-input-field")}
            >
              <input type="text" autoFocus={true} onChange={handleInputChange} value={localTitle} />
              <button type={"submit"}>
                <SaveIcon />
              </button>
            </form>
          ) : (
            <div className={styles.viewSelect} onClick={() => props.onSelectView(props.index)}>
              <div>{props.view.title}</div>
              <div className={styles.editView} onClick={() => setEditView(true)}>
                <EditIcon />
              </div>
            </div>
          )}
          {props.singleView ? (
            <div className={classNames(styles.removeView, styles.disabled)}>
              <RemoveIcon />
            </div>
          ) : (
            <div className={styles.removeView} onClick={() => deleteTableView({ id: props.index })}>
              <RemoveIcon />
            </div>
          )}
        </div>
      )}
    </Draggable>
  );
}
