import { ReactElement, useLayoutEffect, useRef, useState } from "react";

function useDefaultStyles(defaultCollapsed: boolean, animate: boolean) {
  let initialStyles = {
    overflow: "hidden",
    transition: animate ? "height 250ms ease-in-out" : "none",
    height: "auto",
  };
  if (defaultCollapsed) {
    initialStyles = { ...initialStyles, height: "0" };
  }
  return initialStyles;
}

export default function Collapse(props: {
  children: ReactElement;
  collapsed: boolean;
  animate: boolean;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);

  const initialStyles = useDefaultStyles(props.collapsed, props.animate);

  const [styleState, setStyleState] = useState(initialStyles);

  useLayoutEffect(() => {
    if (!ref.current) {
      // setStyleState({ ...styleState, height: props.collapsed ? "0" : ref.current.scrollHeight + "px" });
      return;
    }
    // The bounding box of the child element will always be there even if the parrent height is 0px.
    // This allows us to use the currenty bounding box to set the height of the parrent.
    const bb = ref.current.getBoundingClientRect();

    if (!props.collapsed) {
      // Expand the body content by setting the body wrapper to the current height of the body.
      setStyleState({ ...styleState, height: ref.current.scrollHeight + "px", overflow: "hidden" });

      // Wait before applying the overflow property to avoid content becoming visible too early,
      // but still allowing dropdowns etc. to show. We also want to set the height to auto, so that
      // the content can expand to fit the available space.
      setTimeout(() => {
        setStyleState({ ...styleState, height: "auto", overflow: "visible" });
      }, 250);
    } else {
      // Since we set the height to auto, we need to set it to the current height and wait 1ms before we can transition
      // to 0px. If not, then the transition will not work and the box size jumps streight to 0.
      setStyleState({ ...styleState, height: ref.current.scrollHeight + "px", overflow: "hidden" });
      setTimeout(() => {
        setStyleState({ ...styleState, height: "0px", overflow: "hidden" });
      }, 1);
    }
  }, [props.collapsed, props.animate]);
  return (
    <div ref={ref} style={styleState} className={props.className}>
      {props.children}
    </div>
  );
}
