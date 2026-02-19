import { useEffect, useState } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/src/components/ui/dialog"
import { Card, CardContent } from "@/src/components/ui/card"
import { Input } from "@/src/components/ui/input"
import { Badge } from "@/src/components/ui/badge"
import { Button } from "@/src/components/ui/button"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/src/components/ui/select"
import { useNodeLogs, useNodeUpdateLog } from "@/src/lib/queries/node-log-queries"
import type { SubResponse } from "@/src/types/sub"
import { Loader2, ChevronLeft, ChevronRight } from "lucide-react"

interface NodeLogDialogProps {
    subscription: SubResponse | null
    isOpen: boolean
    onOpenChange: (open: boolean) => void
}

const LEVEL_CONFIG = {
    info: { color: "bg-blue-500/10 text-blue-500 border-blue-500/20", label: "信息" },
    warn: { color: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20", label: "警告" },
    error: { color: "bg-red-500/10 text-red-500 border-red-500/20", label: "错误" },
}

const PAGE_SIZE = 50

export function NodeLogDialog({ subscription, isOpen, onOpenChange }: NodeLogDialogProps) {
    const [level, setLevel] = useState<string>("all")
    const [source, setSource] = useState<string>("all")
    const [keyword, setKeyword] = useState("")
    const [page, setPage] = useState(1)

    useEffect(() => {
        if (!isOpen) {
            setLevel("all")
            setSource("all")
            setKeyword("")
            setPage(1)
        }
    }, [isOpen])

    useEffect(() => {
        setPage(1)
    }, [level, source, keyword])

    const { data: statsData, isLoading: statsLoading } = useNodeUpdateLog(
        subscription?.id ?? null,
        1,
        isOpen
    )

    const {
        data: logData,
        isLoading,
    } = useNodeLogs({
        subId: subscription?.id ?? null,
        source: source === "all" ? undefined : source,
        level: level === "all" ? undefined : (level as 'info' | 'warn' | 'error'),
        keyword: keyword || undefined,
        page,
        pageSize: PAGE_SIZE,
        enabled: isOpen,
    })

    const logs = logData?.list ?? []
    const total = logData?.total ?? 0
    const totalPages = Math.ceil(total / PAGE_SIZE)
    const latest = statsData?.latest
    const title = subscription ? `${subscription.name} - 节点日志` : "节点日志"

    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent className="!max-w-[90vw] !sm:max-w-[1400px] !w-[1400px] max-h-[85vh] flex flex-col">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                </DialogHeader>

                {statsLoading ? (
                    <Card>
                        <CardContent className="py-4">
                            <div className="flex items-center justify-center gap-2 text-muted-foreground">
                                <Loader2 className="h-4 w-4 animate-spin" />
                                加载统计...
                            </div>
                        </CardContent>
                    </Card>
                ) : latest ? (
                    <Card>
                        <CardContent className="py-4">
                            <div className="grid grid-cols-3 md:grid-cols-6 gap-4 text-sm">
                                <div className="space-y-1">
                                    <div className="text-muted-foreground text-xs">原始节点</div>
                                    <div className="font-medium">{latest.raw_count}</div>
                                </div>
                                <div className="space-y-1">
                                    <div className="text-muted-foreground text-xs">候选节点</div>
                                    <div className="font-medium">{latest.candidate}</div>
                                </div>
                                <div className="space-y-1">
                                    <div className="text-muted-foreground text-xs">重复节点</div>
                                    <div className="font-medium text-yellow-600">{latest.duplicate}</div>
                                </div>
                                <div className="space-y-1">
                                    <div className="text-muted-foreground text-xs">无效节点</div>
                                    <div className="font-medium text-red-600">{latest.invalid}</div>
                                </div>
                                <div className="space-y-1">
                                    <div className="text-muted-foreground text-xs">初测失败</div>
                                    <div className="font-medium text-orange-600">{latest.test_failed}</div>
                                </div>
                                <div className="space-y-1">
                                    <div className="text-muted-foreground text-xs">入库节点</div>
                                    <div className="font-medium text-green-600">{latest.merged}</div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ) : null}

                <div className="flex gap-3 items-center">
                    <Select value={source} onValueChange={setSource}>
                        <SelectTrigger className="w-32">
                            <SelectValue placeholder="来源" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">全部来源</SelectItem>
                            <SelectItem value="init">初测</SelectItem>
                            <SelectItem value="check">检测</SelectItem>
                        </SelectContent>
                    </Select>

                    <Select value={level} onValueChange={setLevel}>
                        <SelectTrigger className="w-32">
                            <SelectValue placeholder="级别" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">全部级别</SelectItem>
                            <SelectItem value="info">信息</SelectItem>
                            <SelectItem value="warn">警告</SelectItem>
                            <SelectItem value="error">错误</SelectItem>
                        </SelectContent>
                    </Select>

                    <Input
                        placeholder="搜索节点名或日志内容..."
                        value={keyword}
                        onChange={(e) => setKeyword(e.target.value)}
                        className="flex-1"
                    />
                </div>

                <Card className="flex-1 overflow-hidden">
                    <CardContent className="p-0">
                        <div className="border-b bg-muted/50">
                            <div className="grid grid-cols-[140px_90px_80px_1fr] gap-3 px-4 py-3 font-medium text-sm">
                                <div>时间</div>
                                <div>级别</div>
                                <div>来源</div>
                                <div>日志内容</div>
                            </div>
                        </div>

                        <div className="overflow-y-auto max-h-[45vh]">
                            {isLoading ? (
                                <div className="py-8 flex items-center justify-center gap-2 text-muted-foreground">
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                    加载日志...
                                </div>
                            ) : logs.length === 0 ? (
                                <div className="py-8 text-center text-muted-foreground">暂无日志</div>
                            ) : (
                                <div className="divide-y">
                                    {logs.map((log) => {
                                        const config = LEVEL_CONFIG[log.level as keyof typeof LEVEL_CONFIG]
                                        return (
                                            <div
                                                key={log.id}
                                                className="grid grid-cols-[140px_90px_80px_1fr] gap-3 px-4 py-3 text-sm hover:bg-muted/50"
                                            >
                                                <div className="text-muted-foreground font-mono text-xs flex items-center">
                                                    {formatTime(log.created_at)}
                                                </div>
                                                <div className="flex items-center">
                                                    <Badge variant="outline" className={config?.color}>
                                                        {config?.label ?? log.level}
                                                    </Badge>
                                                </div>
                                                <div className="text-xs text-muted-foreground flex items-center">{log.source}</div>
                                                <div className="space-y-1 min-w-0">
                                                    <div className="font-medium text-xs text-muted-foreground truncate">
                                                        {log.node_name}
                                                    </div>
                                                    <div className="break-all text-sm">{log.message}</div>
                                                </div>
                                            </div>
                                        )
                                    })}
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>

                {totalPages > 1 && (
                    <div className="flex items-center justify-between py-2">
                        <div className="text-sm text-muted-foreground">
                            共 {total} 条记录，{totalPages} 页
                        </div>
                        <div className="flex items-center gap-2">
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setPage(p => Math.max(1, p - 1))}
                                disabled={page <= 1 || isLoading}
                            >
                                <ChevronLeft className="h-4 w-4" />
                            </Button>
                            <div className="flex items-center gap-1">
                                {generatePageNumbers(page, totalPages).map((p, i) => (
                                    p === '...' ? (
                                        <span key={`ellipsis-${i}`} className="px-2 text-muted-foreground">...</span>
                                    ) : (
                                        <Button
                                            key={p}
                                            variant={page === p ? "default" : "outline"}
                                            size="sm"
                                            onClick={() => setPage(p as number)}
                                            disabled={isLoading}
                                            className="min-w-[32px]"
                                        >
                                            {p}
                                        </Button>
                                    )
                                ))}
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                                disabled={page >= totalPages || isLoading}
                            >
                                <ChevronRight className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>
                )}
            </DialogContent>
        </Dialog>
    )
}

function formatTime(iso: string): string {
    const d = new Date(iso)
    return (
        `${(d.getMonth() + 1).toString().padStart(2, "0")}-${d.getDate().toString().padStart(2, "0")} ` +
        `${d.getHours().toString().padStart(2, "0")}:${d.getMinutes().toString().padStart(2, "0")}:${d.getSeconds().toString().padStart(2, "0")}`
    )
}

function generatePageNumbers(current: number, total: number): (number | string)[] {
    const pages: (number | string)[] = []

    if (total <= 7) {
        for (let i = 1; i <= total; i++) {
            pages.push(i)
        }
    } else {
        if (current <= 4) {
            for (let i = 1; i <= 5; i++) {
                pages.push(i)
            }
            pages.push('...')
            pages.push(total)
        } else if (current >= total - 3) {
            pages.push(1)
            pages.push('...')
            for (let i = total - 4; i <= total; i++) {
                pages.push(i)
            }
        } else {
            pages.push(1)
            pages.push('...')
            for (let i = current - 1; i <= current + 1; i++) {
                pages.push(i)
            }
            pages.push('...')
            pages.push(total)
        }
    }

    return pages
}
