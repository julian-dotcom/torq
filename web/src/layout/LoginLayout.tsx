import React from 'react';
import { Outlet } from "react-router-dom";
import './login_layout.scss'
// import { useAppSelector, useAppDispatch } from '../../store/hooks';
// import {selectHidden, toggleNav} from '../../components/navigation/navSlice'

import classNames from "classnames";

function LoginLayout() {
  // const hidden = useAppSelector(selectHidden);
  // const dispatch = useAppDispatch();
  return (
    <div className={classNames("login-layout")}>
      <Outlet />
    </div>
  )
}

export default LoginLayout;
