import { ReactElement } from "react";

interface boxProps {
  children: ReactElement;
  title?: string;
  minWidth?: number;
}

function Box(props: boxProps) {
  const dynamicStyles = {
    box: {
      backgroundColor: "var(--bg-default)",
      border: "1px solid var(--fg-subtle)",
      borderRadius: "3px",
      padding: "20px 10px",
      marginTop: "10px",
      width: "100%",
    },
    container: {
      width: "100%",
    },
  };
  if (props.minWidth) {
    dynamicStyles.container["minWidth" as keyof typeof dynamicStyles.container] = props.minWidth + "px";
  }
  return (
    <div style={dynamicStyles.container}>
      <span>{props.title}</span>
      <div style={dynamicStyles.box}>{props.children}</div>
    </div>
  );
}

export default Box;
