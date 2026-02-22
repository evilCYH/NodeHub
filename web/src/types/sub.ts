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
    runs: SubRunLog[]
    events: SubRunEvent[]
}

export interface SubNameAndID {
    id: number
    name: string
}
