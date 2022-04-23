import styles from './cell.module.scss'
const classNames = require('classnames');

interface AliasCell {
  current: string,
  open?: number,
  className?: string
}

function OpenText(open: number) {
  if (open > 1) {
    return `Open (${open})`
  } else if (open === 1) {
    return `Open`
  } else {
    return `Closed`
  }
}

function AliasCell({current, open, className}: AliasCell) {
  return (
    <div className={classNames(styles.cell, styles.alignLeft, className)}>
      <div className={styles.current}>
        {current}
      </div>
      {open !== undefined && (
        <div className={classNames(styles.past, {[styles.positive]: open, [styles.negative]: !open})}>
          {OpenText(open)}
        </div>
      )}
    </div>
  )
}

export default AliasCell;
