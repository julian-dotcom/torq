import { isStringLiteral } from "typescript";

export interface settings {
  defaultDateRange: string;
  preferredTimezone: string;
  weekStartsOn: string;
}

export interface timeZone {
  name: string;
}

export interface localNode {
  localNodeId?: number;
  implementation: string;
  grpcAddress?: string;
  tlsFileName?: string;
  tlsData?: Uint8Array;
  macaroonFileName?: string;
  macaroonData?: Uint8Array;
  createdOn?: Date;
  updatedOn?: Date;
}
