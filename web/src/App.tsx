import React from 'react';
import { Routes, Route, Link } from "react-router-dom";
import { useAppSelector, useAppDispatch } from './store/hooks';
import {selectHidden, toggleNav, NavState} from './components/navigation/navSlice'

import Navigation from "./components/navigation/Navigation";
import TablePage from "./pages/TablePage";
import './App.scss';
import classNames from "classnames";

function App() {

  const hidden = useAppSelector(selectHidden);
  const dispatch = useAppDispatch();

  return (
    <div className="App torq">
      <div className={classNames("main-content-wrapper", {'nav-hidden': hidden})}>
        <div className="navigation-wrapper">
          <Navigation/>
        </div>
        <div className="page-wrapper">
          <div className="dismiss-navigation-background" onClick={()=> dispatch(toggleNav()) }/>
          <Routes>
            <Route path="/" element={<TablePage/>}/>
          </Routes>
        </div>
      </div>
    </div>
  );
}

export default App;
