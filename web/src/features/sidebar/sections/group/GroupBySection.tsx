import { VirtualNetwork20Regular as ChannelsIcon, Iot20Regular as PeersIcon } from "@fluentui/react-icons";
import styles from "./group-section.module.scss";
import classNames from "classnames";
import { AllViewsResponse } from "features/viewManagement/types";
import { selectViews, updateGroupBy } from "features/viewManagement/viewSlice";
import { useAppDispatch, useAppSelector } from "store/hooks";
import { userEvents } from "utils/userEvents";

type GroupPopoverProps = {
  page: keyof AllViewsResponse;
  viewIndex: number;
  groupBy?: "channels" | "peers";
};

function GroupBySection(props: GroupPopoverProps) {
  const dispatch = useAppDispatch();
  const viewResponse = useAppSelector(selectViews)(props.page);
  const view = viewResponse?.views[props.viewIndex]?.view;
  const { track } = userEvents();

  const handleUpdate = (by: "channels" | "peers") => {
    if (view) {
      track(`View Update Group By`, {
        viewPage: props.page,
        viewIndex: props.viewIndex,
        viewTitle: view.title,
        viewGroupBy: by,
      });
    }

    dispatch(
      updateGroupBy({
        page: props.page,
        viewIndex: props.viewIndex,
        groupByUpdate: by,
      })
    );
  };

  return (
    <div className={styles.groupRowWrapper}>
      <button
        className={classNames(styles.groupRow, {
          [styles.groupRowSelected]: props.groupBy == "channels",
        })}
        onClick={() => {
          handleUpdate("channels");
        }}
      >
        <div className="icon">
          <ChannelsIcon />
        </div>
        <div>Channels</div>
      </button>
      <button
        className={classNames(styles.groupRow, {
          [styles.groupRowSelected]: props.groupBy == "peers",
        })}
        onClick={() => {
          handleUpdate("peers");
        }}
      >
        <div className="icon">
          <PeersIcon />
        </div>
        <div>Peers</div>
      </button>
    </div>
  );
}

export default GroupBySection;
