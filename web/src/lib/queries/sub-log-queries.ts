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
        notifyOnChangeProps: ['data', 'error', 'isLoading'],
    })
}
