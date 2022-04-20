import styles from './cell.module.scss'

const classNames = require('classnames');

function EmptyCell(key?: string, index?: number | string, className?: string) {
  return (<div className={classNames(styles.cell, styles.empty, className)} key={"last-cell-" + index} />)
}

export default EmptyCell;
