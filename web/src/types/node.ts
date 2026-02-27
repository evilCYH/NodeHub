export interface NodeResponse {
    sub_id: number
    unique_key: number
    unique_key_str?: string
    name: string
    type: string
    reason?: string
    delay: number
    speed_up: number
    speed_down: number
    risk: number
    alive_status: number
    country: string
    init_status?: 'unknown' | 'passed' | 'failed'
    last_check_at?: string
    last_check_source?: string
    last_fail_reason?: string
}

export interface NodeDetailResponse extends NodeResponse {
    raw: Record<string, unknown>
}

export interface NodeUpdateLog {
    id?: number
    sub_id: number
    run_id: number
    created_at: string
    duration_ms: number
    raw_count: number
    candidate: number
    duplicate: number
    invalid: number
    test_failed: number
    accepted: number
    merged: number
    dropped: number
    details?: string[]
}

export interface NodeUpdateLogResponse {
    latest?: NodeUpdateLog
    history: NodeUpdateLog[]
}

// 节点详细测试日志
export interface NodeTestLog {
    id: number
    sub_id: number
    node_name: string
    level: 'info' | 'warn' | 'error'
    message: string
    created_at: string
}

export interface NodeTestLogResponse {
    total: number
    list: NodeTestLog[]
}

export interface NodeLog {
    id: number
    sub_id: number
    node_key: string
    node_name: string
    level: 'info' | 'warn' | 'error'
    source: 'init' | 'check' | 'unknown'
    run_id: number
    check_id: number
    message: string
    created_at: string
}

export interface NodeLogResponse {
    total: number
    list: NodeLog[]
}
