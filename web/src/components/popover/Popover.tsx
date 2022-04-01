import './popover.scss'
import {ReactChild, useEffect, useRef, useState} from "react";
import classNames from "classnames";


function useOutsideClose(ref: any, setIsPopoverOpen: Function) {
  useEffect(() => {
    function handleClickOutside(event: any) {
      if (ref.current && !ref.current.contains(event.target)) {
        setIsPopoverOpen(false)
      }
    }
    // Bind the event listener
    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      // Unbind the event listener on clean up
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [ref]);
}

interface PopoverInterface {
  className?: string,
  button?: ReactChild,
  children?: ReactChild,
}

const PopoverButton = ({className, button, children }: PopoverInterface) => {
  const wrapperRef = useRef(null)
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  useOutsideClose(wrapperRef, setIsPopoverOpen)

  return (
    <div onClick={() => setIsPopoverOpen(!isPopoverOpen)}
         ref={wrapperRef}
         className={classNames("torq-popover-button-wrapper", className)} >
      {button ? button : "button"}
      <div className={classNames("popover-wrapper", {"popover-open": isPopoverOpen})}
           onClick={(e) =>{e.stopPropagation()}}>
        {children}
      </div>
    </div>
  )
}
export default PopoverButton;
