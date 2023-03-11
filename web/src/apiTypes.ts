export interface settings {
  defaultDateRange: string;
  defaultLanguage: "en" | "nl";
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
  torqUuid: string;
  mixpanelOptOut: boolean;
}
export interface updateSettingsRequest {
  defaultDateRange: string;
  defaultLanguage: "en" | "nl";
  preferredTimezone: string;
  weekStartsOn: "saturday" | "sunday" | "monday";
  mixpanelOptOut: boolean;
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
    statusString: string;
    bootTime: string | null;
    version: string;
  };
  lndServices: {
    status: number;
    statusString: string;
    bootTime: string | null;
    nodeId: number;
    type: number;
    typeString: string;
    streamStatus: {
      status: number;
      statusString: string;
      bootTime: string | null;
      nodeId: number;
      type: number;
      typeString: string;
    }[];
  }[];
  services: {
    status: number;
    statusString: string;
    bootTime: string | null;
    nodeId: number | null;
    type: number;
    typeString: string;
  }[];
  vectorServices: {
    status: number;
    statusString: string;
    bootTime: string | null;
    nodeId: number;
  }[];
  ambossServices: {
    status: number;
    statusString: string;
    bootTime: string | null;
    nodeId: number;
  }[];
}

export interface lndServices {
  status: number;
  statusString: string;
  bootTime: string | null;
  nodeId: number;
  type: number;
  typeString: string;
  streamStatus: {
    status: number;
    statusString: string;
    bootTime: string | null;
    nodeId: number;
    type: number;
    typeString: string;
  }[];
}

export interface nodeWalletBalances {
  nodeId: number;
  totalBalance: number;
  confirmedBalance: number;
  unconfirmedBalance: number;
  lockedBalance: number;
  reservedBalanceAnchorChan: number;
}
