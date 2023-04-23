import styles from "./form.module.scss";
import React from "react";
import classNames from "classnames";

export type formProps = {
  intercomTarget: string;
  children: React.ReactNode;
  ref?: React.RefObject<HTMLFormElement>;
} & React.FormHTMLAttributes<HTMLFormElement>;

const Form = React.forwardRef(
  ({ intercomTarget, children, ...formProps }: formProps, ref: React.LegacyRef<HTMLFormElement> | undefined) => {
    return (
      <form {...formProps} className={classNames(styles.formContainer, formProps.className)} ref={ref} data-intercom-target={intercomTarget}>
        {children}
      </form>
    );
  }
);

Form.displayName = "Form";

export default Form;
