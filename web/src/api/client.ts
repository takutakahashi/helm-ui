import axios from 'axios';
import type {
  Release,
  ChartVersion,
  ReleaseHistory,
  Repository,
  VersionUpgradeRequest,
  AddRepositoryRequest,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Release APIs
export const getReleases = async (): Promise<Release[]> => {
  const { data } = await client.get<Release[]>('/releases');
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

// Repository APIs
export const getRepositories = async (): Promise<Repository[]> => {
  const { data } = await client.get<Repository[]>('/repositories');
  return data;
};

export const addRepository = async (request: AddRepositoryRequest): Promise<void> => {
  await client.post('/repositories', request);
};

export const removeRepository = async (name: string): Promise<void> => {
  await client.delete(`/repositories/${name}`);
};

export const updateRepository = async (name: string): Promise<void> => {
  await client.post(`/repositories/${name}/update`);
};
