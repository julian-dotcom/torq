import styles from "./form.module.scss";
import React from "react";
import classNames from "classnames";

export type formProps = {
  children: React.ReactNode;
  ref?: React.RefObject<HTMLFormElement>;
} & React.FormHTMLAttributes<HTMLFormElement>;

const Form = React.forwardRef(
  ({ children, ...formProps }: formProps, ref: React.LegacyRef<HTMLFormElement> | undefined) => {
    return (
      <form {...formProps} className={classNames(styles.formContainer, formProps.className)} ref={ref}>
        {children}
      </form>
    );
  }
);

Form.displayName = "Form";

export default Form;
