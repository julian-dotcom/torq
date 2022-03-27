const classNames = require('classnames');

function AliasCell(current: string, key: string, index: number | string, className?: string) {
  return (
    <div className={classNames("cell", "align-left", key, className)} key={key + index}>
      <div className="current">
        {current}
      </div>
      <div className="past">
        Open
      </div>
    </div>
  )
}

export default AliasCell;
