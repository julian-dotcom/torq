import { ReactElement } from "react";

interface boxProps {
  children: ReactElement;
  title?: string;
  maxWidth?: number;
}

function Box(props: boxProps) {
  const styles = {
    box: {
      backgroundColor: "var(--bg-extra-faint)",
      border: "1px solid var(--bg-subtle)",
      borderRadius: "3px",
      padding: "10px",
      marginTop: "5px"
    },
    container: {
      display: "default"
    }
  };
  if (props.maxWidth) {
    styles.container["maxWidth" as keyof typeof styles.container] =
      props.maxWidth + "px";
  }
  return (
    <div style={styles.container}>
      <span>{props.title}</span>
      <div style={styles.box}>{props.children}</div>
    </div>
  );
}

export default Box;
