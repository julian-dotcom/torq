import styles from "./views.module.scss";
import { DragDropContext, Droppable } from "react-beautiful-dnd";
import { AddSquare20Regular as AddIcon } from "@fluentui/react-icons";
import TabButton from "components/buttons/TabButton";
import Popover from "features/popover/Popover";
import Button, { buttonColor, buttonSize } from "components/buttons/Button";
import { ViewInterface, ViewInterfaceResponse, ViewOrderInterface } from "./types";
import { useCreateTableViewMutation, useUpdateTableViewsOrderMutation } from "./viewsApiSlice";
import ViewRow from "./ViewRow";

type ViewsPopover<T> = {
  page: string;
  views: Array<ViewInterface<T>>;
  DefaultView: ViewInterface<T>;
  onSelectView: (index: number) => void;
  selectedView: number;
};

function ViewsPopover<T>(props: ViewsPopover<T>) {
  const [updateTableViewsOrder] = useUpdateTableViewsOrderMutation();
  const [createTableView] = useCreateTableViewMutation();

  const droppableContainerId = "views-list-droppable-" + props.page;

  const onDragEnd = (result: any) => {
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

    const order: ViewOrderInterface[] = newViewsOrder.map((view, index) => {
      return { id: view.id, view_order: index };
    });
    updateTableViewsOrder(order);
  };

  const button = <TabButton title={props.views[props.selectedView].title} dropDown={true} />;

  const singleView = props.views.length <= 1;
  return (
    <Popover button={button}>
      <DragDropContext onDragEnd={onDragEnd}>
        <div className={styles.viewsPopoverContent}>
          <Droppable droppableId={droppableContainerId}>
            {(provided) => (
              <div className={styles.viewRows} ref={provided.innerRef} {...provided.droppableProps}>
                {props.views.map((view, index) => {
                  return (
                    <ViewRow
                      view={view}
                      key={index}
                      index={index}
                      onSelectView={props.onSelectView}
                      singleView={singleView}
                    />
                  );
                })}
                {provided.placeholder}
              </div>
            )}
          </Droppable>
          <div className={styles.buttonsRow}>
            <Button
              buttonColor={buttonColor.ghost}
              buttonSize={buttonSize.small}
              text={"Add View"}
              icon={<AddIcon />}
              onClick={() => {
                createTableView(props.DefaultView as ViewInterfaceResponse);
              }}
            />
          </div>
        </div>
      </DragDropContext>
    </Popover>
  );
}

export default ViewsPopover;
