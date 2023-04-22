import mixpanel from "mixpanel-browser";
import { useIntercom } from "react-use-intercom";
import { AnyObject } from "./types";

export const userEvents = () => {
  const { trackEvent, update } = useIntercom();

  const register = (properties: AnyObject) => {
    mixpanel.register(properties);
    update({ customAttributes: properties });
  };

  // Track with both mixpanel and Intercom
  const track = (eventName: string, properties?: AnyObject) => {
    mixpanel.track(eventName, properties);
    trackEvent(eventName, properties);
  };

  return {
    track,
    register,
  };
};
