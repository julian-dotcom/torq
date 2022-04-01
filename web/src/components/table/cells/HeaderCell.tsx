import {
  LockClosed12Regular as LockIcon
} from "@fluentui/react-icons";

const classNames = require('classnames');

interface headerCell {
  heading: string;
  className?: string;
  locked?: boolean;
}

function HeaderCell({heading, className, locked}: headerCell) {
  return (
    <div className={classNames("header", className)}>
      <div className="content">
        {locked ? <div className="icon small"><LockIcon/></div> : ""}
        <div className="text">{heading}</div>
      </div>
    </div>
  )
}

export default HeaderCell;
