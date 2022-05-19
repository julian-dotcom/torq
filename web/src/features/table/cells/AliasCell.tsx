import styles from "./cell.module.scss";
import { Link } from "react-router-dom";
const classNames = require("classnames");

interface AliasCell {
  current: string;
  chanId: string;
  open?: number;
  className?: string;
}

function OpenText(open: number) {
  if (open > 1) {
    return `Open (${open})`;
  } else if (open === 1) {
    return `Open`;
  } else {
    return `Closed`;
  }
}

function AliasCell({ current, chanId, open, className }: AliasCell) {
  return (
    <Link className={classNames(styles.cell, styles.alignLeft, className)} to={"/channel/" + chanId}>
      <div className={classNames(styles.current, styles.text)}>{current}</div>
      {open !== undefined && (
        <div className={classNames(styles.past, { [styles.positive]: open, [styles.negative]: !open })}>
          {OpenText(open)}
        </div>
      )}
    </Link>
  );
}

export default AliasCell;
