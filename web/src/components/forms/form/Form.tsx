import styles from "./form.module.scss";
import React from "react";
import classNames from "classnames";

export type formProps = {
  children: React.ReactNode;
} & React.FormHTMLAttributes<HTMLFormElement>;

function Form({ children, ...formProps }: formProps) {
  return (
    <form {...formProps} className={classNames(styles.formContainer)}>
      {children}
    </form>
  );
}

export default Form;
