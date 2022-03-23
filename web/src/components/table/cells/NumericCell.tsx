import './cell.scss'
import NumberFormat from 'react-number-format';

const classNames = require('classnames');

function NumericCell(current: number, previous: number, key: string, index: number|string, className?: string) {
  return (
    <div className={classNames("cell", "numeric-cell", key, className)} key={key + index}>
      <div className="current">
        <NumberFormat displayType="text" thousandSeparator=',' value={current} />
      </div>
      <div className="past">
        <NumberFormat displayType="text" thousandSeparator=',' value={previous} />
      </div>
    </div>
  )
}

export default NumericCell;
