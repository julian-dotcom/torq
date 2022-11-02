import React from "react";
import { ReactComponent as TorqLogo } from "icons/torq-logo.svg";
import { LockOpen20Regular as UnlockIcon } from "@fluentui/react-icons";
import "./login_page.scss";
import { useLocation, useNavigate } from "react-router-dom";
import { useLoginMutation } from "apiSlice";
import ToastContext from "features/toast/context";
import { toastCategory } from "features/toast/Toasts";
import type { LoginResponse } from "types/api";

function LoginPage() {
  const [login] = useLoginMutation();

  const navigate = useNavigate();
  const location = useLocation();
  interface LocationState {
    from: {
      pathname: string;
    };
  }
  const toastRef = React.useContext(ToastContext);
  let from = (location.state as LocationState)?.from?.pathname || "/";
  // Don't redirect back to logout.
  if (from === "/logout" || from === "/login" || from === "logout" || from === "login" || from === "" || from === "/") {
    from = "/";
  }

  const submit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    formData.append("username", "admin");
    const res = (await login(formData)) as LoginResponse;
    if (res?.error) {
      const errorMessage = res.error?.data?.error ? "Incorrect Password!" : "Api not reacheable!";
      toastRef?.current?.addToast(errorMessage, toastCategory.error);
    } else {
      navigate(from, { replace: true });
    }
  };

  // TODO: unify the styling here once standardised button styles are done.
  return (
    <div className="login-page-wrapper">
      <div className="login-form-wrapper">
        <div className="logo">
          <TorqLogo />
        </div>
        <form className="login-form" onSubmit={submit}>
          <input type="password" name={"password"} className={"password-field"} placeholder={"Password..."} />
          <button type="submit" className={"submit-button"}>
            <UnlockIcon />
            Login
          </button>
        </form>
      </div>
    </div>
  );
}

export default LoginPage;
