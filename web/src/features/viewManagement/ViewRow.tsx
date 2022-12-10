import { FormEvent, useState } from "react";
import classNames from "classnames";
import { Draggable } from "react-beautiful-dnd";
import {
  Delete16Regular as RemoveIcon,
  Edit16Regular as EditIcon,
  ReOrder16Regular as DragHandle,
  Save16Regular as SaveIcon,
} from "@fluentui/react-icons";
import styles from "./views.module.scss";
import { AllViewsResponse } from "./types";
import { useDeleteTableViewMutation } from "./viewsApiSlice";
import { useAppDispatch } from "store/hooks";
import { deleteView, updateSelectedView, updateViewTitle } from "./viewSlice";
import Input from "components/forms/input/Input";
import { InputSizeVariant } from "components/forms/input/variants";

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
        <form
          onSubmit={handleInputSubmit}
          className={classNames(styles.viewRow, { dragging: snapshot.isDragging, [styles.selected]: props.selected })}
          ref={provided.innerRef}
          {...provided.draggableProps}
        >
          <div className={styles.viewRowDragHandle} {...provided.dragHandleProps}>
            <DragHandle />
          </div>

          {editView ? (
            <Input
              type="text"
              autoFocus={true}
              onChange={handleInputChange}
              value={localTitle}
              sizeVariant={InputSizeVariant.small}
            />
          ) : (
            <div className={styles.viewSelect} onClick={handleSelectView}>
              <div>{props.title}</div>
            </div>
          )}
          {editView ? (
            <button type={"submit"}>
              <div className={styles.viewRowEdit}>
                <SaveIcon />
              </div>
            </button>
          ) : (
            <div className={styles.viewRowEdit} onClick={() => setEditView(true)}>
              <EditIcon />
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
        </form>
      )}
    </Draggable>
  );
}
