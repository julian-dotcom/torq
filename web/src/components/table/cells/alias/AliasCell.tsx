import styles from "components/table/cells/cell.module.scss";
import { Link } from "react-router-dom";
import React from "react";
import classNames from "classnames";

interface AliasCell {
  current: string;
  lndShortChannelId: string;
  open?: boolean;
  className?: string;
}

function OpenText(open: boolean) {
  if (open) {
    return `Open`;
  }
  return `Closed`;
}

function AliasCell({ current, lndShortChannelId, open, className }: AliasCell) {
  return (
    <Link className={classNames(styles.cell, styles.alignLeft, className)} to={"/analyse/inspect/" + lndShortChannelId}>
      <div className={classNames(styles.current, styles.text)}>{current}</div>
      {open !== undefined && (
        <div className={classNames(styles.past, { [styles.positive]: open, [styles.negative]: !open })}>
          {OpenText(open)}
        </div>
      )}
    </Link>
  );
}
const AliasCellMemo = React.memo(AliasCell);
export default AliasCellMemo;
