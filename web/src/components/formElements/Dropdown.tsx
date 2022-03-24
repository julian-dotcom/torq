import React from 'react';
import './dropdown.scss'
import { Menu, Transition } from '@headlessui/react'
import { Fragment, useEffect, useRef, useState } from 'react'
import {
  ChevronDown16Regular as ChevronIcon,
  ReOrder16Regular as DragHandleIcon,
  Delete16Regular as DeleteIcon,
} from '@fluentui/react-icons'
import classNames from "classnames";

function Dropdown() {
    return (
    <div className="">
      <Menu as="div" className="dropdown">
        <div>
          <Menu.Button className="button">
            Top Revenue
            <ChevronIcon/>
          </Menu.Button>
        </div>
        <Transition
          as={Fragment}
          enter="transition ease-out duration-100"
          enterFrom="transform opacity-0 scale-95"
          enterTo="transform opacity-100 scale-100"
          leave="transition ease-in duration-75"
          leaveFrom="transform opacity-100 scale-100"
          leaveTo="transform opacity-0 scale-95"
        >
          <Menu.Items className="dropdown-menu-items-container ">
              <Menu.Item>
                {({active}) => (
                  <div className={classNames("dropdown-item", {active: active})}>
                    <div className="drag-icon"><DragHandleIcon/></div>
                    <div className="content">
                      <input type="text" value={"Top Revenue"}/>
                    </div>
                    <div className="trash-icon"><DeleteIcon/></div>
                  </div>
                )}
              </Menu.Item>
              <Menu.Item>
                {({active}) => (
                  <div className={classNames("dropdown-item", {active: active})}>
                    <div className="drag-icon"><DragHandleIcon/></div>
                    <div className="content">
                      <input type="text" value={"Source channels"}/>
                    </div>
                    <div className="trash-icon"><DeleteIcon/></div>
                  </div>
                )}
              </Menu.Item>
              <Menu.Item>
                {({active}) => (
                  <div className={classNames("dropdown-item", {active: active})}>
                    <div className="drag-icon"><DragHandleIcon/></div>
                    <div className="content">
                      <input type="text" value={"Destination channels"}/>
                    </div>
                    <div className="trash-icon"><DeleteIcon/></div>
                  </div>
                )}
              </Menu.Item>
          </Menu.Items>
        </Transition>
      </Menu>
    </div>
    );
}

export default Dropdown;
