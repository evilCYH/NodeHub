import { useQuery } from '@tanstack/react-query'
import { api } from '@/src/lib/api/client'

const subLogKeys = {
    all: ['sub-log'] as const,
    list: (subId: number | null, limit: number, includeEvents: boolean) =>
        [...subLogKeys.all, { subId, limit, includeEvents }] as const,
}

export function useSubRunLogs(subId: number | null, limit = 10, enabled = true, includeEvents = false) {
    return useQuery({
        queryKey: subLogKeys.list(subId, limit, includeEvents),
        queryFn: () => api.getSubRunLogs(subId as number, limit, includeEvents),
        enabled: subId !== null && enabled,
        staleTime: 0,
        gcTime: 0,
        refetchInterval: (query) => {
            const data = query.state.data as { runs?: { status: string; stats_pending?: boolean }[] } | undefined
            const hasPending = data?.runs?.some(r => r.status === 'running' || r.stats_pending) ?? false
            return hasPending ? 3000 : false
        },
    })
}
