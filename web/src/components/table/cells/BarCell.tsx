import './cell.scss'
import NumberFormat from 'react-number-format';
const classNames = require('classnames');

interface barCell {
  current: number,
  total: number,
  previous: number,
  index: number | string,
  className?: string
}

function BarCell({current, total, previous, index, className}: barCell) {
  let previousPercent = Math.round((current/previous)*Math.random()*200)
  const percent = current / total;
  const positive: number = Math.round(Math.random())
  return (
    <div className={classNames("cell", "bar-cell", className)}>
      <div className="current">
        <NumberFormat displayType="text" thousandSeparator=',' value={current} decimalScale={2} />
      </div>
      <div className="bar-wrapper">
        <div className="bar" style={{width: percent*100+"%"}}/>
        <div className="total-bar"/>
      </div>
      {/*<div className="bar-wrapper">*/}
      {/*  <div className="bar" style={{width: percent*100+"%"}}/>*/}
      {/*  <div className="total-bar"/>*/}
      {/*</div>*/}
      {/*<div className={"past " + (previousPercent >= 100 ? "positive" : "negative")}>*/}
      {/*  <NumberFormat displayType="text" thousandSeparator=',' value={previousPercent}  /> %*/}
      {/*</div>*/}
      {/*<div className={classNames("past", {"positive": positive, "negative": !positive})}>*/}
      {/*  {positive ? "+" : "-"}*/}
      {/*  <NumberFormat displayType="text" thousandSeparator=',' value={Math.round(previous*0.141)}  />*/}
      {/*</div>*/}
    </div>
  )
}

export default BarCell;
