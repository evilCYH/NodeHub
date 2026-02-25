/**
 * 格式化相关工具函数
 */

import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * 合并CSS类名
 */
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
}

/**
 * 格式化持续时间
 */
export function formatDuration(milliseconds: number): string {
    if (milliseconds < 1000) {
        return `${milliseconds}ms`
    } else if (milliseconds < 60000) {
        return `${(milliseconds / 1000).toFixed(1)}s`
    } else {
        const minutes = Math.floor(milliseconds / 60000)
        const seconds = Math.floor((milliseconds % 60000) / 1000)
        return `${minutes}m ${seconds}s`
    }
}



/**
 * 格式化最后运行时间（通用）
 */
export function formatLastRunTime(lastRun: string | undefined): string {
    if (!lastRun) return '从未运行'

    // 检查是否为零时间
    const zeroTimePatterns = [
        '0001-01-01T00:00:00Z',
        '0001-01-01T00:00:00.000Z',
        '1970-01-01T00:00:00Z',
        '1970-01-01T00:00:00.000Z'
    ]

    if (zeroTimePatterns.includes(lastRun)) {
        return '从未运行'
    }

    try {
        return new Date(lastRun).toLocaleString('zh-CN')
    } catch {
        return '时间格式错误'
    }
}

/**
 * 格式化布尔值显示
 */
export function formatBooleanText(value: boolean): string {
    return value ? '启用' : '禁用'
}

export function formatBytes(bytes: number): string {
    if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'

    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let value = bytes
    let unitIndex = 0

    while (value >= 1024 && unitIndex < units.length - 1) {
        value /= 1024
        unitIndex++
    }

    if (unitIndex === 0) {
        return `${Math.round(value)} ${units[unitIndex]}`
    }
    return `${value.toFixed(2)} ${units[unitIndex]}`
}

export interface TrafficSummary {
    usedBytes: number
    usedText: string
    totalText: string
    isUnlimited: boolean
    isOverLimit: boolean
    usagePercent: number | null
}

export function formatTrafficSummary(upload: number, download: number, total: number): TrafficSummary {
    const safeUpload = Number.isFinite(upload) ? upload : 0
    const safeDownload = Number.isFinite(download) ? download : 0
    const safeTotal = Number.isFinite(total) ? total : 0

    const usedBytes = Math.max(0, safeUpload) + Math.max(0, safeDownload)
    const isUnlimited = safeTotal === 0 || safeTotal === -1
    const isOverLimit = safeTotal > 0 && usedBytes > safeTotal
    const usagePercent = safeTotal > 0 ? (usedBytes / safeTotal) * 100 : null

    return {
        usedBytes,
        usedText: formatBytes(usedBytes),
        totalText: isUnlimited ? '无限流量' : formatBytes(Math.max(0, safeTotal)),
        isUnlimited,
        isOverLimit,
        usagePercent,
    }
}

export type ExpireState = 'permanent' | 'expired' | 'warning' | 'normal'

export interface ExpireStatus {
    label: string
    state: ExpireState
}

export function hasSubscriptionInfo(infoUpdatedAt?: string | null): boolean {
    if (!infoUpdatedAt) return false
    const time = new Date(infoUpdatedAt).getTime()
    return Number.isFinite(time)
}

export function formatExpireStatus(expire: number): ExpireStatus {
    const safeExpire = Number.isFinite(expire) ? expire : 0
    if (safeExpire === 0 || safeExpire >= 9999999999) {
        return { label: '永久有效', state: 'permanent' }
    }

    const nowMs = Date.now()
    const expireMs = safeExpire * 1000
    if (expireMs <= nowMs) {
        return { label: '已过期', state: 'expired' }
    }

    const warningThresholdMs = 7 * 24 * 60 * 60 * 1000
    if (expireMs-nowMs <= warningThresholdMs) {
        return { label: '即将到期', state: 'warning' }
    }

    return {
        label: new Date(expireMs).toLocaleString('zh-CN'),
        state: 'normal',
    }
}
