import React from "react";
import ReactDOM from "react-dom";
import reportWebVitals from "./reportWebVitals";
import { Provider } from "react-redux";
import App from "./App";
import { store } from "./store/store";

import "./styles/main.scss";
import { BrowserRouter } from "react-router-dom";

const resolveBasename = () => {
  const path = window.location.pathname.split('/');
  const topLevelPath = path.length > 1 && path[1];

  return topLevelPath === 'torq' ? topLevelPath : '';
}

const basename = resolveBasename();

ReactDOM.render(
  <BrowserRouter basename={basename}>
    <React.StrictMode>
      <Provider store={store}>
        <App />
      </Provider>
    </React.StrictMode>
  </BrowserRouter>,
  document.getElementById("root")
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
