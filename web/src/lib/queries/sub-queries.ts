import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/src/lib/api/client'
import type { SubResponse, SubRequest, SubOrderItem } from '@/src/types'

const subKeys = {
    all: ['subs'] as const,
    lists: () => [...subKeys.all, 'list'] as const,
    details: () => [...subKeys.all, 'detail'] as const,
    detail: (id: number) => [...subKeys.details(), id] as const,
}

export function useSubs() {
    return useQuery({
        queryKey: subKeys.lists(),
        queryFn: () => api.getSub(),
        refetchInterval: (query) =>
            Array.isArray(query.state.data) &&
            query.state.data.some(sub => sub.status === 'running')
                ? 2000
                : 60 * 1000,
        notifyOnChangeProps: ['data', 'error', 'isLoading'],
    })
}

export function useCreateSub() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (data: SubRequest) => api.createSubscription(data),
        retry: false, // 禁用重试，避免创建重复订阅

        onSuccess: (newSub) => {
            // 直接更新缓存，不触发重新获取避免重复
            queryClient.setQueryData<SubResponse[]>(
                subKeys.lists(),
                (oldData) => {
                    // 检查是否已存在相同 id 的订阅，避免重复添加
                    if (oldData?.some(sub => sub.id === newSub.id)) {
                        return oldData
                    }
                    return oldData ? [...oldData, newSub] : [newSub]
                }
            )

            queryClient.setQueryData(subKeys.detail(newSub.id), newSub)
        },

        onError: () => {
            queryClient.invalidateQueries({ queryKey: subKeys.lists() })
        },
    })
}

export function useUpdateSub() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: ({ id, data }: { id: number; data: SubRequest }) =>
            api.updateSubscription(id, data),

        onSuccess: (updatedSub, { id }) => {
            queryClient.setQueryData<SubResponse[]>(
                subKeys.lists(),
                (oldData) => oldData?.map(sub =>
                    sub.id === id ? updatedSub : sub
                )
            )

            queryClient.setQueryData(subKeys.detail(id), updatedSub)

            queryClient.invalidateQueries({
                queryKey: subKeys.detail(id),
                refetchType: 'active'
            })
        },

        onError: () => {
            queryClient.invalidateQueries({ queryKey: subKeys.lists() })
        },
    })
}

export function useDeleteSub() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (id: number) => api.deleteSubscription(id),

        onSuccess: (_, id) => {
            queryClient.setQueryData<SubResponse[]>(
                subKeys.lists(),
                (oldData) => oldData?.filter(sub => sub.id !== id)
            )

            queryClient.removeQueries({ queryKey: subKeys.detail(id) })

            queryClient.invalidateQueries({
                queryKey: subKeys.lists(),
                refetchType: 'active'
            })
        },

        onError: () => {
            queryClient.invalidateQueries({ queryKey: subKeys.lists() })
        },
    })
}

export function useRefreshSub() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (id: number) => api.refreshSubscription(id),

        onSuccess: (_, id) => {
            queryClient.invalidateQueries({
                queryKey: subKeys.lists(),
                refetchType: 'active'
            })

            queryClient.invalidateQueries({
                queryKey: subKeys.detail(id),
                refetchType: 'active'
            })
        },

        onError: () => {
            queryClient.invalidateQueries({ queryKey: subKeys.lists() })
        },
    })
}

export function useBatchCreateSub() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (subscriptions: SubRequest[]) => api.batchCreateSubscriptions(subscriptions),

        onSuccess: (results) => {
            if (results.length > 0) {
                queryClient.setQueryData<SubResponse[]>(
                    subKeys.lists(),
                    (oldData) => oldData ? [...oldData, ...results] : results
                )

                results.forEach(sub => {
                    queryClient.setQueryData(subKeys.detail(sub.id), sub)
                })

                queryClient.invalidateQueries({
                    queryKey: subKeys.lists(),
                    refetchType: 'active'
                })
            }
        },

        onError: () => {
            queryClient.invalidateQueries({ queryKey: subKeys.lists() })
        },
    })
} 

export function useUpdateSubOrder() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (orders: SubOrderItem[]) => api.updateSubOrder(orders),
        onMutate: async (orders) => {
            await queryClient.cancelQueries({ queryKey: subKeys.lists() })
            const previousSubs = queryClient.getQueryData<SubResponse[]>(subKeys.lists())

            if (previousSubs) {
                const orderMap = new Map<number, number>(orders.map(item => [item.id, item.sort_order]))
                const sortedSubs = previousSubs.slice().sort((a, b) => {
                    const orderA = orderMap.get(a.id) ?? Number.MAX_SAFE_INTEGER
                    const orderB = orderMap.get(b.id) ?? Number.MAX_SAFE_INTEGER
                    if (orderA === orderB) return a.id - b.id
                    return orderA - orderB
                })
                queryClient.setQueryData(subKeys.lists(), sortedSubs)
            }

            return { previousSubs }
        },
        onError: (_error, _orders, context) => {
            if (context?.previousSubs) {
                queryClient.setQueryData(subKeys.lists(), context.previousSubs)
            }
        },
        onSettled: () => {
            queryClient.invalidateQueries({
                queryKey: subKeys.lists(),
                refetchType: 'active'
            })
        },
    })
}
