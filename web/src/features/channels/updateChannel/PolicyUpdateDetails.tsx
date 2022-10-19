
import { ProgressStepState } from "features/progressTabs/ProgressHeader";
import { useState } from "react";
import Button, { buttonColor, ButtonWrapper } from "features/buttons/Button";
import { ProgressTabContainer } from "features/progressTabs/ProgressTab";
import styles from "./updateChannel.module.scss";
import NumberFormat, { NumberFormatValues } from "react-number-format";


type PolicyStepProps = {
  setStepIndex: (index: number) => void;
  setChannelState: (state: ProgressStepState) => void;
  setPolicyState: (state: ProgressStepState) => void;
  setResultState: (state: ProgressStepState) => void;
};


export default function PolicyUpdateDetails(props: PolicyStepProps) {
  const [feeRatePPM, setFeeRatePPM] = useState<number>(0);
  const [baseFee, setbaseFee] = useState<number>(0);
  const [minHTLC, setMinHTLC] = useState<number>(0);
  const [maxHTLC, setMaxHTLC] = useState<number>(0);
  const [timeLockDelta, setTimeLockDelta] = useState<number>(0);

  return (
  <ProgressTabContainer>
      <div className={styles.activeColumns}>
        <span className={styles.label}>{"Fee rate (PPM)"}</span>
        <div className={styles.input}>
          <NumberFormat
            className={styles.input}
            suffix={" ppm"}
            thousandSeparator={false}
            value={feeRatePPM}
            placeholder={"1000 ppm"}
            onValueChange={(values: NumberFormatValues) => {
              setFeeRatePPM(values.floatValue as number);
            }}
          />
        </div>
        <span className={styles.label}>{"Base fee"}</span>
        <div className={styles.input}>
          <NumberFormat
            className={styles.input}
            suffix={" sat"}
            thousandSeparator={false}
            value={baseFee}
            placeholder={"1 sat"}
            onValueChange={(values: NumberFormatValues) => {
              setbaseFee(values.floatValue as number);
            }}
          />
        </div>
        <span className={styles.label}>{"Min HTLC amount"}</span>
        <div className={styles.input}>
          <NumberFormat
            className={styles.input}
            suffix={" sat"}
            thousandSeparator={false}
            value={minHTLC}
            placeholder={"1 sat"}
            onValueChange={(values: NumberFormatValues) => {
              setMinHTLC(values.floatValue as number);
            }}
          />
        </div>
        <span className={styles.label}>{"Max HTLC amount"}</span>
        <div className={styles.input}>
          <NumberFormat
            className={styles.input}
            suffix={" sat"}
            thousandSeparator={true}
            value={maxHTLC}
            placeholder={"10000000 sat"}
            onValueChange={(values: NumberFormatValues) => {
              setMaxHTLC(values.floatValue as number);
            }}
          />
        </div>
        <span className={styles.label}>{"Time Lock Delta"}</span>
        <div className={styles.input}>
          <NumberFormat
            className={styles.input}
            thousandSeparator={false}
            value={timeLockDelta}
            placeholder={"3600"}
            onValueChange={(values: NumberFormatValues) => {
              setTimeLockDelta(values.floatValue as number);
            }}
          />
        </div>
        <ButtonWrapper
            className={styles.customButtonWrapperStyles}
            rightChildren={
              <Button
                text={"Update Channel"}
                onClick={() => {
                  props.setStepIndex(2);
                  props.setChannelState(ProgressStepState.completed);
                  props.setResultState(ProgressStepState.processing);
                  //setResponses([]);
                }}
                buttonColor={buttonColor.subtle}
              />
            }
          />
      </div>
  </ProgressTabContainer>
  );
}
