import { useEffect } from "react";
import { Cookies } from "react-cookie";
import { RouteObject, useRoutes } from "react-router";
import { useLocation, useNavigate } from "react-router-dom";
import { useLogoutMutation } from "apiSlice";
import RequireTorq from "RequireTorq";
import DefaultLayout from "layout/DefaultLayout";
import LoginLayout from "layout/LoginLayout";
import LoginPage from "features/auth/LoginPage";
import CookieLoginPage from "features/auth/CookieLoginPage";
import ServicesPage from "features/services/ServicesPage";
import ChannelPage from "features/channel/ChannelPage";
import ChannelsPage from "features/channels/ChannelsPage";
import DashboardPage from "features/channel/DashboardPage";
import ForwardsPage from "features/forwards/ForwardsPage";
import NoMatch from "features/no_match/NoMatch";
import SettingsPage from "features/settings/SettingsPage";
import AllTxPage from "features/transact/AllTxPage";
import InvoicesPage from "features/transact/Invoices/InvoicesPage";
import OnChainPage from "features/transact/OnChain/OnChainPage";
import NewPaymentModal from "features/transact/NewPayment/NewPaymentModal";
import NewAddressModal from "features/transact/newAddress/NewAddressModal";
import UpdateChannelModal from "features/channels/updateChannel/UpdateChannelModal";
import OpenChannelModal from "features/channels/openChannel/OpenChannelModal";
import CloseChannelModal from "features/channels/closeChannel/CloseChannelModal";
import PaymentsPage from "features/transact/Payments/PaymentsPage";
import NewInvoiceModal from "features/transact/newInvoice/NewInvoiceModal";
import * as routes from "constants/routes";
import WorkflowPage from "pages/WorkflowPage/WorkflowPage";
import WorkflowsTablePage from "pages/WorkflowPage/WorkflowsTablePage";
import TagsPage from "pages/tags/tagsPage/TagsPage";
import TagsModal from "pages/tags/tagPage/TagsModal";
import AddTag from "pages/tags/addTagPage/AddTag";
import MessageVerificationModal from "./pages/messageVerificationPage/MessageVerificationModal";
import ClosedChannelsPage from "features/channelsClosed/ChannelsClosedPage";
import ChannelsPendingPage from "./features/channelsPending/ChannelsPendingPage";

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
    { path: routes.COOKIELOGIN, element: <CookieLoginPage /> },
    { path: routes.SERVICES, element: <ServicesPage /> },
    { path: routes.LOGOUT, element: <Logout /> },
  ],
};

const modalRoutes: RouteObject = {
  children: [
    { path: routes.NEW_INVOICE, element: <NewInvoiceModal /> },
    { path: routes.NEW_PAYMENT, element: <NewPaymentModal /> },
    { path: routes.NEW_ADDRESS, element: <NewAddressModal /> },
    { path: routes.UPDATE_CHANNEL, element: <UpdateChannelModal /> },
    { path: routes.OPEN_CHANNEL, element: <OpenChannelModal /> },
    { path: routes.CLOSE_CHANNEL, element: <CloseChannelModal /> },
    { path: routes.TAG, element: <TagsModal /> },
    { path: routes.UPDATE_TAG, element: <TagsModal /> },
    { path: routes.TAG_NODE, element: <AddTag /> },
    { path: routes.TAG_CHANNEL, element: <AddTag /> },
    { path: routes.INSPECT_CHANNEL, element: <ChannelPage /> },
    { path: routes.MESSAGE_VERIFICATION, element: <MessageVerificationModal /> },
  ],
};

const authenticatedRoutes: RouteObject = {
  element: <DefaultLayout />,
  children: [
    {
      element: <RequireTorq />,
      children: [
        {
          path: routes.ROOT,
          element: <DashboardPage />,
          children: modalRoutes.children,
        },
        {
          path: routes.CHANNELS,
          children: [
            { path: routes.OPEN_CHANNELS, element: <ChannelsPage /> },
            { path: routes.PENDING_CHANNELS, element: <ChannelsPendingPage /> },
            { path: routes.CLOSED_CHANNELS, element: <ClosedChannelsPage /> },
          ],
        },
        {
          path: routes.MANAGE,
          children: [
            { path: routes.WORKFLOWS, element: <WorkflowsTablePage /> },
            { path: routes.TAGS, element: <TagsPage /> },
            { path: routes.WORKFLOW, element: <WorkflowPage /> },
          ],
        },
        {
          path: routes.ANALYSE,
          children: [
            { path: routes.FORWARDS, element: <ForwardsPage /> },
            { path: routes.FORWARDS_CUSTOM_VIEW, element: <ForwardsPage /> },
          ],
        },
        {
          path: routes.TRANSACTIONS,
          children: [
            { path: routes.PAYMENTS, element: <PaymentsPage /> },
            { path: routes.INVOICES, element: <InvoicesPage /> },
            { path: routes.ONCHAIN, element: <OnChainPage /> },
            { path: routes.ALL, element: <AllTxPage /> },
          ],
        },
        { path: routes.SETTINGS, element: <SettingsPage /> },
        { path: "*", element: <NoMatch /> },
      ],
    },
  ],
};

const Router = () => {
  const location = useLocation();
  const background = location.state && location.state.background;
  const currentLocation = background || location;

  const routes = [publicRoutes, authenticatedRoutes];

  const router = useRoutes(routes, currentLocation);
  const modalRouter = useRoutes([modalRoutes, publicRoutes, authenticatedRoutes]);

  return (
    <>
      {router}
      {background && modalRouter}
    </>
  );
};

export default Router;
