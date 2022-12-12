import styles from "components/table/cells/cell.module.scss";
import React from "react";
import classNames from "classnames";

interface LinkCell {
  text: string;
  link: string;
  className?: string;
  totalCell?: boolean;
}

function LinkCell(props: LinkCell) {
  return (
    <a
      className={classNames(styles.cell, styles.numericCell, styles.alignRight, props.className)}
      rel="noreferrer"
      target="_blank"
      href={props.link}
    >
      {props.totalCell && <div className={classNames(styles.current, styles.text, styles.link)}>{props.text}</div>}
    </a>
  );
}
const LinkellMemo = React.memo(LinkCell);
export default LinkellMemo;
