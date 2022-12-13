import styles from "./largeAmountInput.module.scss";
import NumberFormat from "react-number-format";

type LargeAmountInputProps = {
  label?: string;
  value?: number;
  autoFocus?: boolean;
  onChange: (value: number | undefined) => void;
};

export default function LargeAmountInput(props: LargeAmountInputProps) {
  return (
    <div className={styles.amountWrapper}>
      <div className={styles.amountLabel}>{props.label}</div>
      <div className={styles.amount}>
        <NumberFormat
          className={styles.amountInput}
          datatype={"number"}
          value={props.value}
          autoFocus={props.autoFocus}
          placeholder={"0 sat"}
          onValueChange={(values) => props.onChange(parseInt(values.value))}
          thousandSeparator=","
          suffix={" sat"}
        />
      </div>
    </div>
  );
}
