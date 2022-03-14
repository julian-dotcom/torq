import React from 'react';
import './button.scss'

function DefaultButton(props: {
  text: string,
  icon?: any,
}) {
    return (
      <div className="button">
        <div className="icon">{props.icon}</div>
        <div className="text">{props.text}</div>
      </div>
    );
}

export default DefaultButton;
