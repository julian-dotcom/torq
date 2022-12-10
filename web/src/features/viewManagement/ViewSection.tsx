import styles from "./views.module.scss";
import { DragDropContext, Droppable, DropResult, ResponderProvided } from "react-beautiful-dnd";
import { AddSquare20Regular as AddIcon } from "@fluentui/react-icons";
import Button, { buttonColor, buttonSize } from "components/buttons/Button";
import { AllViewsResponse, TableResponses, ViewOrderInterface, ViewResponse } from "./types";
import { useCreateTableViewMutation, useUpdateTableViewsOrderMutation } from "./viewsApiSlice";
import ViewRowComponent from "./ViewRow";
import { useStrictDroppable } from "utils/UseStrictDroppable";
import { addView, selectViews, updateViewsOrder } from "./viewSlice";
import { useAppDispatch, useAppSelector } from "store/hooks";
import useTranslations from "services/i18n/useTranslations";

type ViewSection<T> = {
  page: keyof AllViewsResponse;
  defaultView: ViewResponse<T>;
};

function ViewsPopover<T>(props: ViewSection<T>) {
  const { t } = useTranslations();
  const dispatch = useAppDispatch();
  const viewResponse = useAppSelector(selectViews)(props.page);

  const [updateTableViewsOrder] = useUpdateTableViewsOrderMutation();
  const [createTableView] = useCreateTableViewMutation();

  const droppableContainerId = "views-list-droppable-" + props.page;

  // function handleSaveView(view: ViewResponse<T>) {
  //   if (view.id) {
  //     updateTableView({
  //       id: view.id,
  //       view: view.view as ViewInterface<TableResponses>,
  //     });
  //   } else {
  //     createTableView({
  //       page: view.page,
  //       view: view.view as ViewInterface<TableResponses>,
  //     });
  //   }
  // }

  // function handleSaveTitle(index: number, title: string) {
  //   const view = views.views[index];
  //   if (view.id !== undefined) {
  //     updateTableView({
  //       id: view.id,
  //       view: { ...view.view, title: title } as ViewInterface<TableResponses>,
  //     });
  //   } else {
  //     createTableView({
  //       page: view.page,
  //       view: { ...view.view, title: title } as ViewInterface<TableResponses>,
  //       viewOrder: index,
  //     });
  //   }
  // }

  // function handleDeleteView(index: number) {
  //   const view = props.views[index];
  //   if (view.id !== undefined) {
  //     deleteTableView({ id: view.id });
  //   } else {
  //     props.views.splice(index, 1);
  //   }
  // }

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
                  return (
                    <ViewRowComponent
                      title={view.view.title}
                      id={view.id}
                      viewIndex={viewIndex}
                      page={view.page}
                      key={"view-row-" + viewIndex}
                      selected={viewResponse.selected === viewIndex}
                      singleView={singleView}
                    />
                  );
                })}
                {provided.placeholder}
              </div>
            )}
          </Droppable>
        )}
        <div className={styles.buttonsRow}>
          <Button
            buttonColor={buttonColor.ghost}
            buttonSize={buttonSize.small}
            text={t.addView}
            icon={<AddIcon />}
            onClick={() => {
              dispatch(addView({ view: props.defaultView as ViewResponse<TableResponses> }));
            }}
          />
        </div>
      </div>
    </DragDropContext>
  );
}

export default ViewsPopover;
