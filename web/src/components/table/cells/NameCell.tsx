function NameCell(current: string, key: string, index: number) {
  return (
    // <div className={"header "} key={item.key}>
    //   <div className="top">{item.primaryHeading}</div>
    //   <div className="bottom">{item.secondaryHeading}</div>
    // </div>
    <div className={"cell align-left " + key} key={key + index}>
      <div className="current">
        {current}
      </div>
      <div className="past">
        Open
      </div>
    </div>
  )
}

export default NameCell;
