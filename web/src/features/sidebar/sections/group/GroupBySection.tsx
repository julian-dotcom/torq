import { VirtualNetwork20Regular as ChannelsIcon, Iot20Regular as PeersIcon } from "@fluentui/react-icons";
import styles from "./group-section.module.scss";
import classNames from "classnames";
import View from "features/viewManagement/View";

type GroupPopoverProps<T> = {
  view: View<T>;
};

function GroupBySection<T>(props: GroupPopoverProps<T>) {
  return (
    <div className={styles.groupRowWrapper}>
      <button
        className={classNames(styles.groupRow, {
          [styles.groupRowSelected]: props.view.groupBy == "channels",
        })}
        onClick={() => {
          props.view.groupBy = "channels";
        }}
      >
        <div className="icon">
          <ChannelsIcon />
        </div>
        <div>Channels</div>
      </button>
      <button
        className={classNames(styles.groupRow, {
          [styles.groupRowSelected]: props.view.groupBy == "peers",
        })}
        onClick={() => {
          props.view.groupBy = "peers";
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
