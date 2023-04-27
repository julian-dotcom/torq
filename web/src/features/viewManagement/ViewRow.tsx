import { ChangeEvent, FormEvent, MouseEvent, useState } from "react";
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
import { useAppDispatch, useAppSelector } from "store/hooks";
import { deleteView, selectViews, updateSelectedView, updateViewTitle } from "./viewSlice";
import Input from "components/forms/input/Input";
import { InputSizeVariant } from "components/forms/input/variants";
import { userEvents } from "utils/userEvents";
import { deserialiseQuery } from "../sidebar/sections/filter/filter";

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
  const viewResponse = useAppSelector(selectViews)(props.page);
  const view = viewResponse.views[props.viewIndex].view;
  const { track } = userEvents();

  function handleInputChange(e: ChangeEvent<HTMLInputElement>) {
    setLocalTitle(e.target.value);
  }

  function handleInputSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setEditView(false);
    track(`View Update Title`, {
      viewPage: props.page,
      viewIndex: props.viewIndex,
      viewNewTitle: localTitle,
      viewOldTitle: props.title,
    });
    dispatch(updateViewTitle({ page: props.page, viewIndex: props.viewIndex, title: localTitle }));
  }

  function handleSelectView(event: MouseEvent<HTMLDivElement>) {
    event.preventDefault();
    event.stopPropagation();
    track(`View Selected`, {
      viewPage: props.page,
      newSelectedView: props.viewIndex,
      newSelectedViewTitle: props.title,
      previousView: props.selected,
      previousViewTitle: props.title,
    });
    dispatch(updateSelectedView({ page: props.page, viewIndex: props.viewIndex }));
  }

  const handleDeleteView = () => {
    if (props.id) {
      deleteTableView({ page: props.page, id: props.id });
    } else {
      if (view) {
        track(`View Deleted`, {
          viewPage: props.page,
          viewCount: viewResponse.views.length,
          viewIndex: props.viewIndex,
          viewName: viewResponse.views[props.viewIndex].view.title,
          viewColumns: (view.columns || []).map((c) => c.heading),
          viewSortedBy: (view.sortBy || []).map((s) => {
            return { key: s.key, direction: s.direction };
          }),
          viewFilterCount: deserialiseQuery(view.filters).length,
        });
      }

      dispatch(deleteView({ page: props.page, viewIndex: props.viewIndex }));
    }
  };

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
            data-intercom-target="view-row-drag-handle"
            className={classNames(styles.viewRowDragHandle, { [styles.dragDisabled]: !props.id })}
            {...provided.dragHandleProps}
          >
            <DragHandle />
          </div>

          {editView ? (
            <Input
              data-intercom-target="view-title-input"
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
            <button type={"submit"} className={styles.viewRowEdit} data-intercom-target={"exit-edit-view-title-button"}>
              <ChckmarkIcon />
            </button>
          ) : (
            <div
              className={styles.viewRowEdit}
              onClick={() => setEditView(true)}
              data-intercom-target={"edit-view-title-button"}
            >
              <EditIcon />
            </div>
          )}
          <div
            data-intercom-target={"save-view-title"}
            className={classNames(styles.viewRowSave, { [styles.clean]: !props.dirty })}
            onClick={() => {
              props.onSaveView(props.viewIndex);
            }}
          >
            <SaveIcon />
          </div>
          {!props.singleView && (
            <div
              data-intercom-target={"delete-view-button"}
              className={styles.removeView}
              onClick={() => {
                handleDeleteView();
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
