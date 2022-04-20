import {
  LockClosed12Regular as LockIcon
} from "@fluentui/react-icons";
import {inspect} from "util";
import styles from './cell.module.scss'

const classNames = require('classnames');

interface headerCell {
  heading: string;
  className?: string;
  locked?: boolean;
}

function HeaderCell({heading, className, locked}: headerCell) {
  return (
    <div className={classNames(styles.header, className)}>
      <div className={styles.content}>
        {locked ? <div className="icon small"><LockIcon/></div> : ""}
        <div className="text">{heading}</div>
      </div>
    </div>
  )
}

export default HeaderCell;
