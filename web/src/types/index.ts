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

export interface RegistryMapping {
  namespace: string;
  releaseName: string;
  chartName: string;
  registry: string;
}

export interface SetRegistryRequest {
  registry: string;
}

export interface VersionUpgradeRequest {
  chartVersion: string;
  values?: Record<string, unknown>;
}
