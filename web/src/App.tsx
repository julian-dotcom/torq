import React, { useEffect } from "react";
import {
  Routes,
  Route,
  useLocation,
  Navigate
} from "react-router-dom";
import DefaultLayout from "./layout/DefaultLayout";
import LoginLayout from "./layout/LoginLayout";
import TablePage from "./features/table/TablePage";
import LoginPage from "./features/auth/LoginPage";
import "./App.scss";
import { Cookies } from "react-cookie";
import { useAppDispatch } from "./store/hooks";
/* import { logoutAsync } from "./features/auth/authSlice"; */
import { useLogoutMutation } from 'apiSlice'

function Logout() {
  const [logout] = useLogoutMutation()
  /* const dispatch = useAppDispatch(); */

  useEffect(() => {
    let c = new Cookies();
    c.remove("torq_session");
    /* dispatch(logoutAsync()); */
    logout()
  });

  return <Navigate to="/login" replace />;
}

function App() {

  return (
    <div className="App torq">
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
        </Route>
      </Routes>
    </div>
  );
}

function RequireAuth({ children }: { children: JSX.Element }) {
  let location = useLocation();

  let c = new Cookies();
  let torqSession = c.get("torq_session");
  if (torqSession === undefined) {
    return <Navigate to={"/login"} state={location} replace={true} />;
  }

  return children;
}

export default App;
