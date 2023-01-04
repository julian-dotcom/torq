import styles from "./views.module.scss";
import { DragDropContext, Droppable, DropResult, ResponderProvided } from "react-beautiful-dnd";
import { Add16Regular as AddIcon } from "@fluentui/react-icons";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { AllViewsResponse, TableResponses, ViewOrderInterface, ViewResponse } from "./types";
import {
  useCreateTableViewMutation,
  useUpdateTableViewMutation,
  useUpdateTableViewsOrderMutation,
} from "./viewsApiSlice";
import ViewRowComponent from "./ViewRow";
import { useStrictDroppable } from "utils/UseStrictDroppable";
import { addView, selectViews, updateViewsOrder } from "./viewSlice";
import { useAppDispatch, useAppSelector } from "store/hooks";
import useTranslations from "services/i18n/useTranslations";
import { useState } from "react";

type ViewSection<T> = {
  page: keyof AllViewsResponse;
  defaultView: ViewResponse<T>;
};

function ViewsPopover<T>(props: ViewSection<T>) {
  const [draftCount, setDraftcount] = useState(1);
  const { t } = useTranslations();
  const dispatch = useAppDispatch();
  const viewResponse = useAppSelector(selectViews)(props.page);

  const [updateTableViewsOrder] = useUpdateTableViewsOrderMutation();
  const [createTableView] = useCreateTableViewMutation();
  const [updateTableView] = useUpdateTableViewMutation();

  const droppableContainerId = "views-list-droppable-" + props.page;

  function handleSaveView(viewIndex: number) {
    const view = viewResponse.views[viewIndex] as ViewResponse<TableResponses>;
    if (view.id) {
      updateTableView({
        id: view.id,
        view: view.view,
      });
    } else {
      createTableView({
        index: viewIndex,
        page: view.page,
        view: view.view,
      });
    }
  }

  const onDragEnd = (result: DropResult, provided: ResponderProvided) => {
    const { destination, source } = result;

    // Dropped outside of container
    if (!destination || destination.droppableId !== droppableContainerId) {
      return;
    }

    // Position not changed
    if (destination.droppableId === source.droppableId && destination.index === source.index) {
      return;
    }

    const newViewsOrder = viewResponse.views.slice();
    const view = viewResponse.views[source.index];
    newViewsOrder.splice(source.index, 1);
    newViewsOrder.splice(destination.index, 0, view as ViewResponse<TableResponses>);
    const order: ViewOrderInterface[] = newViewsOrder.map((v, index: number) => {
      return { id: v.id, viewOrder: index };
    });
    dispatch(updateViewsOrder({ page: props.page, fromIndex: source.index, toIndex: destination.index }));
    updateTableViewsOrder(order);
  };

  // Workaround for incorrect handling of React.StrictMode by react-beautiful-dnd
  // https://github.com/atlassian/react-beautiful-dnd/issues/2396#issuecomment-1248018320
  const [strictDropEnabled] = useStrictDroppable(!viewResponse.views.length);

  const singleView = viewResponse.views.length <= 1;
  return (
    <DragDropContext onDragEnd={onDragEnd}>
      <div className={styles.viewsPopoverContent}>
        {strictDropEnabled && (
          <Droppable droppableId={droppableContainerId}>
            {(provided) => (
              <div className={styles.viewRows} ref={provided.innerRef} {...provided.droppableProps}>
                {viewResponse.views.map((view, viewIndex) => {
                  if (view.id) {
                    return (
                      <ViewRowComponent
                        title={view.view.title}
                        id={view.id}
                        viewIndex={viewIndex}
                        page={view.page}
                        dirty={view.dirty}
                        key={"view-row-" + viewIndex + "-" + view.id}
                        selected={viewResponse.selected === viewIndex}
                        singleView={singleView}
                        onSaveView={handleSaveView}
                      />
                    );
                  }
                })}
              </div>
            )}
          </Droppable>
        )}
        {viewResponse.views.findIndex((v) => v.id === undefined) !== -1 && (
          <div className={styles.unSavedViewsDivider}>{t.draftViews}</div>
        )}
        <Droppable droppableId={droppableContainerId + "-unsaved"} isDropDisabled={true}>
          {(provided) => (
            <div className={styles.viewRows} ref={provided.innerRef} {...provided.droppableProps}>
              {viewResponse.views.map((view, viewIndex) => {
                if (!view.id) {
                  return (
                    <ViewRowComponent
                      title={view.view.title}
                      id={view.id}
                      viewIndex={viewIndex}
                      page={view.page}
                      dirty={view.dirty}
                      key={"view-row-" + viewIndex + "-dirty-" + view.view.title}
                      selected={viewResponse.selected === viewIndex}
                      singleView={singleView}
                      onSaveView={handleSaveView}
                    />
                  );
                }
              })}
            </div>
          )}
        </Droppable>
        <div className={styles.buttonsRow}>
          <Button
            buttonColor={ColorVariant.primary}
            buttonSize={SizeVariant.small}
            icon={<AddIcon />}
            onClick={() => {
              const view: ViewResponse<TableResponses> = {
                ...(props.defaultView as ViewResponse<TableResponses>),
                view: {
                  ...props.defaultView.view,
                  title: t.draft + " " + draftCount,
                } as ViewResponse<TableResponses>["view"],
              };
              dispatch(addView({ view: view }));
              setDraftcount(draftCount + 1);
            }}
          >
            {t.addView}
          </Button>
        </div>
      </div>
    </DragDropContext>
  );
}

export default ViewsPopover;
