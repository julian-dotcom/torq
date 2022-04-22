import styles from './cell.module.scss'
const classNames = require('classnames');

interface AliasCell {
  current: string,
  className?: string
}

function AliasCell({current, className}: AliasCell) {
  return (
    <div className={classNames(styles.cell, styles.alignLeft, className)}>
      <div className={styles.current}>
        {current}
      </div>
      <div className={styles.past}>
        Open
      </div>
    </div>
  )
}

export default AliasCell;
