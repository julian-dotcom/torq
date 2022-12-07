
import React from "react";
import { Link } from "react-router-dom";
import { ReactComponent as TorqLogo } from "icons/torq-logo.svg";
import { useGetServicesQuery } from "apiSlice";
import { services } from "apiTypes";
import "features/services/services_page.scss";


function ServicesPage() {
  const { data: servicesData } = useGetServicesQuery();
  const [servicesState, setServicesState] = React.useState({} as services);
  const version = servicesState?.torqService ? servicesState?.torqService.version : "Unknown";

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
        Torq ({version}) is bootstrapping.<br />
        <Link key="retry" to={`/`}>
          Click here to retry
        </Link>
      </div>
    </div>
  );
}

export default ServicesPage;
