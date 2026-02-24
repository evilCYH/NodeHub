"use client"

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState } from 'react'

type ErrorWithCode = {
    code?: number
}

function shouldRetry(failureCount: number, error: unknown, maxAttempts: number) {
    const code = typeof error === 'object' && error !== null && 'code' in error
        ? (error as ErrorWithCode).code
        : undefined

    if (typeof code === 'number' && code >= 400 && code < 500) {
        return false
    }
    return failureCount < maxAttempts
}

export function QueryProvider({ children }: { children: React.ReactNode }) {
    const [queryClient] = useState(() => new QueryClient({
        defaultOptions: {
            queries: {
                staleTime: 5 * 60 * 1000,
                gcTime: 10 * 60 * 1000,
                refetchOnWindowFocus: true,
                refetchOnReconnect: true,
                refetchInterval: 5 * 60 * 1000,
                refetchIntervalInBackground: false,
                retry: (failureCount, error) => shouldRetry(failureCount, error, 3),
                retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
            },
            mutations: {
                retry: (failureCount, error) => shouldRetry(failureCount, error, 2),
                retryDelay: 1500,
            },
        },
    }))

    return (
        <QueryClientProvider client={queryClient}>
            {children}
        </QueryClientProvider>
    )
} 
