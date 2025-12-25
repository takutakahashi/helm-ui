import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getRepositories,
  addRepository,
  removeRepository,
  updateRepository,
} from '../api/client';
import type { AddRepositoryRequest } from '../types';

export const useRepositories = () => {
  return useQuery({
    queryKey: ['repositories'],
    queryFn: getRepositories,
  });
};

export const useAddRepository = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: AddRepositoryRequest) => addRepository(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['repositories'] });
    },
  });
};

export const useRemoveRepository = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (name: string) => removeRepository(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['repositories'] });
    },
  });
};

export const useUpdateRepository = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (name: string) => updateRepository(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['repositories'] });
      queryClient.invalidateQueries({ queryKey: ['versions'] });
    },
  });
};
