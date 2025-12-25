export interface Release {
  name: string;
  namespace: string;
  chart: string;
  chartVersion: string;
  appVersion: string;
  status: string;
  updated: string;
  revision: number;
}

export interface ChartVersion {
  version: string;
  appVersion: string;
  description: string;
}

export interface ReleaseHistory {
  revision: number;
  updated: string;
  status: string;
  chart: string;
  appVersion: string;
  description: string;
}

export interface Repository {
  name: string;
  url: string;
}

export interface VersionUpgradeRequest {
  chartVersion: string;
  values?: Record<string, unknown>;
}

export interface AddRepositoryRequest {
  name: string;
  url: string;
}
