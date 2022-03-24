import './cell.scss'
import NumberFormat from 'react-number-format';

const classNames = require('classnames');

function BarCell(current: number, total: number, previous: number, key: string, index: number|string, className?: string) {
  let previousPercent = previous*Math.round(Math.random()*200)
  return (
    <div className={classNames("cell", "bar-cell", key, className)} key={key + index}>
      <div className="current">
        <NumberFormat displayType="text" thousandSeparator=',' value={current} />
      </div>
      <div className="bar-wrapper">
        <div className="bar" style={{width: Math.random()*100+"%"}}/>
        <div className="total-bar"/>
      </div>
      <div className={"past " + (previousPercent >= 100 ? "positive" : "negative")}>
        <NumberFormat displayType="text" thousandSeparator=',' value={previousPercent}  /> %
      </div>
    </div>
  )
}

export default BarCell;
