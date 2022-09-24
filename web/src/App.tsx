import React, { useEffect } from "react";
import { Routes, Route, useLocation, useNavigate } from "react-router-dom";
import DefaultLayout from "./layout/DefaultLayout";
import LoginLayout from "./layout/LoginLayout";
import ForwardsPage from "./features/forwards/ForwardsPage";

import LoginPage from "./features/auth/LoginPage";
import SettingsPage from "./features/settings/SettingsPage";
import styles from "./app.module.scss";
import { Cookies } from "react-cookie";
import { useLogoutMutation } from "apiSlice";
import Toasts, { addToastHandle } from "features/toast/Toasts";
import ToastContext from "features/toast/context";
import ChannelPage from "./features/channel/ChannelPage";
import DashboardPage from "./features/channel/DashboardPage";
import PaymentsPage from "features/transact/Payments/PaymentsPage";
import InvoicesPage from "features/transact/Invoices/InvoicesPage";
import OnChainPage from "features/transact/OnChain/OnChainPage";
import AllTxPage from "./features/transact/AllTxPage";
import NoMatch from "./features/no_match/NoMatch";
import NewPaymentModal from "features/transact/Payments/NewPaymentModal";
import { CREATE_PAYMENT } from "constants/routes";

function Logout() {
  const [logout] = useLogoutMutation();
  const navigate = useNavigate();

  useEffect(() => {
    const c = new Cookies();
    c.remove("torq_session");
    logout();
    navigate("/login", { replace: true });
  });

  return <div />;
}

function App() {
  const location = useLocation();
  const background = location.state && location.state.background;

  const toastRef = React.useRef<addToastHandle>();

  return (
    <ToastContext.Provider value={toastRef}>
        <div className={styles.app}>
          <Toasts ref={toastRef} />
          <Routes location={background || location}>
            <Route element={<LoginLayout />}>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/logout" element={<Logout />} />
            </Route>

            <Route element={<DefaultLayout />}>
              <Route
                path="/"
                element={
                  <RequireAuth>
                    <DashboardPage />
                  </RequireAuth>
                }
              >
                <Route path={CREATE_PAYMENT} element={<NewPaymentModal />} />
              </Route>
              <Route path="/analyse">
                <Route
                  path="forwards"
                  element={
                    <RequireAuth>
                      <ForwardsPage />
                    </RequireAuth>
                  }
                />
                <Route
                  path="forwards/:viewId"
                  element={
                    <RequireAuth>
                      <ForwardsPage />
                    </RequireAuth>
                  }
                />
                <Route
                  path="inspect/:chanId"
                  element={
                    <RequireAuth>
                      <ChannelPage />
                    </RequireAuth>
                  }
                />
              </Route>
              <Route path="/transactions">
                <Route
                  path="payments"
                  element={
                    <RequireAuth>
                      <PaymentsPage newPayment={false} />
                    </RequireAuth>
                  }
                />
                <Route
                  path="invoices"
                  element={
                    <RequireAuth>
                      <InvoicesPage />
                    </RequireAuth>
                  }
                />
                <Route
                  path="onchain"
                  element={
                    <RequireAuth>
                      <OnChainPage />
                    </RequireAuth>
                  }
                />
                <Route
                  path="all"
                  element={
                    <RequireAuth>
                      <AllTxPage />
                    </RequireAuth>
                  }
                />
              </Route>
              <Route
                path="/settings"
                element={
                  <RequireAuth>
                    <SettingsPage />
                  </RequireAuth>
                }
              />
              <Route path="*" element={<NoMatch />} />
            </Route>
          </Routes>
          {background && (
            <Routes>
              <Route path={CREATE_PAYMENT} element={<NewPaymentModal />} />
            </Routes>
          )}
        </div>
    </ToastContext.Provider>
  );
}

function RequireAuth({ children }: { children: JSX.Element }) {
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const c = new Cookies();
    const cookies = c.get("torq_session");
    if (cookies === undefined) {
      navigate("/login", { replace: true, state: location });
    }
  });

  return children;
}

export default App;
