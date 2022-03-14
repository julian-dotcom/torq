import './cell.scss'
import NumberFormat from 'react-number-format';

function NumericCell(current: number, previous: number, key: string, index: number) {
  return (
    <div className={"cell " + key} key={key + index}>
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
