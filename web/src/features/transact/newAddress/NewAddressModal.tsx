import { MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";
import { useGetNodeConfigurationsQuery, WS_URL } from "apiSlice";
import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
import ProgressHeader, { ProgressStepState, Step } from "features/progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import PopoutPageTemplate from "features/templates/popoutPageTemplate/PopoutPageTemplate";
import { useState } from "react";
import { useNavigate } from "react-router";
import useWebSocket from "react-use-websocket";
import styles from "features/transact/newAddress/newAddress.module.scss";
import useTranslations from "services/i18n/useTranslations";
import { nodeConfiguration } from "apiTypes";
import Select from "features/forms/Select";
import Note, { NoteType } from "features/note/Note";
import {
  DetailsContainer,
  DetailsRowLinkAndCopy,
} from "features/templates/popoutPageTemplate/popoutDetails/PopoutDetails";
import { StatusIcon } from "features/templates/popoutPageTemplate/popoutDetails/StatusIcon";
import mixpanel from "mixpanel-browser";

export type NewAddressRequest = {
  nodeId: number;
  addressType: string;
};

export type NewAddressResponse = {
  requestId: string;
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

  const { data: nodeConfigurations } = useGetNodeConfigurationsQuery();

  let nodeConfigurationOptions: Array<{ value: number; label?: string }> = [{ value: 0, label: "Select a local node" }];
  if (nodeConfigurations !== undefined) {
    nodeConfigurationOptions = nodeConfigurations.map((nodeConfiguration: nodeConfiguration) => {
      return { value: nodeConfiguration.nodeId, label: nodeConfiguration.name };
    });
  }

  const addressTypeOptions = [
    { label: t.p2wpkh, value: AddressType.P2WPKH }, // Wrapped Segwit
    { label: t.p2wkh, value: AddressType.P2WKH }, // Segwit
    { label: t.p2tr, value: AddressType.P2TR }, // Taproot
  ];

  const [response, setResponse] = useState<NewAddressResponse>();
  const [newAddressError, setNewAddressError] = useState("");
  const [selectedNodeId, setSelectedNodeId] = useState<number>(nodeConfigurationOptions[0].value);

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

  function onNewAddressResponse(resp: NewAddressResponse) {
    setResponse(resp);
    if (resp.address.length) {
      setDoneState(ProgressStepState.completed);
    } else {
      setNewAddressError(resp.failureReason);
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
    shouldReconnect: () => true,
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
      requestId: "randId",
      type: "newAddress",
      newAddressRequest: {
        nodeId: selectedNodeId,
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
          <div className={styles.nodeSelectionWrapper}>
            <div className={styles.nodeSelection}>
              <Select
                label={t.yourNode}
                onChange={(newValue: unknown) => {
                  if (typeof newValue === "number") {
                    setSelectedNodeId(newValue);
                    mixpanel.track("New Address - Select Node", { nodeId: newValue });
                  }
                }}
                options={nodeConfigurationOptions}
                value={nodeConfigurationOptions.find((option) => option.value === selectedNodeId)}
              />
            </div>
          </div>
          <div className={styles.addressTypeWrapper}>
            <div className={styles.addressTypes}>
              {addressTypeOptions.map((addType, index) => {
                return (
                  <Button
                    disabled={!selectedNodeId}
                    buttonColor={ColorVariant.primary}
                    key={index + addType.label}
                    onClick={() => {
                      if (selectedNodeId) {
                        handleClickNext(addType.value);
                        mixpanel.track("New Address - Select Address Type", { addressType: addType.label });
                      }
                    }}
                  >
                    {addType.label}
                  </Button>
                );
              })}
            </div>
          </div>
        </ProgressTabContainer>

        <ProgressTabContainer>
          {newAddressError && (
            <Note title={"Error"} noteType={NoteType.error}>
              {newAddressError}
            </Note>
          )}
          {!newAddressError && <StatusIcon state={response?.address ? "success" : "processing"} />}
          <DetailsContainer>
            <DetailsRowLinkAndCopy label={"Address:"} copy={response?.address}>
              {response?.address}
            </DetailsRowLinkAndCopy>
          </DetailsContainer>
          <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                onClick={() => {
                  setAddressTypeState(ProgressStepState.active);
                  setDoneState(ProgressStepState.disabled);
                  setStepIndex(0);
                  setResponse(undefined);
                  mixpanel.track("New Address - Back to Address Type");
                }}
                buttonColor={ColorVariant.primary}
              >
                {t.newAddress}
              </Button>
            }
          />
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NewAddressModal;
