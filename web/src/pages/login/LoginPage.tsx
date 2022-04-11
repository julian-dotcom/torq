import React, { useEffect } from 'react';
import { useAppDispatch } from '../../store/hooks';
import { loginAsync } from './authSlice';
import { ReactComponent as TorqLogo } from '../../icons/torq-logo.svg'
import {
  LockOpen20Regular as UnlockIcon,
} from "@fluentui/react-icons";
import './login_page.scss'
import { useLocation, useNavigate } from "react-router-dom";
import { Cookies } from "react-cookie";

function LoginPage() {

  const navigate = useNavigate();
  const location = useLocation();
  interface LocationState {
    from: {
      pathname: string;
    };
  }

  let from = (location.state as LocationState)?.from?.pathname || "/"
  // Don't redirect back to logout.
  if (from == "/logout" || "/login") {
    from = "/"
  }

  let c = new Cookies
  const cookies = c.get('torq_session');
  if (cookies !== undefined) {
    navigate(from, { replace: true });
  }
  const dispatch = useAppDispatch()

  const submit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const password = formData.get("password") as string;

    dispatch(loginAsync({ password }))

    navigate(from, { replace: true });
  }

  // TODO: unify the styling here once standardised button styles are done.
  return (
    <div className="login-page-wrapper">
      <div className="login-form-wrapper">
        <div className="logo">
          <TorqLogo />
        </div>
        <form className="login-form" onSubmit={submit}>
          <input type="password"
            name={"password"}
            className={"password-field"}
            placeholder={"Password..."}
          />
          <button type="submit"
            className={"submit-button"}>
            <UnlockIcon />
            Login
          </button>
        </form>
      </div>
    </div>
  );
}

export default LoginPage;
