import type { NodeUpdateLog } from './node'

export interface SubConfig {
    url: string
    proxy?: boolean
    timeout?: number
    protocol_filter_enable?: boolean
    protocol_filter_mode?: boolean
    protocol_filter?: string[]
}

export interface SubRequest {
    name: string
    tags: string[]
    enable: boolean
    cron_expr: string
    config: SubConfig
}


export interface SubResult {
    success: number
    fail: number
    msg: string
    last_status?: 'success' | 'error'
    raw_count: number
    node_null_count?: number
    last_run: string
    duration: number
}

export interface SubNodeInfo {
    speed_up: number
    speed_down: number
    delay: number
    risk: number
    count: number
}

export interface SubResponse {
    id: number
    name: string
    tags: string[]
    enable: boolean
    cron_expr: string
    config: SubConfig
    status: string
    result: SubResult
    info: SubNodeInfo
    upload: number
    download: number
    total: number
    expire: number
    info_updated_at?: string | null
    created_at: string
    updated_at: string
}

export interface SubRunLog {
    id: number
    sub_id: number
    status: string
    message?: string
    raw_count: number
    accepted: number
    created_at: string
    duration_ms: number
}

export interface SubRunWithStats extends SubRunLog {
    stats?: NodeUpdateLog
    stats_pending?: boolean
}

export interface SubRunEvent {
    id: number
    run_id: number
    sub_id: number
    run_time?: string
    step: string
    level: string
    message: string
    created_at: string
}

export interface SubRunLogResponse {
    runs: SubRunWithStats[]
    events: SubRunEvent[]
}

export interface SubNameAndID {
    id: number
    name: string
}

export interface SubOrderItem {
    id: number
    sort_order: number
}
