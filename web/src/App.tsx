import React, { useEffect } from "react";
import {
  Routes,
  Route,
  useNavigate,
  useLocation,
  Navigate
} from "react-router-dom";
import DefaultLayout from "./layout/DefaultLayout";
import LoginLayout from "./layout/LoginLayout";
import TablePage from "./components/table/TablePage";
import LoginPage from "./components/auth/LoginPage";
import "./App.scss";
import { Cookies, useCookies } from "react-cookie";
import { useAppDispatch } from "./store/hooks";
import { logoutAsync } from "./components/auth/authSlice";
import {fetchTableViewsAsync} from "./components/table/tableSlice";

function Logout() {
  const dispatch = useAppDispatch();

  useEffect(() => {
    let c = new Cookies();
    c.remove("torq_session");
    dispatch(logoutAsync());
  });

  return <Navigate to="/login" replace />;
}

function App() {

  const dispatch = useAppDispatch();

  useEffect(() =>{
    dispatch(fetchTableViewsAsync());
  })

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
  if (torqSession == undefined) {
    return <Navigate to={"/login"} state={location} replace={true} />;
  }

  return children;
}

export default App;
