import { useAppDispatch, useAppSelector } from "store/hooks";
import { Globe20Regular as GlobeIcon } from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import Popover from "features/popover/Popover";
import Button, { buttonColor, buttonSize } from "components/buttons/Button";
import { selectActiveNetwork, setActiveNetwork, Network } from "features/network/networkSlice";

function QuickToggles() {
  const dispatch = useAppDispatch();
  const activeNetwork = useAppSelector(selectActiveNetwork);

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
            <Button
              buttonColor={activeNetwork === Network.MainNet ? buttonColor.primary : buttonColor.subtle}
              text="MainNet"
              onClick={() => dispatch(setActiveNetwork(Network.MainNet))}
            ></Button>
            <Button
              buttonColor={activeNetwork === Network.TestNet ? buttonColor.primary : buttonColor.subtle}
              text="TestNet"
              onClick={() => dispatch(setActiveNetwork(Network.TestNet))}
            ></Button>
            <Button
              buttonColor={activeNetwork === Network.RegTest ? buttonColor.primary : buttonColor.subtle}
              text="RegTest"
              onClick={() => dispatch(setActiveNetwork(Network.RegTest))}
            ></Button>
            <Button
              buttonColor={activeNetwork === Network.SigNet ? buttonColor.primary : buttonColor.subtle}
              text="SigNet"
              onClick={() => dispatch(setActiveNetwork(Network.SigNet))}
            ></Button>
            <Button
              buttonColor={activeNetwork === Network.SimNet ? buttonColor.primary : buttonColor.subtle}
              text="SimNet"
              onClick={() => dispatch(setActiveNetwork(Network.SimNet))}
            ></Button>
          </div>
        </Popover>
      </div>
    </div>
  );
}

export default QuickToggles;
