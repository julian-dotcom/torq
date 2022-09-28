import React, { useEffect } from "react";
import { BrowserRouter } from "react-router-dom";

import styles from "./app.module.scss";
import Router from "./Router";
import Toasts, { addToastHandle } from "./features/toast/Toasts";
import ToastContext from "./features/toast/context";

function App() {
  const [locationState, setLocationState] = React.useState("");

  useEffect(() => {
    const splitLocation = window.location.pathname.split("/");
    if (splitLocation.length > 1) {
      const path = splitLocation[1];
      if (path === "torq") {
        setLocationState(path);
      }
    }
  });

  const toastRef = React.useRef<addToastHandle>();

  return (
    <ToastContext.Provider value={toastRef}>
      <BrowserRouter basename={locationState}>
        <div className={styles.app}>
          <Toasts ref={toastRef} />
          <Router />
        </div>
      </BrowserRouter>
    </ToastContext.Provider>
  );
}

export default App;
