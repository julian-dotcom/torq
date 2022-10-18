import { MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";
import { useGetLocalNodesQuery, WS_URL } from "apiSlice";
import classNames from "classnames";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useState } from "react";
import { useNavigate } from "react-router";
import useWebSocket from "react-use-websocket";
import styles from "features/transact/OnChain/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { localNode } from "apiTypes";
import Select from "features/forms/Select";

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
  P2WPKH = 1,
  P2WKH = 2,
  P2TR = 4,
}

function NewAddressModal() {
  const { t } = useTranslations();

  const { data: localNodes } = useGetLocalNodesQuery();

  let localNodeOptions: Array<{ value: number; label?: string }> = [{ value: 0, label: "Select a local node" }];
  if (localNodes !== undefined) {
    localNodeOptions = localNodes.map((localNode: localNode) => {
      return { value: localNode.localNodeId, label: localNode.grpcAddress };
    });
  }

  const addressTypeOptions = [
    { label: t.p2wpkh, value: AddressType.P2WPKH }, // Wrapped Segwit
    { label: t.p2wkh, value: AddressType.P2WKH }, // Segwit
    { label: t.p2tr, value: AddressType.P2TR }, // Taproot
  ];

  const [response, setResponse] = useState<NewAddressResponse>();
  const [newAddressError, setNewAddressError] = useState("");
  const [selectedLocalNode, setSelectedLocalNode] = useState<number>(localNodeOptions[0].value);

  const [addressTypeState, setAddressTypeState] = useState(ProgressStepState.active);
  const [doneState, setDoneState] = useState(ProgressStepState.disabled);
  const [stepIndex, setStepIndex] = useState(0);

  function onNewAddressMessage(event: MessageEvent<string>) {
    const response = JSON.parse(event.data);
    setDoneState(ProgressStepState.completed);
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
  const { sendJsonMessage } = useWebSocket(WS_URL, {
    //Will attempt to reconnect on all close events, such as server shutting down
    shouldReconnect: (closeEvent) => true,
    share: true,
    onMessage: onNewAddressMessage,
  });

  const closeAndReset = () => {
    setStepIndex(0);
    setAddressTypeState(ProgressStepState.active);
    setDoneState(ProgressStepState.disabled);
  };

  const handleClickNext = (addType: AddressType) => {
    setStepIndex(1);
    setAddressTypeState(ProgressStepState.completed);
    setDoneState(ProgressStepState.processing);
    sendJsonMessage({
      reqId: "randId",
      type: "newAddress",
      newAddressRequest: {
        nodeId: selectedLocalNode,
        type: addType,
        // TODO: account empty so the default wallet account is used
        // account: {account},
      },
    });
  };

  const navigate = useNavigate();

  return (
    <PopoutPageTemplate
      title={t.header.newAddress}
      show={true}
      onClose={() => navigate(-1)}
      icon={<TransactionIconModal />}
    >
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={t.addressType} state={addressTypeState} last={false} />
        <Step label={"Done"} state={doneState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <Select
            label={t.yourNode}
            onChange={(newValue: any) => {
              setSelectedLocalNode(newValue?.value || 0);
            }}
            options={localNodeOptions}
            value={localNodeOptions.find((option) => option.value === selectedLocalNode)}
          />
          <div className={styles.addressTypeWrapper}>
            <div className={styles.addressTypes}>
              {addressTypeOptions.map((addType, index) => {
                return (
                  <Button
                    text={addType.label}
                    disabled={!selectedLocalNode}
                    buttonColor={buttonColor.subtle}
                    key={index + addType.label}
                    onClick={() => {
                      if (selectedLocalNode) {
                        handleClickNext(addType.value);
                      }
                    }}
                  />
                );
              })}
            </div>
          </div>
        </ProgressTabContainer>

        <ProgressTabContainer>
          <div className={classNames(styles.newAddressError)}>{newAddressError}</div>
          {response && <div className={classNames(styles.destinationType)}>{response.address}</div>}
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={t.newAddress}
                onClick={() => {
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
