import { useMemo, useState, useEffect } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/src/components/ui/dialog"
import { Card, CardContent } from "@/src/components/ui/card"
import { Badge } from "@/src/components/ui/badge"
import { useSubRunLogs } from "@/src/lib/queries/sub-log-queries"
import type { SubResponse } from "@/src/types/sub"
import type { SubRunEvent } from "@/src/types/sub"
import { Loader2, ChevronRight, ChevronDown, CheckCircle2, XCircle, Clock } from "lucide-react"

const STATUS_CONFIG: Record<string, { color: string; label: string; icon: typeof CheckCircle2 }> = {
    success: { color: "bg-green-500/10 text-green-600 border-green-500/20", label: "成功", icon: CheckCircle2 },
    error: { color: "bg-red-500/10 text-red-600 border-red-500/20", label: "失败", icon: XCircle },
    running: { color: "bg-blue-500/10 text-blue-600 border-blue-500/20", label: "运行中", icon: Clock },
}

const EVENT_LEVEL_CONFIG: Record<string, { color: string; label: string }> = {
    info: { color: "bg-blue-500/10 text-blue-500 border-blue-500/20", label: "信息" },
    warn: { color: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20", label: "警告" },
    error: { color: "bg-red-500/10 text-red-500 border-red-500/20", label: "错误" },
}

interface SubLogDialogProps {
    subscription: SubResponse | null
    isOpen: boolean
    onOpenChange: (open: boolean) => void
}

export function SubLogDialog({ subscription, isOpen, onOpenChange }: SubLogDialogProps) {
    const title = subscription ? `${subscription.name} - 订阅日志` : "订阅日志"
    const [expandedRunId, setExpandedRunId] = useState<number | null>(null)

    const { data: subRunData, isLoading } = useSubRunLogs(subscription?.id ?? null, 10, isOpen, true)
    const runs = subRunData?.runs ?? []

    // 按 run_id 分组事件
    const eventsByRunId = useMemo(() => {
        const events = subRunData?.events ?? []
        const map = new Map<number, SubRunEvent[]>()
        for (const event of events) {
            const list = map.get(event.run_id) ?? []
            list.push(event)
            map.set(event.run_id, list)
        }
        return map
    }, [subRunData?.events])

    const latest = runs[0]
    const latestStats = latest?.stats
    const latestStatsPending = Boolean(latest?.stats_pending && !latestStats)

    // 关闭弹窗时重置展开状态
    useEffect(() => {
        if (!isOpen) setExpandedRunId(null)
    }, [isOpen])

    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent className="!max-w-[90vw] !sm:max-w-[1400px] !w-[1400px] max-h-[85vh] flex flex-col overflow-hidden">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                </DialogHeader>

                {isLoading ? (
                    <Card>
                        <CardContent className="py-6">
                            <div className="flex items-center justify-center gap-2 text-muted-foreground">
                                <Loader2 className="h-4 w-4 animate-spin" />
                                加载日志...
                            </div>
                        </CardContent>
                    </Card>
                ) : runs.length > 0 ? (
                    <>
                        {/* 最近一次运行统计卡片 */}
                        {latest && (
                            <Card>
                                <CardContent className="py-4">
                                    <div className="grid grid-cols-3 md:grid-cols-9 gap-4 text-sm">
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">状态</div>
                                            <div className="font-medium">
                                                <Badge variant="outline" className={STATUS_CONFIG[latest.status]?.color}>
                                                    {STATUS_CONFIG[latest.status]?.label ?? latest.status}
                                                </Badge>
                                            </div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">耗时</div>
                                            <div className="font-medium">{latest.duration_ms}ms</div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">原始节点</div>
                                            <div className="font-medium">{latest.raw_count}</div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">解析后节点</div>
                                            <div className="font-medium text-green-600">{latest.accepted}</div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">解析率</div>
                                            <div className="font-medium">
                                                {latest.raw_count > 0
                                                    ? `${Math.round((latest.accepted / latest.raw_count) * 100)}%`
                                                    : "-"}
                                            </div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">通过初测</div>
                                            <div className="font-medium text-green-600">
                                                {latestStats ? latestStats.accepted : latestStatsPending ? "统计生成中" : "-"}
                                            </div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">并入池</div>
                                            <div className="font-medium text-green-600">
                                                {latestStats ? latestStats.merged : latestStatsPending ? "统计生成中" : "-"}
                                            </div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">被淘汰</div>
                                            <div className="font-medium">
                                                {latestStats ? latestStats.dropped : latestStatsPending ? "统计生成中" : "-"}
                                            </div>
                                        </div>
                                        <div className="space-y-1">
                                            <div className="text-muted-foreground text-xs">运行时间</div>
                                            <div className="font-medium font-mono text-xs">
                                                {formatTime(latest.created_at)}
                                            </div>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        )}

                        {/* 运行记录表格（可展开行） */}
                        <Card className="flex-1 min-h-0 overflow-hidden">
                            <CardContent className="p-0 flex h-full min-h-0 flex-col">
                                {/* 表头 */}
                                <div className="border-b bg-muted/50">
                                    <div className="grid grid-cols-[28px_120px_80px_80px_80px_90px_80px_1fr] gap-3 px-4 py-3 font-medium text-sm">
                                        <div></div>
                                        <div>时间</div>
                                        <div>状态</div>
                                        <div>耗时</div>
                                        <div>原始</div>
                                        <div>解析后</div>
                                        <div>初测通过</div>
                                        <div>消息</div>
                                    </div>
                                </div>

                                {/* 表体 */}
                                <div className="flex-1 min-h-0 overflow-y-auto">
                                    <div className="divide-y pb-2">
                                        {runs.map((run) => {
                                            const isExpanded = expandedRunId === run.id
                                            const runEvents = eventsByRunId.get(run.id) ?? []
                                            const hasEvents = runEvents.length > 0
                                            const statusCfg = STATUS_CONFIG[run.status]
                                            const StatusIcon = statusCfg?.icon

                                            return (
                                                <div key={run.id}>
                                                    {/* 运行记录主行 */}
                                                    <div
                                                        className={`grid grid-cols-[28px_120px_80px_80px_80px_90px_80px_1fr] gap-3 px-4 py-3 text-sm hover:bg-muted/50 ${hasEvents ? "cursor-pointer" : ""} ${isExpanded ? "bg-muted/30" : ""}`}
                                                        onClick={() => {
                                                            if (hasEvents) {
                                                                setExpandedRunId(isExpanded ? null : run.id)
                                                            }
                                                        }}
                                                    >
                                                        <div className="flex items-center">
                                                            {hasEvents ? (
                                                                isExpanded
                                                                    ? <ChevronDown className="h-4 w-4 text-muted-foreground" />
                                                                    : <ChevronRight className="h-4 w-4 text-muted-foreground" />
                                                            ) : (
                                                                <div className="w-4" />
                                                            )}
                                                        </div>
                                                        <div className="text-muted-foreground font-mono text-xs flex items-center">
                                                            {formatTime(run.created_at)}
                                                        </div>
                                                        <div className="flex items-center">
                                                            <Badge variant="outline" className={`text-xs ${statusCfg?.color ?? ""}`}>
                                                                {StatusIcon && <StatusIcon className="h-3 w-3 mr-1" />}
                                                                {statusCfg?.label ?? run.status}
                                                            </Badge>
                                                        </div>
                                                        <div className="text-muted-foreground flex items-center font-mono text-xs">
                                                            {run.duration_ms}ms
                                                        </div>
                                                        <div className="flex items-center font-mono text-xs">
                                                            {run.raw_count}
                                                        </div>
                                                        <div className="flex items-center font-mono text-xs text-green-600">
                                                            {run.accepted}
                                                        </div>
                                                        <div className="flex items-center font-mono text-xs text-green-600">
                                                            {run.stats
                                                                ? run.stats.accepted
                                                                : run.stats_pending
                                                                    ? "生成中"
                                                                    : "-"}
                                                        </div>
                                                        <div className="text-muted-foreground text-xs flex items-center truncate">
                                                            {run.message || (hasEvents ? `${runEvents.length} 条事件` : "")}
                                                        </div>
                                                    </div>

                                                    {/* 展开的事件详情 */}
                                                    {isExpanded && runEvents.length > 0 && (
                                                        <div className="bg-muted/20 border-t">
                                                            <div className="ml-7 mr-4 my-2">
                                                                {/* 事件表头 */}
                                                                <div className="grid grid-cols-[100px_80px_70px_1fr] gap-3 px-3 py-2 text-xs font-medium text-muted-foreground border-b border-border/50">
                                                                    <div>时间</div>
                                                                    <div>阶段</div>
                                                                    <div>级别</div>
                                                                    <div>消息</div>
                                                                </div>
                                                                {/* 事件行 */}
                                                                <div className="divide-y divide-border/30">
                                                                    {runEvents.map((event) => {
                                                                        const levelCfg = EVENT_LEVEL_CONFIG[event.level]
                                                                        return (
                                                                            <div
                                                                                key={event.id}
                                                                                className="grid grid-cols-[100px_80px_70px_1fr] gap-3 px-3 py-2 text-xs hover:bg-muted/30"
                                                                            >
                                                                                <div className="text-muted-foreground font-mono flex items-center">
                                                                                    {formatTime(event.run_time ?? event.created_at)}
                                                                                </div>
                                                                                <div className="flex items-center">
                                                                                    <Badge variant="outline" className="text-[10px] px-1.5 py-0 rounded-sm">
                                                                                        {event.step}
                                                                                    </Badge>
                                                                                </div>
                                                                                <div className="flex items-center">
                                                                                    <Badge variant="outline" className={`text-[10px] px-1.5 py-0 rounded-sm ${levelCfg?.color ?? ""}`}>
                                                                                        {levelCfg?.label ?? event.level}
                                                                                    </Badge>
                                                                                </div>
                                                                                <div className="break-all flex items-center">
                                                                                    {event.message}
                                                                                </div>
                                                                            </div>
                                                                        )
                                                                    })}
                                                                </div>
                                                            </div>
                                                        </div>
                                                    )}
                                                </div>
                                            )
                                        })}
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    </>
                ) : (
                    <Card>
                        <CardContent className="py-6">
                            <div className="text-center text-muted-foreground">暂无订阅日志</div>
                        </CardContent>
                    </Card>
                )}
            </DialogContent>
        </Dialog>
    )
}

function formatTime(iso: string): string {
    const d = new Date(iso)
    if (isNaN(d.getTime())) return iso
    return (
        `${(d.getMonth() + 1).toString().padStart(2, "0")}-${d.getDate().toString().padStart(2, "0")} ` +
        `${d.getHours().toString().padStart(2, "0")}:${d.getMinutes().toString().padStart(2, "0")}:${d.getSeconds().toString().padStart(2, "0")}`
    )
}
