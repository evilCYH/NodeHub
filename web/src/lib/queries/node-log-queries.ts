import { useQuery, useInfiniteQuery } from '@tanstack/react-query'
import { api } from '@/src/lib/api/client'

const nodeLogKeys = {
    all: ['node-log'] as const,
    detail: (subId: number | null) => [...nodeLogKeys.all, { subId }] as const,
    testLogs: (subId: number | null, level?: string, keyword?: string) =>
        [...nodeLogKeys.all, 'test', { subId, level, keyword }] as const,
    logs: (subId: number | null, source?: string, level?: string, keyword?: string, checkId?: number) =>
        [...nodeLogKeys.all, 'logs', { subId, source, level, keyword, checkId }] as const,
}

export function useNodeUpdateLog(subId: number | null, limit = 5, enabled = true, runId?: number | null) {
    return useQuery({
        queryKey: [...nodeLogKeys.detail(subId), { runId: runId ?? null, limit }],
        queryFn: () => api.getNodeUpdateLog(subId as number, limit, runId ?? undefined),
        enabled: subId !== null && enabled,
        refetchInterval: 60 * 1000,
        notifyOnChangeProps: ['data', 'error', 'isLoading'],
    })
}

interface UseNodeTestLogsOptions {
    subId: number | null
    level?: string | undefined
    keyword?: string | undefined
    pageSize?: number
    enabled?: boolean
}

export function useNodeTestLogs(options: UseNodeTestLogsOptions) {
    const { subId, level, keyword, pageSize = 50, enabled = true } = options

    return useInfiniteQuery({
        queryKey: nodeLogKeys.testLogs(subId, level, keyword),
        queryFn: async ({ pageParam = 1 }) => {
            return api.getNodeTestLogs({
                subId: subId as number,
                level,
                keyword,
                page: pageParam,
                pageSize,
            })
        },
        getNextPageParam: (lastPage, pages) => {
            const loaded = pages.length * pageSize
            return loaded < lastPage.total ? pages.length + 1 : undefined
        },
        enabled: subId !== null && enabled,
        initialPageParam: 1,
    })
}

interface UseNodeTestLogsPaginatedOptions {
    subId: number | null
    level?: string | undefined
    keyword?: string | undefined
    page: number
    pageSize?: number
    enabled?: boolean
}

interface UseNodeLogsOptions {
    subId: number | null
    source?: string | undefined
    level?: 'info' | 'warn' | 'error' | undefined
    keyword?: string | undefined
    checkId?: number
    page: number
    pageSize?: number
    enabled?: boolean
}

export function useNodeLogs(options: UseNodeLogsOptions) {
    const { subId, source, level, keyword, checkId, page, pageSize = 50, enabled = true } = options

    return useQuery({
        queryKey: [...nodeLogKeys.logs(subId, source, level, keyword, checkId), { page, pageSize }],
        queryFn: async () => {
            return api.getNodeLogs({
                subId: subId as number,
                source,
                level,
                keyword,
                checkId,
                page,
                pageSize,
            })
        },
        enabled: subId !== null && enabled,
        notifyOnChangeProps: ['data', 'error', 'isLoading'],
    })
}

export function useNodeTestLogsPaginated(options: UseNodeTestLogsPaginatedOptions) {
    const { subId, level, keyword, page, pageSize = 50, enabled = true } = options

    return useQuery({
        queryKey: [...nodeLogKeys.testLogs(subId, level, keyword), { page }],
        queryFn: async () => {
            return api.getNodeTestLogs({
                subId: subId as number,
                level,
                keyword,
                page,
                pageSize,
            })
        },
        enabled: subId !== null && enabled,
        notifyOnChangeProps: ['data', 'error', 'isLoading'],
    })
}
