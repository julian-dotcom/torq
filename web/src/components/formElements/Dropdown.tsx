import React from 'react';
import './dropdown.scss'
// import { Menu, Transition } from '@headlessui/react'
// import { Fragment, useEffect, useRef, useState } from 'react'
import {
  ChevronDown16Regular as ChevronIcon,
  ReOrder16Regular as DragHandleIcon,
  Delete16Regular as DeleteIcon,
} from '@fluentui/react-icons'
import classNames from "classnames";

function Dropdown() {
    return (
    <div className="">
      {"Top Revenue"}
      {/*<Menu as="div" className="dropdown">*/}
      {/*  <div>*/}
      {/*    <Menu.Button className="button">*/}
      {/*      Top Revenue*/}
      {/*      <ChevronIcon/>*/}
      {/*    </Menu.Button>*/}
      {/*  </div>*/}
      {/*  /!*<Menu.Items className="dropdown-menu-items-container ">*!/*/}
      {/*  /!*    <Menu.Item>*!/*/}
      {/*  /!*      {({active: any}) => (*!/*/}
      {/*  /!*        <div className={classNames("dropdown-item", {active: active})}>*!/*/}
      {/*  /!*          <div className="drag-icon"><DragHandleIcon/></div>*!/*/}
      {/*  /!*          <div className="content">*!/*/}
      {/*  /!*            {"Top Revenue"}*!/*/}
      {/*  /!*          </div>*!/*/}
      {/*  /!*          <div className="trash-icon"><DeleteIcon/></div>*!/*/}
      {/*  /!*        </div>*!/*/}
      {/*  /!*      )}*!/*/}
      {/*  /!*    </Menu.Item>*!/*/}
      {/*  /!*    <Menu.Item>*!/*/}
      {/*  /!*      {({active: any}) => (*!/*/}
      {/*  /!*        <div className={classNames("dropdown-item", {active: active})}>*!/*/}
      {/*  /!*          <div className="drag-icon"><DragHandleIcon/></div>*!/*/}
      {/*  /!*          <div className="content">*!/*/}
      {/*  /!*            {"Source channels"}*!/*/}
      {/*  /!*          </div>*!/*/}
      {/*  /!*          <div className="trash-icon"><DeleteIcon/></div>*!/*/}
      {/*  /!*        </div>*!/*/}
      {/*  /!*      )}*!/*/}
      {/*  /!*    </Menu.Item>*!/*/}
      {/*  /!*    <Menu.Item>*!/*/}
      {/*  /!*      {({active: any}) => (*!/*/}
      {/*  /!*        <div className={classNames("dropdown-item", {active: active})}>*!/*/}
      {/*  /!*          <div className="drag-icon"><DragHandleIcon/></div>*!/*/}
      {/*  /!*          <div className="content">*!/*/}
      {/*  /!*            {"Destination channels"}*!/*/}
      {/*  /!*          </div>*!/*/}
      {/*  /!*          <div className="trash-icon"><DeleteIcon/></div>*!/*/}
      {/*  /!*        </div>*!/*/}
      {/*  /!*      )}*!/*/}
      {/*  /!*    </Menu.Item>*!/*/}
      {/*  </Menu.Items>*/}
      {/*</Menu>*/}
    </div>
    );
}

export default Dropdown;
