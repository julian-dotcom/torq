import react from "react";
import { useCookieLoginMutation } from "apiSlice";
import { useSearchParams, useNavigate } from "react-router-dom";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";

const CookieLoginPage = (_: unknown) => {
  const [searchParams, __] = useSearchParams();
  const navigate = useNavigate();
  const accessKey = searchParams.get("access-key");
  const [login, { error }] = useCookieLoginMutation();

  react.useEffect(() => {
    if (accessKey) {
      login(accessKey ?? "")
        .unwrap()
        .then((_) => navigate("/"));
    }
  }, [accessKey]);

  if (error) {
    return (
      <div>
        {(error as FetchBaseQueryError).status} {JSON.stringify((error as FetchBaseQueryError).data)}
      </div>
    );
  }

  return <div></div>;
};

export default CookieLoginPage;
