import { FormEvent, MouseEvent, useState } from "react";
import classNames from "classnames";
import { Draggable } from "react-beautiful-dnd";
import {
  Delete16Regular as RemoveIcon,
  Edit16Regular as EditIcon,
  ReOrder16Regular as DragHandle,
  Save16Regular as SaveIcon,
  Checkmark16Regular as ChckmarkIcon,
} from "@fluentui/react-icons";
import styles from "./views.module.scss";
import { AllViewsResponse } from "./types";
import { useDeleteTableViewMutation } from "./viewsApiSlice";
import { useAppDispatch } from "store/hooks";
import { deleteView, updateSelectedView, updateViewTitle } from "./viewSlice";
import Input from "components/forms/input/Input";
import { InputSizeVariant } from "components/forms/input/variants";

type ViewRow = {
  id?: number;
  title: string;
  page: keyof AllViewsResponse;
  viewIndex: number;
  selected: boolean;
  dirty?: boolean;
  singleView: boolean;
  onSaveView: (viewIndex: number) => void;
};

export default function ViewRowComponent(props: ViewRow) {
  const dispatch = useAppDispatch();
  const [deleteTableView] = useDeleteTableViewMutation();
  const [editView, setEditView] = useState(false);
  const [localTitle, setLocalTitle] = useState(props.title);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function handleInputChange(e: any) {
    setLocalTitle(e.target.value);
  }

  function handleInputSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setEditView(false);
    dispatch(updateViewTitle({ page: props.page, viewIndex: props.viewIndex, title: localTitle }));
  }

  function handleSelectView(event: MouseEvent<HTMLDivElement>) {
    event.preventDefault();
    event.stopPropagation();
    dispatch(updateSelectedView({ page: props.page, viewIndex: props.viewIndex }));
  }

  return (
    <Draggable draggableId={`draggable-view-id-${props.viewIndex}`} index={props.viewIndex} isDragDisabled={!props.id}>
      {(provided, snapshot) => (
        <form
          onSubmit={handleInputSubmit}
          className={classNames(styles.viewRow, { dragging: snapshot.isDragging, [styles.selected]: props.selected })}
          ref={provided.innerRef}
          {...provided.draggableProps}
        >
          <div
            className={classNames(styles.viewRowDragHandle, { [styles.dragDisabled]: !props.id })}
            {...provided.dragHandleProps}
          >
            <DragHandle />
          </div>

          {editView ? (
            <Input
              type="text"
              autoFocus={true}
              onChange={handleInputChange}
              value={localTitle}
              sizeVariant={InputSizeVariant.inline}
            />
          ) : (
            <div className={styles.viewSelect} onClick={handleSelectView}>
              <div className={styles.viewSelectTitle}>{props.title}</div>
            </div>
          )}
          {editView ? (
            <button type={"submit"} className={styles.viewRowEdit}>
              <ChckmarkIcon />
            </button>
          ) : (
            <div className={styles.viewRowEdit} onClick={() => setEditView(true)}>
              <EditIcon />
            </div>
          )}
          <div
            className={classNames(styles.viewRowSave, { [styles.clean]: !props.dirty })}
            onClick={() => {
              props.onSaveView(props.viewIndex);
            }}
          >
            <SaveIcon />
          </div>
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
