import "./popover.scss";
import React, { useRef, useState, ReactNode } from "react";
import classNames from "classnames";
import { useClickOutside } from "utils/hooks";

interface PopoverInterface {
  className?: string;
  button?: ReactNode;
  intercomTarget?: string;
  children?: ReactNode;
}

const PopoverButton = React.forwardRef(function popoverButton(
  { className, button, intercomTarget, children }: PopoverInterface,
  ref
) {
  const wrapperRef = useRef(null);
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  useClickOutside(wrapperRef, () => setIsPopoverOpen(false));

  React.useImperativeHandle(ref, () => ({
    close() {
      setIsPopoverOpen(false);
    },
  }));

  return (
    <div
      onClick={() => setIsPopoverOpen(!isPopoverOpen)}
      ref={wrapperRef}
      className={classNames("torq-popover-button-wrapper", className)}
      data-intercom-target={intercomTarget}
    >
      {button ? button : "button"}
      <div
        className={classNames("popover-wrapper right", {
          "popover-open": isPopoverOpen,
        })}
        onClick={(e) => {
          e.stopPropagation();
        }}
      >
        <div className={"popover-mobile-dismiss"}>
          <div className="left-container" onClick={(e) => e.stopPropagation()}>
            {button ? button : ""}
          </div>
          <div className="right-container dismiss-button" onClick={() => setIsPopoverOpen(false)}>
            Close
          </div>
        </div>
        {isPopoverOpen && <div className="popover-container">{children}</div>}
      </div>
    </div>
  );
});
export default PopoverButton;
