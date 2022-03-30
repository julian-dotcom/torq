import React from 'react';
import {ReactComponent as TorqLogo} from '../../icons/torq-logo.svg'
import {
  LockOpen20Regular as UnlockIcon,
} from "@fluentui/react-icons";
import './login_page.scss'

function LoginPage() {
  // TODO: unify the styling here once standardised button styles are done.
  return (
    <div className="login-page-wrapper">
      <div className="login-form-wrapper">
        <div className="logo">
          <TorqLogo/>
        </div>
        <form className="login-form">
          <input type="password"
                 name={"password"}
                 className={"password-field"}
                 placeholder={"Password..."}
          />
          <button type="submit"
                 className={"submit-button"}>
            <UnlockIcon/>
            Login
          </button>
        </form>
      </div>
    </div>
  );
}

export default LoginPage;
