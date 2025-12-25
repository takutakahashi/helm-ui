import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getReleases,
  getRelease,
  getAvailableVersions,
  upgradeRelease,
  getReleaseHistory,
} from '../api/client';
import type { VersionUpgradeRequest } from '../types';

export const useReleases = () => {
  return useQuery({
    queryKey: ['releases'],
    queryFn: getReleases,
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
