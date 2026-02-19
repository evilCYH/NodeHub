import { useMemo, useState } from "react"
import { Card, CardContent } from "@/src/components/ui/card"
import { Button } from "@/src/components/ui/button"
import { InlineLoading } from "@/src/components/ui/loading"
import { FileText } from "lucide-react"
import { cn, formatLastRunTime } from "@/src/utils"
import { formatSpeed } from "@/src/components/features/sub/utils"
import { NodeLogDialog } from "./node-log-dialog"
import type { SubResponse } from "@/src/types"

interface SubscriptionColumnProps {
    subs: SubResponse[]
    isLoading: boolean
    error: Error | null
    selectedId: number | null
    onSelect: (sub: SubResponse) => void
}

export function SubscriptionColumn({ subs, isLoading, error, selectedId, onSelect }: SubscriptionColumnProps) {
    const [isLogOpen, setIsLogOpen] = useState(false)
    const [logSub, setLogSub] = useState<SubResponse | null>(null)
    const orderedSubs = useMemo(() => subs.slice().sort((a, b) => a.id - b.id), [subs])
    const selectedSub = useMemo(() => orderedSubs.find((sub) => sub.id === selectedId) ?? null, [orderedSubs, selectedId])

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
        <div className="space-y-3 min-w-0">
            {orderedSubs.map((sub) => (
                <Card
                    key={sub.id}
                    className={cn("py-4 gap-0", selectedId === sub.id && "ring-2 ring-primary")}
                    onClick={() => onSelect(sub)}
                >
                    <CardContent className="px-4 py-3">
                        <div className="flex items-start justify-between gap-3">
                            <div className="min-w-0 flex-1">
                                <div className="text-sm font-medium cursor-pointer hover:text-blue-600 truncate">
                                    {sub.name}
                                </div>
                                <div className="mt-2 text-xs text-muted-foreground">
                                    最后运行: {formatLastRunTime(sub.result?.last_run)}
                                </div>
                                <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                                    <div>平均延迟: <span className="text-foreground">{sub.info?.delay || 0}ms</span></div>
                                    <div>平均上下行: <span className="text-foreground">↑{formatSpeed(sub.info?.speed_up || 0)} ↓{formatSpeed(sub.info?.speed_down || 0)}</span></div>
                                </div>
                            </div>
                            <div className="flex flex-col items-end justify-between self-stretch text-xs text-muted-foreground whitespace-nowrap">
                                <div>
                                    存活率: {
                                        sub.result?.raw_count
                                            ? `${Math.round(((sub.info?.count || 0) / sub.result.raw_count) * 100)}%`
                                            : 'N/A'
                                    }
                                </div>
                                <Button
                                    size="sm"
                                    variant="outline"
                                    onClick={(event) => {
                                        event.stopPropagation()
                                        onSelect(sub)
                                        setLogSub(sub)
                                        setIsLogOpen(true)
                                    }}
                                >
                                    <FileText className="h-4 w-4" />
                                </Button>
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
