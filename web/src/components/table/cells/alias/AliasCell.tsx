import styles from "components/table/cells/cell.module.scss";
import { Link } from "react-router-dom";
import classNames from "classnames";

interface AliasCell {
  current: string;
  channelId?: string;
  open?: boolean;
  className?: string;
}

function OpenText(open: boolean) {
  if (open) {
    return `Open`;
  }
  return `Closed`;
}

function AliasCell({ current, channelId, open, className }: AliasCell) {

  const content = (
    <>
    <div className={classNames(styles.current, styles.text)}>{current}</div>
      {open !== undefined && (
        <div className={classNames(styles.past, { [styles.positive]: open, [styles.negative]: !open })}>
          {OpenText(open)}
        </div>
      )}
    </>
  )

  if (channelId) {
    return (
      <Link className={classNames(styles.cell, styles.alignLeft, className)} to={"/analyse/inspect/" + channelId}>
      {content}
      </Link>
    )
  }

  return <div className={classNames(styles.cell, styles.alignLeft, className)}>{content}</div>
}
export default AliasCell;
