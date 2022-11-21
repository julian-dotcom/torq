export interface settings {
  defaultDateRange: string;
  defaultLanguage: "en" | "nl";
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
}

export interface timeZone {
  name: string;
}

export interface nodeConfiguration {
  nodeId: number;
  name?: string;
  implementation: number;
  grpcAddress?: string;
  tlsFileName?: string;
  tlsFile: File | null;
  macaroonFileName?: string;
  macaroonFile: File | null;
  createdOn?: Date;
  updatedOn?: Date;
  status: number;
}

export interface stringMap<T> {
  [key: string]: T;
}
