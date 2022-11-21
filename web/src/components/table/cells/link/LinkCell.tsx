import styles from "components/table/cells/cell.module.scss";
import React from "react";
import classNames from "classnames";

interface LinkCell {
  current: string;
  link: string;
  className?: string;
}

function LinkCell(props: LinkCell) {
  return (
    <a
      className={classNames(styles.cell, styles.numericCell, styles.alignRight, props.className)}
      rel="noreferrer"
      target="_blank"
      href={props.link}
    >
      <div className={classNames(styles.current, styles.text, styles.link)}>{props.current}</div>
    </a>
  );
}
const LinkellMemo = React.memo(LinkCell);
export default LinkellMemo;
