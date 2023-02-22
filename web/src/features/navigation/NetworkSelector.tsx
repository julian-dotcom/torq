import { useAppDispatch, useAppSelector } from "store/hooks";
import { Globe20Regular as GlobeIcon } from "@fluentui/react-icons";
import styles from "./nav.module.scss";
import Popover from "features/popover/Popover";
import Button, { ColorVariant, SizeVariant } from "components/buttons/Button";
import { selectActiveNetwork, setActiveNetwork, Network } from "features/network/networkSlice";
import useTranslations from "services/i18n/useTranslations";
import mixpanel from "mixpanel-browser";

function NetworkSelector() {
  const { t } = useTranslations();
  const dispatch = useAppDispatch();
  const activeNetwork = useAppSelector(selectActiveNetwork);

  return (
    <div className={styles.quickToggles}>
      <Popover
        button={
          <Button
            buttonColor={ColorVariant.ghost}
            buttonSize={SizeVariant.small}
            icon={<GlobeIcon />}
            hideMobileText={true}
          />
        }
        className={"right"}
      >
        <div className={styles.quickToggleContent}>
          <Button
            buttonColor={activeNetwork === Network.MainNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.MainNet));
              mixpanel.track("Select Network", { networkSelected: "MainNet" });
            }}
          >
            {t.MainNet}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.TestNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.TestNet));
              mixpanel.track("Select Network", { networkSelected: "TestNet" });
            }}
          >
            {t.TestNet}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.RegTest ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.RegTest));
              mixpanel.track("Select Network", { networkSelected: "RegTest" });
            }}
          >
            {t.RegTest}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.SigNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.SigNet));
              mixpanel.track("Select Network", { networkSelected: "SigNet" });
            }}
          >
            {t.SigNet}
          </Button>
          <Button
            buttonColor={activeNetwork === Network.SimNet ? ColorVariant.success : ColorVariant.primary}
            onClick={() => {
              dispatch(setActiveNetwork(Network.SimNet));
              mixpanel.track("Select Network", { networkSelected: "SimNet" });
            }}
          >
            {t.SimNet}
          </Button>
        </div>
      </Popover>
    </div>
  );
}

export default NetworkSelector;
