import React from 'react';
import './interval_select.scss'
import {
  CalendarLtr20Regular as IntervalIcon,
} from "@fluentui/react-icons";


function TimeIntervalSelect() {
  return (
    <div className="time-interval-wrapper">
      <div className="icon">
        <IntervalIcon/>
      </div>
      <div className="interval">
        <div className="current">3rd - 31st March</div>
        <div className="previous">1st - 28th February</div>
      </div>

    </div>
  );
}
export default TimeIntervalSelect;
