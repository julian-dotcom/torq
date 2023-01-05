import { Options20Regular as OptionsIcon } from "@fluentui/react-icons";
import Button, { ColorVariant, ButtonWrapper } from "components/buttons/Button";
import { ChangeEvent, useState } from "react";
import Input from "components/forms/input/Input";
import { ProgressStepState } from "features/progressTabs/ProgressHeader";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import { SectionContainer } from "features/section/SectionContainer";

import styles from "./newPayments.module.scss";
import { PaymentType, PaymentTypeLabel } from "./types";
import { MutationTrigger } from "@reduxjs/toolkit/dist/query/react/buildHooks";
import { BaseQueryFn, FetchArgs, FetchBaseQueryError, MutationDefinition } from "@reduxjs/toolkit/query";
import { SendOnChainRequest } from "types/api";
import LargeAmountInput from "components/forms/largeAmountInput/LargeAmountInput";
import useTranslations from "services/i18n/useTranslations";

type BtcStepProps = {
  sendCoinsMutation: MutationTrigger<
    MutationDefinition<
      SendOnChainRequest,
      BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError>,
      "channels" | "settings" | "tableView" | "nodeConfigurations",
      any, // eslint-disable-line @typescript-eslint/no-explicit-any
      "api"
    >
  >;
  amount: number;
  setAmount: (amount: number) => void;
  destinationType: PaymentType;
  destination: string;
  setStepIndex: (index: number) => void;
  setDestState: (state: ProgressStepState) => void;
  setConfirmState: (state: ProgressStepState) => void;
  setProcessState: (state: ProgressStepState) => void;
};

export default function OnChanPaymentDetails(props: BtcStepProps) {
  const { t } = useTranslations();
  const [expandAdvancedOptions, setExpandAdvancedOptions] = useState(false);
  const [satPerVbyte, setSatPerVbyte] = useState<number | undefined>(undefined);
  const [description, setDescription] = useState<string | undefined>(undefined);

  return (
    <ProgressTabContainer>
      <div className={styles.amountWrapper}>
        <span className={styles.destinationType}>{PaymentTypeLabel[props.destinationType] + " Detected"}</span>
        <div className={styles.amount}>
          <LargeAmountInput
            value={props.amount === 0 ? undefined : props.amount}
            onChange={(values) => {
              props.setAmount(values || 0);
            }}
          />
          {/*<NumberFormat*/}
          {/*  className={styles.amountInput}*/}
          {/*  suffix={" sat"}*/}
          {/*  thousandSeparator=","*/}
          {/*  value={props.amount === 0 ? undefined : props.amount}*/}
          {/*  placeholder={"0 sat"}*/}
          {/*  onValueChange={(values: NumberFormatValues) => {*/}
          {/*    props.setAmount(values.floatValue || 0);*/}
          {/*  }}*/}
          {/*/>*/}
        </div>
        <div className={styles.label}>To</div>
        <div className={styles.destinationPreview}>{props.destination}</div>
      </div>
      <div className={styles.destinationWrapper}>
        <div className={styles.labelWrapper}>
          <label htmlFor={"destination"} className={styles.destinationLabel}>
            Description (only seen by you)
          </label>
        </div>
        <textarea
          id={"lnDescription"}
          name={"lnDescription"}
          className={styles.destinationTextArea}
          autoComplete="off"
          value={description}
          onChange={(e: ChangeEvent<HTMLTextAreaElement>) => {
            setDescription(e.target.value);
          }}
          rows={3}
        />
      </div>
      <SectionContainer
        title={"Advanced Options"}
        icon={OptionsIcon}
        expanded={expandAdvancedOptions}
        handleToggle={() => {
          setExpandAdvancedOptions(!expandAdvancedOptions);
        }}
      >
        <Input
          label={"Sat per vByte"}
          type={"number"}
          value={satPerVbyte}
          onChange={(e: ChangeEvent<HTMLInputElement>) => {
            setSatPerVbyte(e.target.valueAsNumber);
          }}
        />
      </SectionContainer>

      <ButtonWrapper
        className={styles.customButtonWrapperStyles}
        leftChildren={
          <Button
            onClick={() => {
              props.setStepIndex(0);
              props.setDestState(ProgressStepState.completed);
              props.setConfirmState(ProgressStepState.active);
            }}
            buttonColor={ColorVariant.primary}
          >
            {t.back}
          </Button>
        }
        rightChildren={
          <Button
            onClick={() => {
              props.setStepIndex(2);
              props.setConfirmState(ProgressStepState.completed);
              props.setProcessState(ProgressStepState.processing);
              props.sendCoinsMutation({
                address: props.destination,
                nodeId: 1,
                amountSat: props.amount,
              });
            }}
            buttonColor={ColorVariant.success}
          >
            {t.confirm}
          </Button>
        }
      />
    </ProgressTabContainer>
  );
}
