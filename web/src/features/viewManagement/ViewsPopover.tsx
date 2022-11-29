import styles from "./views.module.scss";
import { DragDropContext, Droppable, DropResult } from "react-beautiful-dnd";
import { AddSquare20Regular as AddIcon } from "@fluentui/react-icons";
import TabButton from "components/buttons/TabButton";
import Button, { buttonColor, buttonSize } from "components/buttons/Button";
import { AllViewsResponse, TableResponses, ViewInterface, ViewOrderInterface, ViewResponse } from "./types";
import {
  useCreateTableViewMutation,
  useDeleteTableViewMutation,
  useUpdateTableViewMutation,
  useUpdateTableViewsOrderMutation,
} from "./viewsApiSlice";
import ViewRowComponent from "./ViewRow";
import { useStrictDroppable } from "../../utils/UseStrictDroppable";

type ViewsPopover<T> = {
  page: keyof AllViewsResponse;
  views: Array<ViewResponse<T>>;
  ViewTemplate: ViewInterface<T>;
  onSelectView: (index: number) => void;
  selectedView: number;
};

function ViewsPopover<T>(props: ViewsPopover<T>) {
  const [updateTableViewsOrder] = useUpdateTableViewsOrderMutation();
  const [createTableView] = useCreateTableViewMutation();
  const [updateTableView] = useUpdateTableViewMutation();
  const [deleteTableView] = useDeleteTableViewMutation();

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

  function handleSaveTitle(index: number, title: string) {
    const view = props.views[index];
    if (view.id !== undefined) {
      updateTableView({
        id: view.id,
        view: { ...view.view, title: title } as ViewInterface<TableResponses>,
      });
    } else {
      createTableView({
        page: view.page,
        view: { ...view.view, title: title } as ViewInterface<TableResponses>,
        viewOrder: index,
      });
    }
  }

  function handleDeleteView(index: number) {
    const view = props.views[index];
    if (view.id !== undefined) {
      deleteTableView({ id: view.id });
    } else {
      props.views.splice(index, 1);
    }
  }

  const onDragEnd = (result: DropResult) => {
    const { destination, source } = result;

    // Dropped outside of container
    if (!destination || destination.droppableId !== droppableContainerId) {
      return;
    }

    // Position not changed
    if (destination.droppableId === source.droppableId && destination.index === source.index) {
      return;
    }

    let newCurrentIndex = props.selectedView;
    // If moved view is the selected view
    if (source.index === props.selectedView) {
      newCurrentIndex = destination.index;
    }
    // If moved view has a source bellow the selected view
    if (source.index < props.selectedView && destination.index >= props.selectedView) {
      props.onSelectView(props.selectedView - 1);
    }
    // If moved view has a source above the selected view
    if (source.index > props.selectedView && destination.index <= props.selectedView) {
      newCurrentIndex = props.selectedView + 1;
    }

    const newViewsOrder = props.views.slice();
    newViewsOrder.splice(source.index, 1);
    newViewsOrder.splice(destination.index, 0, props.views[source.index]);
    const order: ViewOrderInterface[] = newViewsOrder.map((v, index: number) => {
      return { id: v.id, viewOrder: index };
    });
    updateTableViewsOrder(order);
  };

  const button = <TabButton title={props.views[props.selectedView].view.title} dropDown={true} />;

  // Workaround for incorrect handling of React.StrictMode by react-beautiful-dnd
  // https://github.com/atlassian/react-beautiful-dnd/issues/2396#issuecomment-1248018320
  const [strictDropEnabled] = useStrictDroppable(!props.views.length);

  const singleView = props.views.length <= 1;
  return (
    // <Popover button={button}>
    <DragDropContext onDragEnd={onDragEnd}>
      <div className={styles.viewsPopoverContent}>
        {strictDropEnabled && (
          <Droppable droppableId={droppableContainerId}>
            {(provided) => (
              <div className={styles.viewRows} ref={provided.innerRef} {...provided.droppableProps}>
                {props.views.map((view, index) => {
                  return (
                    <ViewRowComponent
                      title={view.view.title}
                      onTitleSave={handleSaveTitle}
                      onDeleteView={handleDeleteView}
                      key={"view-row-" + view.id || "unsaved-" + index}
                      index={index}
                      onSelectView={props.onSelectView}
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
            text={"Add View"}
            icon={<AddIcon />}
            onClick={() => {
              createTableView({
                page: props.page,
                view: props.ViewTemplate as ViewInterface<TableResponses>,
                viewOrder: props.views.length,
              });
            }}
          />
        </div>
      </div>
    </DragDropContext>
    // </Popover>
  );
}

export default ViewsPopover;
