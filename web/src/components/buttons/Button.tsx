import React, {MouseEventHandler} from 'react';
import './button.scss'
import classNames from "classnames";

function DefaultButton(props: {
  text: string,
  icon?: any,
  onClick?: MouseEventHandler<HTMLDivElement> | undefined,
  className?: string,
}) {
    return (
      <div className={classNames("button", props.className)} onClick={props.onClick}>
        {props.icon && (<div className="icon">{props.icon}</div>)}
        <div className="text">{props.text}</div>
      </div>
    );
}

export default DefaultButton;
