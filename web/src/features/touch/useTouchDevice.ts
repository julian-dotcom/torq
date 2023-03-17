import { useEffect, useState } from "react";

const useTouchDevice = () => {
  const [isTouchDeviceState, setIsTouchDeviceState] = useState(false);
  useEffect(() => {
    setIsTouchDeviceState("ontouchstart" in window || navigator.maxTouchPoints > 0);
  }, []);
  return { isTouchDevice: isTouchDeviceState };
};
export default useTouchDevice;
