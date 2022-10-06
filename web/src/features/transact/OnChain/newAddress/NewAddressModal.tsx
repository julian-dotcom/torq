import { MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";
import { WS_URL } from "apiSlice";
import classNames from "classnames";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import {useCallback, useState} from "react";
import { useNavigate } from "react-router";
import useWebSocket from "react-use-websocket";
import styles from "features/transact/OnChain/newAddress/newAddress.module.scss";
import Select, { SelectOption } from "features/forms/Select";
import useTranslations from "services/i18n/useTranslations";

export type NewAddressRequest = {
  nodeId: number;
  addressType: string;
};

export type NewAddressResponse = {
  reqId: string;
  type: string;
  status: string;
  address: string;
  failureReason: string;
};

export type NewAddressError = { id: string; type: "Error"; error: string };

export enum AddressType {
  Unknown,
  P2WPKH,
  P2WKH,
  NP2WKH,
  P2TR,
}

function NewAddressModal() {
  const { t } = useTranslations();

  const addressTypeOptions: SelectOption[] = [
    { label: t.unknown, value: AddressType.Unknown.valueOf().toString() },
    { label: t.p2wpkh, value: AddressType.P2WPKH.valueOf().toString() },
    { label: t.p2wkh, value: AddressType.P2WKH.valueOf().toString() },
    { label: t.np2wkh, value: AddressType.NP2WKH.valueOf().toString() },
    { label: t.p2tr, value: AddressType.P2TR.valueOf().toString() }
  ];

  const [response, setResponse] = useState<NewAddressResponse>();
  const [newAddressError, setNewAddressError] = useState("");
  const [addressType, setAddressType] = useState("0");

  const [addressTypeState, setAddressTypeState] = useState(ProgressStepState.active);
  const [doneState, setDoneState] = useState(ProgressStepState.disabled);
  const [stepIndex, setStepIndex] = useState(0);

  function onNewAddressMessage(event: MessageEvent<string>) {
    const response = JSON.parse(event.data);
    if (response?.type == "Error") {
      setNewAddressError(response.error);
      onNewAddressError(response as NewAddressError);
      return;
    }
    onNewAddressResponse(response as NewAddressResponse);
  }

  function onNewAddressResponse(response: NewAddressResponse) {
    setResponse(response);
    if (response.status == "SUCCEEDED") {
      setDoneState(ProgressStepState.completed);
    } else if (response.status == "FAILED") {
      setNewAddressError(response.failureReason);
      setDoneState(ProgressStepState.error);
    }
  }

  function onNewAddressError(message: NewAddressError) {
    setDoneState(ProgressStepState.error);
    console.log("error", message);
  }

  // This can also be an async getter function. See notes below on Async Urls.
  const { sendMessage, sendJsonMessage, lastMessage, lastJsonMessage, readyState, getWebSocket } = useWebSocket(
    WS_URL,
    {
      //Will attempt to reconnect on all close events, such as server shutting down
      shouldReconnect: (closeEvent) => true,
      share: true,
      onMessage: onNewAddressMessage,
    }
  );

  const closeAndReset = () => {
    setStepIndex(0);
    setAddressTypeState(ProgressStepState.active);
    setDoneState(ProgressStepState.disabled);
  };

  const handlerAddressTypeChange = (combiner: any) => {
    if (combiner.value === "") {
      setAddressType("0");
    } else {
      setAddressType(combiner.value);
    }
  };

  const handleClickNext = useCallback(
    () => {
      setStepIndex(1);
      setAddressTypeState(ProgressStepState.completed);
      setDoneState(ProgressStepState.active);
      sendJsonMessage({
        reqId: "randId",
        type: "newAddress",
        newAddressRequest: {
          // TODO: Don't just pick the first one!!!
          nodeId: 1,
          type: Number({addressType}),
          // TODO: account empty so the default wallet account is used
          // account: {account},
        }
      });
    }, [sendJsonMessage]
  );

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate title={t.header.newAddress} show={true} onClose={() => navigate(-1)} icon={<TransactionIconModal />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Type"} state={addressTypeState} last={false} />
        <Step label={"Done"} state={doneState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <div className={styles.addressTypeWrapper}>
            <div className={styles.addressType}>
              <Select
                id={"addressType"}
                name={"addressType"}
                label={t.addressType}
                value={addressType}
                className={styles.addressType}
                onChange={handlerAddressTypeChange}
                options={addressTypeOptions}
              />
            </div>
          </div>

          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={t.next}
                disabled={addressType === "0"}
                onClick={handleClickNext}
                buttonColor={buttonColor.ghost}
              />
            }
          />
        </ProgressTabContainer>

        <ProgressTabContainer>
          <div className={classNames(styles.newAddressError)}>
            {newAddressError}
          </div>
          {response && (
            <div className={classNames(styles.newAddressStatusMessage)}>
              {response.address}
            </div>
          )}
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={t.newAddress}
                  onClick={() => {
                  setAddressType("0");
                  setAddressTypeState(ProgressStepState.active);
                  setDoneState(ProgressStepState.disabled);
                  setStepIndex(0);
                  setResponse(undefined);
                }}
                buttonColor={buttonColor.subtle}
              />
            }
          />
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NewAddressModal;
