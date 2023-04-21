import mixpanel from "mixpanel-browser";
import { useIntercom } from "react-use-intercom";
import { AnyObject } from "./types";

export const userEvents = () => {
  const { trackEvent } = useIntercom();

  // Track with both mixpanel and Intercom
  const track = (eventName: string, properties?: AnyObject) => {
    mixpanel.track(eventName, properties);
    trackEvent(eventName, properties);
  };

  return {
    track,
  };
};
