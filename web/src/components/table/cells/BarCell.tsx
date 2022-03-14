import './cell.scss'
import NumberFormat from 'react-number-format';

function BarCell(current: number, total: number, previous: number, key: string, index: number) {
  let previousPercent = previous*Math.round(Math.random()*200)
  return (
    <div className={"cell bar-cell " + key} key={key + index}>
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
