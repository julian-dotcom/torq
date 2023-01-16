import styles from "components/table/cells/cell.module.scss";
import React from "react";
import classNames from "classnames";
import { Link } from "react-router-dom";
import mixpanel from "mixpanel-browser";

interface LinkCell {
  text: string;
  link: string;
  className?: string;
  totalCell?: boolean;
}

function LinkCell(props: LinkCell) {
  return (
    <Link
      className={classNames(styles.cell, styles.numericCell, styles.alignRight, props.className)}
      rel="noreferrer"
      target="_blank"
      to={props.link}
      onClick={() => {
        mixpanel.track("Link Cell Clicked", { href: props.link });
      }}
    >
      {props.totalCell && <div className={classNames(styles.current, styles.text, styles.link)}>{props.text}</div>}
    </Link>
  );
}
const LinkellMemo = React.memo(LinkCell);
export default LinkellMemo;
