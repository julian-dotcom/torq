import React from 'react';
import './menu_item.scss'


function MenuItem(props: {
  text: string,
  selected?: boolean,
  icon?: any,
  children?: any,
  actions?: any,
  routeTo?: string,
}) {

  let TitleComponent = function(routeTo?: string) {
    if (routeTo) {
      return (<a href={""}  className={props.icon ? "title" : "title no-icon"} >
        <div className="icon">{props.icon}</div>
        <div className="text">{props.text}</div>
      </a>)
    } else {
      return (<div  className={props.icon ? "title" : "title no-icon"} >
        <div className="icon">{props.icon}</div>
        <div className="text">{props.text}</div>
      </div>)
    }

  }

  return (
      <div className={"item " +  (props.selected ? "selected" : "")}>

          <div className="content-wrapper">

            {TitleComponent(props.routeTo)}

            {props.actions && (<div className="actions">
              <div className="icon action">{props.actions}</div>
            </div>)}

          </div>

        {props.children && (<div className="item sub">{props.children}</div>)}
      </div>
    );
}

export default MenuItem;
