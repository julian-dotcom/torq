import React, { useEffect } from "react";
import { ReactComponent as TorqLogo } from "icons/torq-logo.svg";
import { useGetServicesQuery } from "apiSlice";
import { services } from "apiTypes";
import "features/services/services_page.scss";
import Button, { ColorVariant } from "components/buttons/Button";
import useTranslations from "services/i18n/useTranslations";
import { useNavigate } from "react-router-dom";

function ServicesPage() {
  const { t } = useTranslations();
  const { data: servicesData, refetch: getServices } = useGetServicesQuery();
  const [servicesState, setServicesState] = React.useState({} as services);
  const version = servicesState?.torqService ? servicesState?.torqService.version : "Unknown";

  const navigate = useNavigate();

  const retryServices = () => {
    getServices();
    if (servicesState.torqService.status == 1) {
      navigate("/");
    }
  };

  useEffect(() => {
    if (servicesData) {
      setServicesState(servicesData);
    }
  }, [servicesData]);

  return (
    <div className="services-page-wrapper">
      <div className="services-form-wrapper">
        <div className="logo">
          <TorqLogo />
        </div>
        Torq ({version}): {t.bootstrapping}
        <br />
        <Button buttonColor={ColorVariant.primary} onClick={retryServices}>
          {t.retry}
        </Button>
      </div>
    </div>
  );
}

export default ServicesPage;
