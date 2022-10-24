import { ReactNode } from "react";
import styles from "./formRow.module.scss";

type FormRowProps = {
  children: ReactNode;
};

function FormRow(props: FormRowProps) {
  return <div className={styles.formRowWrapper}>{props.children}</div>;
}

export default FormRow;
