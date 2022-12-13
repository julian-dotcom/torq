/* import { useAppDispatch } from "store/hooks"; */
/* import { toggleNav } from "./navSlice"; */
/* import classNames from "classnames"; */
import { Globe20Regular as GlobeIcon } from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import Popover from "features/popover/Popover";
import Button, { buttonColor, buttonSize } from "components/buttons/Button";

function QuickToggles() {
  /* const dispatch = useAppDispatch(); */
  return (
    <div className={styles.quickToggles}>
      <div>
        <Popover
          button={
            <Button
              buttonColor={buttonColor.ghost}
              buttonSize={buttonSize.small}
              isOpen={false}
              text={``}
              icon={<GlobeIcon />}
              className={"collapse-tablet"}
            />
          }
          className={"right"}
        >
          <div>
            <h3>I have opened myself to allow someone to toggle on me</h3>
          </div>
        </Popover>
      </div>
    </div>
  );
}

export default QuickToggles;
