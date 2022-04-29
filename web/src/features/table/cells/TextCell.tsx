import styles from './cell.module.scss'
const classNames = require('classnames');

interface textCell {
  current: string | Array<string>,
  className?: string
}

function TextCell({current, className}: textCell) {
  let displayCurrent = current
  if (Array.isArray(current)) {
    displayCurrent = current.join(", ")
  }
  return (
    <div className={classNames(styles.cell, styles.alignLeft, className)}>
      <div className={styles.current}>
        {displayCurrent}
      </div>
    </div>
  )
}

export default TextCell;
