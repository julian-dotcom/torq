import DetailsPageTemplate from "../templates/detailsPageTemplate/DetailsPageTemplate";
import useTranslations from "services/i18n/useTranslations";
import styles from "features/dashboard/dashboard-page.module.scss";

function DashboardPage() {
  const { t } = useTranslations();

  return (
    <DetailsPageTemplate title={t.dashboard}>
      <div className={styles.dashboardWrapper}>
        <p>dsfsdfs</p>
      </div>
    </DetailsPageTemplate>
  );
}

export default DashboardPage;
