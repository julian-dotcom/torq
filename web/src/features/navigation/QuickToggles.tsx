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
          <div className={styles.quickToggleContent}>
            <Button buttonColor={buttonColor.subtle} text="Mainnet"></Button>
            <Button buttonColor={buttonColor.subtle} text="Testnet"></Button>
            <Button buttonColor={buttonColor.primary} text="Regtest"></Button>
            <Button buttonColor={buttonColor.subtle} text="Simnet"></Button>
          </div>
        </Popover>
      </div>
    </div>
  );
}

export default QuickToggles;
