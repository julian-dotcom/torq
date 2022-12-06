import { FormEvent, useState } from "react";
import classNames from "classnames";
import { Draggable } from "react-beautiful-dnd";
import {
  Dismiss20Regular as RemoveIcon,
  Edit16Regular as EditIcon,
  Save20Regular as SaveIcon,
  Reorder20Regular as DragHandle,
} from "@fluentui/react-icons";
import styles from "./views.module.scss";
import { AllViewsResponse } from "./types";
import { useDeleteTableViewMutation } from "./viewsApiSlice";
import { useAppDispatch } from "../../store/hooks";
import { deleteView, updateSelectedView, updateViewTitle } from "./viewSlice";

type ViewRow<T> = {
  id?: number;
  title: string;
  page: keyof AllViewsResponse;
  viewIndex: number;
  selected: boolean;
  singleView: boolean;
};

export default function ViewRowComponent<T>(props: ViewRow<T>) {
  const dispatch = useAppDispatch();
  const [deleteTableView] = useDeleteTableViewMutation();
  const [editView, setEditView] = useState(false);
  const [localTitle, setLocalTitle] = useState(props.title);

  function handleInputChange(e: any) {
    setLocalTitle(e.target.value);
  }

  function handleInputSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setEditView(false);
    dispatch(updateViewTitle({ page: props.page, viewIndex: props.viewIndex, title: localTitle }));
  }

  function handleSelectView() {
    dispatch(updateSelectedView({ page: props.page, viewIndex: props.viewIndex }));
  }

  return (
    <Draggable draggableId={`draggable-view-id-${props.viewIndex}`} index={props.viewIndex}>
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
            <form onSubmit={handleInputSubmit} className={classNames(styles.viewEdit, "torq-input-field")}>
              <input type="text" autoFocus={true} onChange={handleInputChange} value={localTitle} />
              <button type={"submit"}>
                <SaveIcon />
              </button>
            </form>
          ) : (
            <div className={styles.viewSelect} onClick={handleSelectView}>
              <div>{props.title}</div>
              <div className={styles.editView} onClick={() => setEditView(true)}>
                <EditIcon />
              </div>
            </div>
          )}

          {!props.singleView && (
            <div
              className={styles.removeView}
              onClick={() => {
                if (props.id) {
                  deleteTableView({ page: props.page, id: props.id });
                } else {
                  dispatch(deleteView({ page: props.page, viewIndex: props.viewIndex }));
                }
              }}
            >
              <RemoveIcon />
            </div>
          )}
        </div>
      )}
    </Draggable>
  );
}
