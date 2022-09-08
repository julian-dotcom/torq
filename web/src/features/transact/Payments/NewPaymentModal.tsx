import styles from "./newPayments.module.scss";
import { Options20Regular as OptionsIcon, MoneyHand24Regular as TransactionIconModal } from "@fluentui/react-icons";

import Button, { buttonColor, buttonPosition, ButtonWrapper } from "features/buttons/Button";
import TextInput from "features/forms/TextInput";
import { SectionContainer } from "../../section/SectionContainer";
import TextArea from "../../forms/TextArea";
import React, { ChangeEvent, useState } from "react";
import Switch from "../../inputs/Slider/Switch";
import PopoutPageTemplate from "../../templates/popoutPageTemplate/PopoutPageTemplate";
import ProgressHeader, { ProgressStepState, Step } from "../../progressTabs/ProgressHeader";
import ProgressTabs, { ProgressTabContainer } from "../../progressTabs/ProgressTab";

type NewPaymentModalProps = {
  show: boolean;
  modalCloseHandler: Function;
};

const LnPayrequestMainnetRegEx = /lnbc[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestTestnetRegEx = /lntb[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestSignetRegEx = /lnbs[0-9][0-9a-zA-Z]*/gm;
const LnPayrequestRegtestRegEx = /lnbcrt[0-9][0-9a-zA-Z]*/gm;
const btcAddressRegEx = /bc1[0-9a-zA-Z]{39,59}/gm;

function NewPaymentModal(props: NewPaymentModalProps) {
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);

  let handleAdvancedToggle = () => {
    setExpandAdvancedOptions(!expandAdvancedOptions);
  };

  const [destinationType, setDestinationType] = useState("");
  const [destination, setDestination] = useState("");

  const [destState, setDestState] = useState(ProgressStepState.active);
  const [confirmState, setConfirmState] = useState(ProgressStepState.disabled);
  const [processState, setProcessState] = useState(ProgressStepState.disabled);
  const [stepIndex, setStepIndex] = useState(0);

  let closeAndReset = () => {
    setStepIndex(0);
    setDestState(ProgressStepState.active);
    setConfirmState(ProgressStepState.disabled);
    setProcessState(ProgressStepState.disabled);
    props.modalCloseHandler();
  };

  const setDestinationHandler = (e: ChangeEvent<HTMLTextAreaElement>) => {
    setDestination(e.target.value);
    if (e.target.value.match(LnPayrequestMainnetRegEx)) {
      setDestinationType("Mainnet Lightning Invoice");
    } else if (e.target.value.match(LnPayrequestTestnetRegEx)) {
      setDestinationType("Testnet Lightning Invoice");
    } else if (e.target.value.match(LnPayrequestSignetRegEx)) {
      setDestinationType("Signet Lightning Invoice");
    } else if (e.target.value.match(LnPayrequestRegtestRegEx)) {
      setDestinationType("Regtest Lightning Invoice");
    } else if (e.target.value.match(btcAddressRegEx)) {
      setDestinationType("Bitcoin Address");
    } else {
      setDestinationType("");
    }
  };

  return (
    <PopoutPageTemplate title={"New Payment"} show={props.show} onClose={closeAndReset} icon={<TransactionIconModal />}>
      <ProgressHeader modalCloseHandler={closeAndReset}>
        <Step label={"Destination"} state={destState} last={false} />
        <Step label={"Confirm"} state={confirmState} last={false} />
        <Step label={"Paying"} state={processState} last={true} />
      </ProgressHeader>

      <ProgressTabs showTabIndex={stepIndex}>
        <ProgressTabContainer>
          <div className={styles.destination}>
            <div className={styles.destinationWrapper}>
              <div className={styles.labelWrapper}>
                <label htmlFor={"destination"} className={styles.destinationLabel}>
                  Destination
                </label>
                {destinationType && <span className={styles.destinationType}>{destinationType}</span>}
              </div>
              <textarea
                id={"destination"}
                name={"destination"}
                placeholder={"E.g. Lightning Invoice, PubKey, On-chain Address"}
                className={styles.destinationTextArea}
                value={destination}
                onChange={setDestinationHandler}
                rows={3}
              />
            </div>
          </div>
          <ButtonWrapper
            rightChildren={
              <Button
                text={"Next"}
                disabled={!destination}
                onClick={() => {
                  if (destination) {
                    setStepIndex(1);
                    setDestState(ProgressStepState.completed);
                    setConfirmState(ProgressStepState.active);
                  }
                }}
                buttonColor={buttonColor.subtle}
              />
            }
          />
        </ProgressTabContainer>
        <ProgressTabContainer>
          <div className={styles.amountWrapper}>
            <div className={styles.label}>You are paying</div>
            <div className={styles.amount}>200,000 sat</div>
            <div className={styles.label}>To</div>
            <div className={styles.destinationPreview}>{destination}</div>
          </div>
          <TextArea label={"Description (only seen by you)"} />
          <SectionContainer
            title={"Advanced Options"}
            icon={OptionsIcon}
            expanded={expandAdvancedOptions}
            handleToggle={handleAdvancedToggle}
          >
            <TextInput label={"Max fee"} />
            <TextInput label={"Max fee rate"} />
            <TextInput label={"Timeout"} />
            <Switch label={"Allow self payment"} labelPosition={"left"} checked={true} />
          </SectionContainer>

          <ButtonWrapper
            leftChildren={
              <Button
                text={"Back"}
                onClick={() => {
                  setStepIndex(0);
                  setDestState(ProgressStepState.completed);
                  setConfirmState(ProgressStepState.active);
                }}
                buttonColor={buttonColor.ghost}
              />
            }
            rightChildren={
              <Button
                text={"Confirm"}
                onClick={() => {
                  setStepIndex(2);
                  setConfirmState(ProgressStepState.completed);
                  setProcessState(ProgressStepState.processing);
                }}
                buttonColor={buttonColor.green}
              />
            }
          />
        </ProgressTabContainer>
        <ProgressTabContainer>
          Payment in progress
          <Button
            text={"New Payment"}
            onClick={() => {
              setStepIndex(0);
              setDestState(ProgressStepState.active);
              setConfirmState(ProgressStepState.disabled);
              setProcessState(ProgressStepState.disabled);
            }}
            buttonColor={buttonColor.secondary}
          />
        </ProgressTabContainer>
      </ProgressTabs>
    </PopoutPageTemplate>
  );
}

export default NewPaymentModal;
