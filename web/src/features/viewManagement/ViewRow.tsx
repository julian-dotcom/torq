import { useState } from "react";
import classNames from "classnames";
import { Draggable } from "react-beautiful-dnd";
import {
  Dismiss20Regular as RemoveIcon,
  Edit16Regular as EditIcon,
  Save20Regular as SaveIcon,
  Reorder20Regular as DragHandle,
} from "@fluentui/react-icons";
import styles from "./views.module.scss";

type ViewRow = {
  title: string;
  index: number;
  onSelectView: (index: number) => void;
  onTitleSave: (index: number, title: string) => void;
  onDeleteView: (id: number) => void;
};

export default function ViewRowComponent<T>(props: ViewRow) {
  const [editView, setEditView] = useState(false);
  const [localTitle, setLocalTitle] = useState(props.title);

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
              onSubmit={() => props.onTitleSave(props.index, localTitle)}
              className={classNames(styles.viewEdit, "torq-input-field")}
            >
              <input type="text" autoFocus={true} onChange={handleInputChange} value={localTitle} />
              <button type={"submit"}>
                <SaveIcon />
              </button>
            </form>
          ) : (
            <div className={styles.viewSelect} onClick={() => props.onSelectView(props.index)}>
              <div>{props.title}</div>
              <div className={styles.editView} onClick={() => setEditView(true)}>
                <EditIcon />
              </div>
            </div>
          )}

          <div className={styles.removeView} onClick={() => props.onDeleteView(props.index)}>
            <RemoveIcon />
          </div>
        </div>
      )}
    </Draggable>
  );
}
