import styles from './cell.module.scss'
import NumberFormat from 'react-number-format';
const classNames = require('classnames');

interface numericCell {
  current: number,
  previous?: number,
  index: number | string,
  className?: string
}

function NumericCell({current, previous, index, className}: numericCell) {
  return (
    <div className={classNames(styles.cell, styles.numericCell, className, index)}>
      <div className={styles.current}>
        <NumberFormat displayType="text" thousandSeparator=',' value={current}  decimalScale={2} />
      </div>
      {/*<div className="past">*/}
      {/*  <NumberFormat displayType="text" thousandSeparator=',' value={previous} />*/}
      {/*</div>*/}
    </div>
  )
}

export default NumericCell;
