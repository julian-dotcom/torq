import { ReactNode } from "react";
import styles from "./formRow.module.scss";
import classNames from "classnames";

type FormRowProps = {
  children: ReactNode;
  className?: string;
};

function FormRow(props: FormRowProps) {
  return <div className={classNames(styles.formRowWrapper,props.className)}>{props.children}</div>;
}

export default FormRow;
