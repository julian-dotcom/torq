import { VirtualNetwork20Regular as ChannelsIcon, Iot20Regular as PeersIcon } from "@fluentui/react-icons";
import styles from "./group-section.module.scss";
import classNames from "classnames";
import { AllViewsResponse } from "features/viewManagement/types";
import { updateGroupBy } from "features/viewManagement/viewSlice";
import { useAppDispatch } from "store/hooks";

type GroupPopoverProps = {
  page: keyof AllViewsResponse;
  viewIndex: number;
  groupBy?: "channels" | "peers";
};

function GroupBySection(props: GroupPopoverProps) {
  const dispatch = useAppDispatch();
  const handleUpdate = (by: "channels" | "peers") => {
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
