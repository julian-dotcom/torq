import React from "react";
import { Outlet, useLocation, useNavigate } from "react-router-dom";
import { useGetServicesQuery } from "apiSlice";

const RequireTorq = () => {
  const { data: servicesData } = useGetServicesQuery();
  const navigate = useNavigate();
  const location = useLocation();

  React.useEffect(() => {
    if (servicesData) {
      if (
        servicesData.mainService === undefined ||
        servicesData.mainService.bootTime === undefined ||
        servicesData.mainService.bootTime == ""
      ) {
        navigate("/services", { replace: true, state: location });
      }
    }
  }, [servicesData]);

  return <Outlet />;
};

export default RequireTorq;
