import { isStringLiteral } from "typescript";

export interface settings {
  defaultDateRange: string;
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
}

export interface timeZone {
  name: string;
}

export interface localNode {
  localNodeId?: number;
  implementation: string;
  grpcAddress?: string;
  tlsFileName?: string;
  tlsFile: File;
  macaroonFileName?: string;
  macaroonFile: File;
  createdOn?: Date;
  updatedOn?: Date;
}
