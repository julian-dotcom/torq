import {
  LockClosed12Regular as LockIcon
} from "@fluentui/react-icons";

const classNames = require('classnames');

function HeaderCell(heading: string, key: string, className?: string, locked?: boolean) {
  return (
    <div className={classNames("header", key, className)} key={key}>
      <div className="content">
        {locked ? <div className="icon"><LockIcon/></div> : ""}
        <div className="text">{heading}</div>
      </div>
    </div>
  )
}

export default HeaderCell;
