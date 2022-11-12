import classNames from "classnames";
import { format } from "d3";
import { ArrowStepIn16Regular, ArrowStepOut16Regular } from "@fluentui/react-icons";
import styles from "./cell.module.scss";

type Props = {
  capacity?: number;
  remote?: number;
  local?: number;
  className?: string;
}

const calculateAvailableBalance = (amount: number, capacity: number) => {
  const calculate = (amount / capacity)
  const formatted = format(".2%")(calculate)

  return formatted;
}

const formatAmount = (amount: number) => format(",.0f")(amount);

const BalanceCell = ({ capacity = 100000, remote = 90678, local = 9222, className }: Props) => {


  return (
    <div className={classNames(styles.cell, styles.balanceCell, className)}>
      <div className={classNames(styles.balances)}>
        <div className={classNames(styles.amount)}>
          <div className={classNames(styles.icon)}>
            <ArrowStepIn16Regular />
          </div>
          <div className={classNames(styles.value)}>{formatAmount(remote)}</div>
        </div>
        <div className={classNames(styles.amount)}>
          <div className={classNames(styles.value)}>{formatAmount(local)}</div>
          <div className={classNames(styles.icon)}>
            <ArrowStepOut16Regular />
          </div>
        </div>
      </div>
      <div className={classNames(styles.bar)}>
        <div className={classNames(styles.percentage)} style={{ width: calculateAvailableBalance(local, capacity) }} />
      </div>
      <div className={classNames(styles.capacity)}>
        <div className={classNames(styles.remote)}>{calculateAvailableBalance(remote, capacity)}</div>
        <div className={classNames(styles.total)}>{formatAmount(capacity)}</div>
        <div className={classNames(styles.local)}>{calculateAvailableBalance(local, capacity)}</div>
      </div>
    </div>
  )
}

export default BalanceCell;