import React, { useEffect } from "react";
import { Routes, Route, useLocation, useNavigate } from "react-router-dom";
import DefaultLayout from "./layout/DefaultLayout";
import LoginLayout from "./layout/LoginLayout";
import TablePage from "./features/table/TablePage";

import LoginPage from "./features/auth/LoginPage";
import SettingsPage from "./features/settings/SettingsPage";
import "./App.scss";
import { Cookies } from "react-cookie";
import { useLogoutMutation } from "apiSlice";
import Toasts, { addToastHandle } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import { BrowserRouter } from "react-router-dom";
import ChannelPage from "./features/channel/ChannelPage";

function Logout() {
  const [logout] = useLogoutMutation();
  const navigate = useNavigate();

  useEffect(() => {
    let c = new Cookies();
    c.remove("torq_session");
    logout();
    navigate("/login", { replace: true });
  });

  return <div />;
}

function App() {
  const toastRef = React.useRef<addToastHandle>();
  return (
    <ToastContext.Provider value={toastRef}>
      <BrowserRouter>
        <div className="App torq">
          <Toasts ref={toastRef} />
          <Routes>
            <Route element={<LoginLayout />}>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/logout" element={<Logout />} />
            </Route>
            <Route element={<DefaultLayout />}>
              <Route
                path="/"
                element={
                  <RequireAuth>
                    <TablePage />
                  </RequireAuth>
                }
              />
              <Route
                path="/channel/:chanId"
                element={
                  <RequireAuth>
                    <ChannelPage />
                  </RequireAuth>
                }
              />
            </Route>
            <Route element={<DefaultLayout />}>
              <Route
                path="/settings"
                element={
                  <RequireAuth>
                    <SettingsPage />
                  </RequireAuth>
                }
              />
            </Route>
          </Routes>
        </div>
      </BrowserRouter>
    </ToastContext.Provider>
  );
}

function RequireAuth({ children }: { children: JSX.Element }) {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    let c = new Cookies();
    const cookies = c.get("torq_session");
    if (cookies === undefined) {
      navigate("/login", { replace: true, state: location });
    }
  });

  return children;
}

export default App;
