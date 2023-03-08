import DashboardPageTemplate from "features/templates/dashboardPageTemplate/DashboardPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/dashboard/dashboard-page.module.scss";
import {
  TableControlsButtonGroup,
  TableControlSection,
  TableControlsTabsGroup,
} from "../templates/tablePageTemplate/TablePageTemplate";
import Button, { ColorVariant } from "components/buttons/Button";
import mixpanel from "mixpanel-browser";
import { useNavigate } from "react-router-dom";
import * as Routes from "constants/routes";
import { useLocation } from "react-router";
import {
  MoneyHand20Regular as TransactionIcon,
  ArrowRouting20Regular as ChannelsIcon,
  Check20Regular as InvoiceIcon,
  LinkEdit20Regular as NewOnChainAddressIcon,
} from "@fluentui/react-icons";

import { NEW_ADDRESS, NEW_INVOICE, NEW_PAYMENT } from "constants/routes";
import SummaryCard from "components/summary/summaryCard/SummaryCard";

function DashboardPage() {
  const { t } = useTranslations();
  const navigate = useNavigate();
  const location = useLocation();

  const controls = (
    <TableControlSection>
      <TableControlsButtonGroup>
        <TableControlsTabsGroup>
          <Button
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<ChannelsIcon />}
            onClick={() => {
              mixpanel.track("Navigate to Open Channel");
              navigate(Routes.OPEN_CHANNEL, { state: { background: location } });
            }}
          >
            {t.openChannel}
          </Button>
          <Button
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<TransactionIcon />}
            onClick={() => {
              mixpanel.track("Navigate to New Payment");
              navigate(NEW_PAYMENT, { state: { background: location } });
            }}
          >
            {t.newPayment}
          </Button>
          <Button
            buttonColor={ColorVariant.success}
            hideMobileText={true}
            icon={<InvoiceIcon />}
            onClick={() => {
              navigate(NEW_INVOICE, { state: { background: location } });
              mixpanel.track("Navigate to New Invoice");
            }}
          >
            {t.header.newInvoice}
          </Button>

          <Button
            buttonColor={ColorVariant.success}
            icon={<NewOnChainAddressIcon />}
            hideMobileText={true}
            onClick={() => {
              navigate(NEW_ADDRESS, { state: { background: location } });
              mixpanel.track("Navigate to New OnChain Address");
            }}
          >
            {t.newAddress}
          </Button>
        </TableControlsTabsGroup>
      </TableControlsButtonGroup>
    </TableControlSection>
  );

  return (
    <DashboardPageTemplate title={t.dashboard} welcomeMessage={t.dashboardPage.welcome}>
      {controls}
      <div className={styles.dashboardWrapper}>
        <div className={styles.summaryCardContainer}>
          <SummaryCard heading={"Total balance"} value={"322.2"} valueLabel={"btc"} canInspect={true}></SummaryCard>
          <SummaryCard heading={"Total On-Chain Balance"} value={"2.21"} valueLabel={"btc"}></SummaryCard>
          <SummaryCard heading={"Total Off-Chain Balance"} value={"320.0"} valueLabel={"btc"}></SummaryCard>
          <SummaryCard heading={"Total Channel Count"} value={"1,235"} valueLabel={""}></SummaryCard>
        </div>
      </div>
    </DashboardPageTemplate>
  );
}

export default DashboardPage;
