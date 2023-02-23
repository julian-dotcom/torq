export const REST_API_PATHNAME = "/api";
export const WS_API_PATHNAME = "/ws";
export const DEV_FRONTEND_PORT = "3000";
export const DEV_BACKEND_PORT = "8080";

export enum Status {
  Inactive,
  Active,
  Pending,
  Deleted,
  Initializing,
  Archived,
}

export enum ChannelStatus {
  Opening = 0,
  Open = 1,
  Closing = 2,
  CooperativeClosed = 100,
  LocalForceClosed = 101,
  RemoteForceClosed = 102,
  BreachClosed = 103,
  FundingCancelledClosed = 104,
  AbandonedClosed = 105,
}
