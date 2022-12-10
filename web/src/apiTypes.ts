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
  tlsFile?: File | null;
  macaroonFileName?: string;
  macaroonFile?: File | null;
  createdOn?: Date;
  updatedOn?: Date;
  status: number;
  pingSystem: number;
  customSettings: number;
}

export interface stringMap<T> {
  [key: string]: T;
}

export interface services {
  torqService: {
    status: number;
    bootTime: string | null;
    version: string;
  };
  lndServices: {
    status: number;
    bootTime: string | null;
    nodeId: number;
    transactionStreamStatus: number;
    transactionStreamBootTime: string | null;
    htlcEventStreamStatus: number;
    htlcEventStreamBootTime: string | null;
    channelEventStreamStatus: number;
    channelEventStreamBootTime: string | null;
    graphEventStreamStatus: number;
    graphEventStreamBootTime: string | null;
    forwardStreamStatus: number;
    forwardStreamBootTime: string | null;
    invoiceStreamStatus: number;
    invoiceStreamBootTime: string | null;
    paymentStreamStatus: number;
    paymentStreamBootTime: string | null;
    inFlightPaymentStreamStatus: number;
    inFlightPaymentStreamBootTime: string | null;
    peerEventStreamStatus: number;
    peerEventStreamBootTime: string | null;
  }[];
  vectorServices: {
    status: number;
    bootTime: string | null;
    nodeId: number;
  }[];
  ambossServices: {
    status: number;
    bootTime: string | null;
    nodeId: number;
  }[];
}
