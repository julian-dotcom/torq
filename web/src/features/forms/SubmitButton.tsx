import { ReactElement } from "react";
import style from "./submitButton.module.scss";

interface SubmitButtonProps {
  children: ReactElement;
}

function SubmitButton(props: SubmitButtonProps) {
  const styles = {};
  return (
    <button type="submit" style={styles} className={style.button}>
      {props.children}
    </button>
  );
}

export default SubmitButton;
