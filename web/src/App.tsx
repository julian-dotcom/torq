import React from 'react';
import { Routes, Route, Link } from "react-router-dom";
import {useDispatch, useSelector} from "react-redux";

import Navigation from "./components/navigation/Navigation";
import TablePage from "./pages/TablePage";
import './App.scss';
import classNames from "classnames";

function App() {

  const dispatch = useDispatch();

  const toggleNav = () => {
    dispatch({type: 'toggleNav'})
  }

  const navHidden: boolean = useSelector((state:{navHidden:boolean}) => {return state.navHidden});
  return (
    <div className="App torq">
      <div className={classNames("main-content-wrapper", {'nav-hidden': navHidden})}>
        <div className="navigation-wrapper">
          <Navigation/>
        </div>
        <div className="page-wrapper">
          {navHidden && (<div className="dismiss-navigation-background" onClick={toggleNav}/>)}
          <Routes>
            <Route path="/" element={<TablePage/>} />
          </Routes>

        </div>
      </div>
    </div>
  );
}

export default App;
