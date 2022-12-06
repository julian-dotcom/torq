
import { ReactComponent as TorqLogo } from "icons/torq-logo.svg";
import { Link } from "react-router-dom";
import "features/services/services_page.scss";

function ServicesPage() {
  return (
    <div className="services-page-wrapper">
      <div className="services-form-wrapper">
        <div className="logo">
          <TorqLogo />
        </div>
        Torq is bootstrapping.<br />
        <Link key="retry" to={`/`}>
          Click here to retry
        </Link>
      </div>
    </div>
  );
}

export default ServicesPage;
