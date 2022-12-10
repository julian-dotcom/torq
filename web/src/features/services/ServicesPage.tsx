
import React from "react";
import { ReactComponent as TorqLogo } from "icons/torq-logo.svg";
import { useGetServicesQuery } from "apiSlice";
import { services } from "apiTypes";
import "features/services/services_page.scss";
import Button, { buttonColor } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import {useLocation, useNavigate} from "react-router-dom";


function ServicesPage() {
  const { t } = useTranslations();
  const { data: servicesData, refetch: getServices } = useGetServicesQuery();
  const [servicesState, setServicesState] = React.useState({} as services);
  const version = servicesState?.torqService ? servicesState?.torqService.version : "Unknown";

  const navigate = useNavigate();
  const location = useLocation();

  const retryServices = () => {
    getServices();
    if (servicesState.torqService.status == 1) {
      navigate("/", { replace: true, state: location });
    }
  };

  React.useEffect(() => {
    if (servicesData) {
      setServicesState(servicesData)
    }
  }, [servicesData]);

  return (
    <div className="services-page-wrapper">
      <div className="services-form-wrapper">
        <div className="logo">
          <TorqLogo />
        </div>
        Torq ({version}): {t.bootstrapping}<br />
        <Button
          buttonColor={buttonColor.primary}
          onClick={retryServices}
          text={t.retry}
        />
      </div>
    </div>
  );
}

export default ServicesPage;
