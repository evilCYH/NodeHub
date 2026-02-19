import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/src/components/ui/dialog"
import { Card, CardContent } from "@/src/components/ui/card"
import { Badge } from "@/src/components/ui/badge"
import { useSubRunLogs } from "@/src/lib/queries/sub-log-queries"
import type { SubResponse } from "@/src/types/sub"
import { Loader2 } from "lucide-react"

interface SubLogDialogProps {
    subscription: SubResponse | null
    isOpen: boolean
    onOpenChange: (open: boolean) => void
}

export function SubLogDialog({ subscription, isOpen, onOpenChange }: SubLogDialogProps) {
    const title = subscription ? `${subscription.name} - 订阅日志` : "订阅日志"

    const { data: subRunData, isLoading } = useSubRunLogs(subscription?.id ?? null, 10, isOpen, true)
    const runs = subRunData?.runs ?? []
    const events = subRunData?.events ?? []

    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent className="!max-w-[90vw] !sm:max-w-[1400px] !w-[1400px] max-h-[85vh] flex flex-col">
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
                    <div className="space-y-4">
                        <Card>
                            <CardContent className="py-4">
                                <div className="space-y-2 text-sm">
                                    {runs.map((run) => (
                                        <div key={run.id} className="flex items-center justify-between">
                                            <div className="text-muted-foreground">{run.created_at}</div>
                                            <div className="font-medium">{run.status}</div>
                                            <div className="text-muted-foreground">{run.duration_ms}ms</div>
                                            <div className="text-muted-foreground">raw:{run.raw_count}</div>
                                            <div className="text-muted-foreground">accepted:{run.accepted}</div>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>

                        <Card>
                            <CardContent className="py-4">
                                {events.length === 0 ? (
                                    <div className="text-center text-muted-foreground">暂无事件日志</div>
                                ) : (
                                    <div className="space-y-2 text-xs">
                                        {events.map((event) => (
                                            <div key={event.id} className="flex items-center gap-2">
                                                <Badge variant="outline">{event.step}</Badge>
                                                <span className="text-muted-foreground">
                                                    {event.run_time ?? event.created_at}
                                                </span>
                                                <span className="text-muted-foreground">{event.message}</span>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </div>
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
