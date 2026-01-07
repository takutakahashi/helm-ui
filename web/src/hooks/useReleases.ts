import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getReleases,
  getRelease,
  getAvailableVersions,
  upgradeRelease,
  getReleaseHistory,
  getRegistry,
  setRegistry,
  getValues,
  updateValues,
} from '../api/client';
import type { VersionUpgradeRequest, SetRegistryRequest, ValuesUpdateRequest, ReleaseFilter } from '../types';

export const useReleases = (filter?: ReleaseFilter) => {
  return useQuery({
    queryKey: ['releases', filter],
    queryFn: () => getReleases(filter),
  });
};

export const useRelease = (namespace: string, name: string) => {
  return useQuery({
    queryKey: ['release', namespace, name],
    queryFn: () => getRelease(namespace, name),
    enabled: !!namespace && !!name,
  });
};

export const useAvailableVersions = (namespace: string, name: string) => {
  return useQuery({
    queryKey: ['versions', namespace, name],
    queryFn: () => getAvailableVersions(namespace, name),
    enabled: !!namespace && !!name,
  });
};

export const useReleaseHistory = (namespace: string, name: string) => {
  return useQuery({
    queryKey: ['history', namespace, name],
    queryFn: () => getReleaseHistory(namespace, name),
    enabled: !!namespace && !!name,
  });
};

export const useUpgradeRelease = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      namespace,
      name,
      request,
    }: {
      namespace: string;
      name: string;
      request: VersionUpgradeRequest;
    }) => upgradeRelease(namespace, name, request),
    onSuccess: (_, { namespace, name }) => {
      queryClient.invalidateQueries({ queryKey: ['releases'] });
      queryClient.invalidateQueries({ queryKey: ['release', namespace, name] });
      queryClient.invalidateQueries({ queryKey: ['history', namespace, name] });
    },
  });
};

export const useRegistry = (namespace: string, name: string) => {
  return useQuery({
    queryKey: ['registry', namespace, name],
    queryFn: () => getRegistry(namespace, name),
    enabled: !!namespace && !!name,
    retry: false,
  });
};

export const useSetRegistry = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      namespace,
      name,
      request,
    }: {
      namespace: string;
      name: string;
      request: SetRegistryRequest;
    }) => setRegistry(namespace, name, request),
    onSuccess: (_, { namespace, name }) => {
      queryClient.invalidateQueries({ queryKey: ['registry', namespace, name] });
      queryClient.invalidateQueries({ queryKey: ['versions', namespace, name] });
    },
  });
};

export const useValues = (namespace: string, name: string) => {
  return useQuery({
    queryKey: ['values', namespace, name],
    queryFn: () => getValues(namespace, name),
    enabled: !!namespace && !!name,
  });
};

export const useUpdateValues = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      namespace,
      name,
      request,
    }: {
      namespace: string;
      name: string;
      request: ValuesUpdateRequest;
    }) => updateValues(namespace, name, request),
    onSuccess: (_, { namespace, name }) => {
      queryClient.invalidateQueries({ queryKey: ['releases'] });
      queryClient.invalidateQueries({ queryKey: ['release', namespace, name] });
      queryClient.invalidateQueries({ queryKey: ['values', namespace, name] });
      queryClient.invalidateQueries({ queryKey: ['history', namespace, name] });
    },
  });
};
