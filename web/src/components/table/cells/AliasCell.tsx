const classNames = require('classnames');

interface AliasCell {
  current: string,
  className?: string
}

function AliasCell({current, className}: AliasCell) {
  return (
    <div className={classNames("cell", "align-left", className)}>
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
