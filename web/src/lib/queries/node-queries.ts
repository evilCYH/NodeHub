import { useQuery } from '@tanstack/react-query'
import { api } from '@/src/lib/api/client'

const nodeKeys = {
    all: ['nodes'] as const,
    lists: () => [...nodeKeys.all, 'list'] as const,
    listBySub: (subId: number | null) => [...nodeKeys.lists(), { subId }] as const,
    detail: (subId: number | null, uniqueKey: number | string | null, scope: 'registry' | 'pool') =>
        [...nodeKeys.all, 'detail', { subId, uniqueKey, scope }] as const,
}

export function useNodes(
    subId: number | null,
    options?: {
        includeFailed?: boolean
        scope?: 'registry' | 'pool'
        status?: 'alive' | 'dead' | 'init_failed' | 'all'
    }
) {
    const queryOptions = {
        ...(subId !== null ? { subId } : {}),
        ...options,
    }

    return useQuery({
        queryKey: [...nodeKeys.listBySub(subId), options ?? {}],
        queryFn: () => api.getNodes(queryOptions),
        enabled: subId !== null,
        refetchInterval: 60 * 1000,
        notifyOnChangeProps: ['data', 'error', 'isLoading'],
    })
}

export function useNodeDetail(
    subId: number | null,
    uniqueKey: number | string | null,
    scope: 'registry' | 'pool' = 'registry',
    enabled = true
) {
    return useQuery({
        queryKey: nodeKeys.detail(subId, uniqueKey, scope),
        queryFn: () => api.getNodeDetail({ subId: subId!, uniqueKey: uniqueKey!, scope }),
        enabled: enabled && subId !== null && uniqueKey !== null,
        staleTime: 30 * 1000,
        notifyOnChangeProps: ['data', 'error', 'isLoading'],
    })
}
