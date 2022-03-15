import React from 'react';
//import logo from './logo.svg';
import { Routes, Route, Link } from "react-router-dom";
import Navigation from "./components/navigation/Navigation";
import TableControls from "./components/table/TableControls";
import './App.scss';
import TablePage from "./pages/TablePage";

function App() {
  return (
    <div className="App torq">
      <div className="main-content-wrapper">
        <div className="navigation-wrapper">
          <Navigation/>
        </div>
        <div className="page-wrapper">
          <Routes>
            <Route path="/" element={<TablePage/>} />
          </Routes>
        </div>
      </div>
    </div>
  );
}

export default App;
