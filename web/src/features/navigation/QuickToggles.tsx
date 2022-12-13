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
              buttonColor={activeNetwork === Network.Mainnet ? buttonColor.primary : buttonColor.subtle}
              text="Mainnet"
              onClick={() => dispatch(setActiveNetwork(Network.Mainnet))}
            ></Button>
            <Button
              buttonColor={activeNetwork === Network.Testnet ? buttonColor.primary : buttonColor.subtle}
              text="Testnet"
              onClick={() => dispatch(setActiveNetwork(Network.Testnet))}
            ></Button>
            <Button
              buttonColor={activeNetwork === Network.Regtest ? buttonColor.primary : buttonColor.subtle}
              text="Regtest"
              onClick={() => dispatch(setActiveNetwork(Network.Regtest))}
            ></Button>
            <Button
              buttonColor={activeNetwork === Network.Simnet ? buttonColor.primary : buttonColor.subtle}
              text="Simnet"
              onClick={() => dispatch(setActiveNetwork(Network.Simnet))}
            ></Button>
          </div>
        </Popover>
      </div>
    </div>
  );
}

export default QuickToggles;
