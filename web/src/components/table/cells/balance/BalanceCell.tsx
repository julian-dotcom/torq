import classNames from "classnames";
import { format } from "d3";
import { ArrowStepIn16Regular, ArrowStepOut16Regular } from "@fluentui/react-icons";
import cellStyles from "components/table/cells/cell.module.scss";
import styles from "./balance_cell.module.scss";

export type BalanceCellProps = {
  remote: number;
  capacity: number;
  local: number;
  totalCell?: boolean;
  className?: string;
};

const calculateAvailableBalance = (local: number, remote: number) => {
  const calculate = local / (local + remote);
  return format(".2%")(calculate);
};

const formatAmount = (amount: number) => format(",.0f")(amount);

const BalanceCell = (props: BalanceCellProps) => {
  return (
    <div
      className={classNames(
        cellStyles.cell,
        styles.balanceCell,
        { [cellStyles.totalCell]: props.totalCell },
        props.className
      )}
    >
      <div className={classNames(styles.balances)}>
        <div className={classNames(styles.amount)}>
          <div className={classNames(styles.icon)}>
            <ArrowStepIn16Regular />
          </div>
          <div className={classNames(styles.value)}>{formatAmount(props.remote)}</div>
        </div>
        <div className={classNames(styles.amount)}>
          <div className={classNames(styles.value)}>{formatAmount(props.local)}</div>
          <div className={classNames(styles.icon)}>
            <ArrowStepOut16Regular />
          </div>
        </div>
      </div>
      <div className={classNames(styles.bar)}>
        <div
          className={classNames(styles.percentage)}
          style={{ width: calculateAvailableBalance(props.local, props.remote) }}
        />
      </div>
      <div className={classNames(styles.capacity)}>
        <div className={classNames(styles.remote)}>{calculateAvailableBalance(props.remote, props.local)}</div>
        <div className={classNames(styles.total)}>{formatAmount(props.capacity)}</div>
        <div className={classNames(styles.local)}>{calculateAvailableBalance(props.local, props.remote)}</div>
      </div>
    </div>
  );
};

export default BalanceCell;
