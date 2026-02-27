import { useMemo, useState } from "react"
import { Card, CardContent } from "@/src/components/ui/card"
import { Button } from "@/src/components/ui/button"
import { InlineLoading } from "@/src/components/ui/loading"
import { FileText, Activity, Zap, ArrowUp, ArrowDown, HeartPulse } from "lucide-react"
import { cn, formatRelativeTime } from "@/src/utils"
import { formatSpeed } from "@/src/components/features/sub/utils"
import { NodeLogDialog } from "./node-log-dialog"
import { NODE_STATUS } from "../constants"
import type { NodeResponse, SubResponse } from "@/src/types"

interface SubscriptionColumnProps {
    subs: SubResponse[]
    registrySummaryNodes: NodeResponse[]
    isLoading: boolean
    error: Error | null
    selectedId: number | null
    onSelect: (sub: SubResponse) => void
}

export function SubscriptionColumn({ subs, registrySummaryNodes, isLoading, error, selectedId, onSelect }: SubscriptionColumnProps) {
    const [isLogOpen, setIsLogOpen] = useState(false)
    const [logSub, setLogSub] = useState<SubResponse | null>(null)
    const orderedSubs = useMemo(() => subs.slice().sort((a, b) => a.id - b.id), [subs])
    const selectedSub = useMemo(() => orderedSubs.find((sub) => sub.id === selectedId) ?? null, [orderedSubs, selectedId])
    const aliveRateBySubId = useMemo(() => {
        const stats = new Map<number, { alive: number; dead: number }>()
        for (const node of registrySummaryNodes) {
            const current = stats.get(node.sub_id) ?? { alive: 0, dead: 0 }
            if ((node.alive_status & NODE_STATUS.ALIVE) !== 0) {
                current.alive += 1
            } else {
                current.dead += 1
            }
            stats.set(node.sub_id, current)
        }
        const rateMap = new Map<number, string>()
        for (const [subId, item] of stats.entries()) {
            const total = item.alive + item.dead
            rateMap.set(subId, total > 0 ? `${Math.round((item.alive / total) * 100)}%` : "N/A")
        }
        return rateMap
    }, [registrySummaryNodes])

    if (isLoading) {
        return (
            <Card>
                <CardContent>
                    <InlineLoading message="加载订阅列表..." />
                </CardContent>
            </Card>
        )
    }

    if (error) {
        return (
            <Card>
                <CardContent>
                    <div className="text-center py-8 text-destructive">
                        加载失败: {error.message}
                    </div>
                </CardContent>
            </Card>
        )
    }

    if (orderedSubs.length === 0) {
        return (
            <Card>
                <CardContent>
                    <div className="text-center py-8 text-muted-foreground">
                        暂无订阅数据
                    </div>
                </CardContent>
            </Card>
        )
    }

    return (
        <div className="space-y-4 min-w-0 pr-1">
            {orderedSubs.map((sub) => (
                <Card
                    key={sub.id}
                    className={cn(
                        "group relative overflow-hidden transition-all duration-200 cursor-pointer shadow border-border p-0",
                        selectedId === sub.id ? "ring-1 ring-primary/40 border-primary" : "hover:border-primary/30"
                    )}
                    onClick={() => onSelect(sub)}
                >
                    <CardContent className="flex flex-col gap-2 p-4">
                        {/* Row 1: Name and Log Button */}
                        <div className="flex items-center justify-between gap-2">
                            <div className="text-[15px] font-bold text-foreground truncate flex-1 leading-tight">
                                {sub.name}
                            </div>
                            <Button
                                size="sm"
                                variant="ghost"
                                className="h-7 w-7 p-0 text-muted-foreground hover:bg-muted hover:text-primary shrink-0"
                                onClick={(event) => {
                                    event.stopPropagation()
                                    onSelect(sub)
                                    setLogSub(sub)
                                    setIsLogOpen(true)
                                }}
                                title="查看日志"
                            >
                                <FileText className="h-4 w-4" />
                            </Button>
                        </div>

                        {/* Row 2 & 3: Statistics Grid */}
                        <div className="grid grid-cols-[1fr_auto_110px] gap-y-2 gap-x-2 text-[11px] text-muted-foreground tabular-nums items-center">
                            {/* Line 2: Last Run & Alive Rate */}
                            <div className="flex items-center gap-1.5 truncate">
                                <Activity className="h-3 w-3 text-blue-500/80 shrink-0" />
                                <span className="truncate whitespace-nowrap">上次运行: {formatRelativeTime(sub.result?.last_run) || '从未运行'}</span>
                            </div>
                            <span className="opacity-20">|</span>
                            <div className="flex items-center gap-1.5 whitespace-nowrap">
                                <HeartPulse className="h-3 w-3 text-emerald-500/80 shrink-0" />
                                <span className="opacity-70">存活率:</span>
                                <span className={cn(
                                    "font-medium",
                                    (aliveRateBySubId.get(sub.id) || "0%").replace('%', '') === '0' ? "text-destructive" : "text-primary"
                                )}>
                                    {aliveRateBySubId.get(sub.id) ?? "N/A"}
                                </span>
                            </div>

                            {/* Line 3: Average Delay & Speeds */}
                            <div className="flex items-center gap-1.5 truncate">
                                <Zap className="h-3 w-3 text-amber-500/80 shrink-0" />
                                <span className="truncate whitespace-nowrap">平均延迟: {sub.info?.delay || 0}ms</span>
                            </div>
                            <span className="opacity-20">|</span>
                            <div className="flex items-center gap-2 whitespace-nowrap">
                                <div className="flex items-center gap-1">
                                    <ArrowUp className="h-2.5 w-2.5 text-emerald-500/80" />
                                    <span>{formatSpeed(sub.info?.speed_up || 0)}</span>
                                </div>
                                <div className="flex items-center gap-1">
                                    <ArrowDown className="h-2.5 w-2.5 text-blue-500/80" />
                                    <span>{formatSpeed(sub.info?.speed_down || 0)}</span>
                                </div>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            ))}

            <NodeLogDialog
                subscription={logSub ?? selectedSub}
                isOpen={isLogOpen}
                onOpenChange={(open) => {
                    setIsLogOpen(open)
                    if (!open) {
                        setLogSub(null)
                    }
                }}
            />
        </div>
    )
}
