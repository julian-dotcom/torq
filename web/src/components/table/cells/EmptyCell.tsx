const classNames = require('classnames');

function EmptyCell(key?: string, index?: number | string, className?: string) {
  return (<div className={classNames("cell empty", className)} key={"last-cell-" + index} />)
}

export default EmptyCell;
