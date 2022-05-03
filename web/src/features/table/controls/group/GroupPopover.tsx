import Popover from "../../../popover/Popover";
import DefaultButton from "../../../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../../../store/hooks";
import { selectGroupBy, updateGroupBy } from "../../tableSlice";
import {
  ArrowJoin20Regular as GroupIcon,
  VirtualNetwork20Regular as ChannelsIcon,
  Iot20Regular as PeersIcon,
  Tag20Regular as TagsIcon,
} from "@fluentui/react-icons";
import styles from "./group_popover.module.scss";
import classNames from "classnames";

const GroupPopover = () => {
  const dispatch = useAppDispatch();
  const groupedBy = useAppSelector(selectGroupBy) || "channels";

  let popOverButton = (
    <DefaultButton
      isOpen={groupedBy === "peers"}
      icon={<GroupIcon />}
      text={groupedBy === "peers" ? "Peers" : "Channels"}
      className={"collapse-tablet "}
    />
  );

  return (
    <Popover button={popOverButton}>
      <div className={styles.groupRowWrapper}>
        <button
          className={classNames(styles.groupRow, {
            [styles.groupRowSelected]: groupedBy == "channels",
          })}
          onClick={() => {
            dispatch(updateGroupBy({ groupBy: "channels" }));
          }}
        >
          <div className="icon">
            <ChannelsIcon />
          </div>
          <div>Channels</div>
        </button>
        <button
          className={classNames(styles.groupRow, {
            [styles.groupRowSelected]: groupedBy == "peers",
          })}
          onClick={() => {
            dispatch(updateGroupBy({ groupBy: "peers" }));
          }}
        >
          <div className="icon">
            <PeersIcon />
          </div>
          <div>Peers</div>
        </button>
        {/*<button*/}
        {/*  className={classNames(styles.groupRow, {[styles.groupRowSelected]: (groupedBy == 'tags')})}*/}
        {/*  onClick={() =>{*/}
        {/*    dispatch(updateGroupBy({groupBy: 'tags'}))*/}
        {/*  }}*/}
        {/*>*/}
        {/*  <div className="icon">*/}
        {/*    <TagsIcon/>*/}
        {/*  </div>*/}
        {/*  <div>Tags</div>*/}
        {/*</button>*/}
      </div>
    </Popover>
  );
};

export default GroupPopover;
