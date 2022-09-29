import { useEffect } from "react";
import { RouteObject, useRoutes } from "react-router";
import { useNavigate } from "react-router-dom";
import { Cookies } from "react-cookie";

import RequireAuth from "./RequireAuth";
import { useLogoutMutation } from "./apiSlice";

import LoginLayout from "./layout/LoginLayout";
import DefaultLayout from "./layout/DefaultLayout";

import LoginPage from "./features/auth/LoginPage";
import DashboardPage from "./features/channel/DashboardPage";
import ForwardsPage from "./features/forwards/ForwardsPage";
import ChannelPage from "./features/channel/ChannelPage";
import PaymentsPage from "./features/transact/Payments/PaymentsPage";
import InvoicesPage from "./features/transact/Invoices/InvoicesPage";
import OnChainPage from "./features/transact/OnChain/OnChainPage";
import AllTxPage from "./features/transact/AllTxPage";
import SettingsPage from "./features/settings/SettingsPage";
import NoMatch from "./features/no_match/NoMatch";

import * as routes from './constants/routes';

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

const publicRoutes: RouteObject = {
  element: <LoginLayout />,
  children: [
    { path: routes.LOGIN, element: <LoginPage /> },
    { path: routes.LOGOUT, element: <Logout /> }
  ]
};

const authenticatedRoutes: RouteObject = {
  element: <DefaultLayout />,
  children: [
    {
      element: <RequireAuth />,
      children: [
        { path: routes.ROOT, element: <DashboardPage /> },
        {
          path: routes.ANALYSE,
          children: [
            { path: routes.FORWARDS, element: <ForwardsPage /> },
            { path: routes.FORWARDS_CUSTOM_VIEW, element: <ForwardsPage /> },
            { path: routes.INSPECT_CHANNEL, element: <ChannelPage /> },
          ],
        },
        {
          path: routes.TRANSACTIONS,
          children: [
            { path: routes.PAYMENTS, element: <PaymentsPage newPayment={false} /> }, // TODO: remove newPayment prop after merging the new modal component
            { path: routes.INVOICES, element: <InvoicesPage /> },
            { path: routes.ONCHAIN, element: <OnChainPage /> },
            { path: routes.ALL, element: <AllTxPage /> },
          ],
        },
        { path: routes.SETTINGS, element: <SettingsPage /> },
        { path: '*', element: <NoMatch /> }
      ]
    }
  ],
};

const Router = () => {
  const router = useRoutes([
    publicRoutes,
    authenticatedRoutes,
  ]);

  return router;
}

export default Router;
