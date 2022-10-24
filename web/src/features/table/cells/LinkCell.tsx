import styles from "./cell.module.scss";
import React from "react";
import classNames from "classnames";

interface LinkCell {
  current: string;
  link: string;
  className?: string;
}

function LinkCell({ current, link, className }: LinkCell) {
  return (
    <a className={classNames(styles.cell, styles.numericCell, styles.alignRight, className)} rel="noreferrer" target="_blank" href={link}>
      <div className={classNames(styles.current, styles.text, styles.link)}>{current}</div>
    </a>
  );
}
const LinkellMemo = React.memo(LinkCell);
export default LinkellMemo;
