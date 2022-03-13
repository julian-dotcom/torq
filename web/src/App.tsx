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

function Home() {
  return (
    <>
      <main>
        <h2>Welcome to the homepage!</h2>
        <p>You can do this, I believe in you.</p>
      </main>
      <nav>
        <Link to="/about">About</Link>
      </nav>
    </>
  );
}

function About() {
  return (
    <>
      <main>
        <h2>Who are we?</h2>
        <p>
          That feels like an existential question, don't you
          think?
        </p>
      </main>
      <nav>
        <Link to="/">Home</Link>
      </nav>
    </>
  );
}

export default App;
