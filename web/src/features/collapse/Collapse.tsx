import { ReactElement, useLayoutEffect, useRef, useState } from "react";
export default (props: { children: ReactElement; collapsed: boolean }) => {
  const ref = useRef<HTMLDivElement>(null);
  const [styleState, setStyleState] = useState({ height: "auto", transition: "0.5s", overflow: "hidden" });
  useLayoutEffect(() => {
    if (ref.current) {
      setStyleState({ ...styleState, height: props.collapsed ? "0" : ref.current.scrollHeight + "px" });
    }
  }, [props]);
  return (
    <div ref={ref} style={styleState}>
      {props.children}
    </div>
  );
};
