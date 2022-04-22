import styles from './cell.module.scss'
const classNames = require('classnames');

interface textCell {
  current: string,
  className?: string
}

function TextCell({current, className}: textCell) {
  return (
    <div className={classNames(styles.cell, styles.alignLeft, className)}>
      <div className={styles.current}>
        {current}
      </div>
    </div>
  )
}

export default TextCell;
