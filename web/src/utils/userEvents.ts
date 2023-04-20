import mixpanel from "mixpanel-browser";
import { useIntercom } from "react-use-intercom";

export const userEvents = () => {
  const { trackEvent } = useIntercom();

  // Track with both mixpanel and Intercom
  const track = (eventName: string, properties?: any) => {
    mixpanel.track(eventName, properties);
    trackEvent(eventName, properties);
  };

  return {
    track,
  };
};
