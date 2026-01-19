import axios from 'axios';
import type {
  Release,
  ReleaseFilter,
  ChartVersion,
  ReleaseHistory,
  RegistryMapping,
  SetRegistryRequest,
  VersionUpgradeRequest,
  ValuesUpdateRequest,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Release APIs
export const getReleases = async (filter?: ReleaseFilter): Promise<Release[]> => {
  const params = new URLSearchParams();
  if (filter?.namespace) {
    params.append('namespace', filter.namespace);
  }
  if (filter?.hasRegistry !== undefined) {
    params.append('hasRegistry', String(filter.hasRegistry));
  }
  const queryString = params.toString();
  const url = queryString ? `/releases?${queryString}` : '/releases';
  const { data } = await client.get<Release[]>(url);
  return data;
};

export const getRelease = async (namespace: string, name: string): Promise<Release> => {
  const { data } = await client.get<Release>(`/releases/${namespace}/${name}`);
  return data;
};

export const getAvailableVersions = async (namespace: string, name: string): Promise<ChartVersion[]> => {
  const { data } = await client.get<ChartVersion[]>(`/releases/${namespace}/${name}/versions`);
  return data;
};

export const upgradeRelease = async (
  namespace: string,
  name: string,
  request: VersionUpgradeRequest
): Promise<Release> => {
  const { data } = await client.put<Release>(`/releases/${namespace}/${name}`, request);
  return data;
};

export const getReleaseHistory = async (namespace: string, name: string): Promise<ReleaseHistory[]> => {
  const { data } = await client.get<ReleaseHistory[]>(`/releases/${namespace}/${name}/history`);
  return data;
};

// Registry APIs
export const getRegistry = async (namespace: string, name: string): Promise<RegistryMapping> => {
  const { data } = await client.get<RegistryMapping>(`/releases/${namespace}/${name}/registry`);
  return data;
};

export const setRegistry = async (
  namespace: string,
  name: string,
  request: SetRegistryRequest
): Promise<RegistryMapping> => {
  const { data } = await client.put<RegistryMapping>(`/releases/${namespace}/${name}/registry`, request);
  return data;
};

export const deleteRegistry = async (namespace: string, name: string): Promise<void> => {
  await client.delete(`/releases/${namespace}/${name}/registry`);
};

// Values APIs
export const getValues = async (namespace: string, name: string): Promise<Record<string, unknown>> => {
  const { data } = await client.get<Record<string, unknown>>(`/releases/${namespace}/${name}/values`);
  return data;
};

export const updateValues = async (
  namespace: string,
  name: string,
  request: ValuesUpdateRequest
): Promise<Release> => {
  const { data } = await client.put<Release>(`/releases/${namespace}/${name}/values`, request);
  return data;
};

// Rollback API
export const rollbackRelease = async (
  namespace: string,
  name: string,
  revision: number
): Promise<Release> => {
  const { data } = await client.post<Release>(`/releases/${namespace}/${name}/rollback`, { revision });
  return data;
};
