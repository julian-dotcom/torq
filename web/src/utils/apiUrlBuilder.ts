import { DEV_BACKEND_PORT, DEV_FRONTEND_PORT, REST_API_PATHNAME, WS_API_PATHNAME } from "constants/backend";
import { BASE_PATHNAME_PREFIX } from "constants/subpath-support";

const isRunningOnSubpath = window.location.pathname.startsWith(BASE_PATHNAME_PREFIX);

const basePathnamePrefix = isRunningOnSubpath ? BASE_PATHNAME_PREFIX : "";

const isRunningOnSecureConnection = window.location.protocol.startsWith("https");

const buildBaseUrl = () => {
  const address =
    window.location.port === DEV_FRONTEND_PORT.toString()
      ? "//" + window.location.hostname + ":" + DEV_BACKEND_PORT.toString()
      : "//" + window.location.host;
  return `//${address}${basePathnamePrefix}`;
};

export const getRestEndpoint = () => buildBaseUrl() + REST_API_PATHNAME;

export const getWsEndpoint = () => {
  const protocol = isRunningOnSecureConnection ? "wss" : "ws";
  const baseApiUrl = buildBaseUrl();
  return `${protocol}:${baseApiUrl}${WS_API_PATHNAME}`;
};

export const getStaticEndpoint = () => buildBaseUrl();
