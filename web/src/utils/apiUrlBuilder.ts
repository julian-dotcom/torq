import { PROD_ENVIRONMENT_PORT, REST_API_PATHNAME } from "constants/backend";
import { BASE_PATHNAME_PREFIX } from "constants/third-party-integration";

const isRunningOnThirdParty = window.location.pathname.startsWith(BASE_PATHNAME_PREFIX);

const basePathnamePrefix = isRunningOnThirdParty ? BASE_PATHNAME_PREFIX : "";

const isRunningOnSecureConnection = window.location.protocol.startsWith("https");

const buildBaseUrl = () => {
  const { hostname } = location;

  const url = `//${hostname}:${PROD_ENVIRONMENT_PORT}${basePathnamePrefix}`;

  return url;
};

export const getRestEndpoint = () => buildBaseUrl() + REST_API_PATHNAME;

export const getWsEndpoint = () => {
  const protocol = isRunningOnSecureConnection ? "wss" : "ws";

  const baseApiUrl = buildBaseUrl();

  const url = `${protocol}:${baseApiUrl}/ws`;

  return url;
};
