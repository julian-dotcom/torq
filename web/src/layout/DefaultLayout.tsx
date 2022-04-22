import React from 'react';
import { Outlet } from "react-router-dom";
import { useAppSelector, useAppDispatch } from '../store/hooks';
import {selectHidden, toggleNav} from '../features/navigation/navSlice'

import Navigation from "../features/navigation/Navigation";
import classNames from "classnames";

function DefaultLayout() {
  const hidden = useAppSelector(selectHidden);
  const dispatch = useAppDispatch();
  return (
    <div className={classNames("main-content-wrapper", {'nav-hidden': hidden})}>
      <div className="navigation-wrapper">
        <Navigation/>
      </div>
      <div className="page-wrapper">
        <div className="dismiss-navigation-background" onClick={()=> dispatch(toggleNav()) }/>
        <Outlet />
      </div>
    </div>
  )
}

export default DefaultLayout;
