import styles from "./form.module.scss";
import React from "react";
import classNames from "classnames";

export type formProps = {
  className?: string;
  children: React.ReactNode;
  formProps?: React.FormHTMLAttributes<HTMLFormElement>;
};

function Form({ className, children, formProps }: formProps) {
  return (
    <form {...formProps} className={classNames(styles.formContainer, className, formProps?.className)}>
      {children}
    </form>
  );
}

export default Form;
