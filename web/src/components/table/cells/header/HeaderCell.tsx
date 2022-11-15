import { LockClosed12Regular as LockIcon } from "@fluentui/react-icons";
import styles from "components/table/cells/cell.module.scss";
import React from "react";
import classNames from "classnames";

interface headerCell {
  heading: string;
  className?: string;
  locked?: boolean;
}

function HeaderCell({ heading, className, locked }: headerCell) {
  return (
    <div className={classNames(styles.header, className)}>
      <div className={styles.content}>
        {locked ? (
          <div className="icon small">
            <LockIcon />
          </div>
        ) : (
          ""
        )}
        <div className="text">{heading}</div>
      </div>
    </div>
  );
}
const HeaderCellMemo = React.memo(HeaderCell);
export default HeaderCellMemo;
