import { useCallback } from "react"
import { Button } from "@/src/components/ui/button"
import { Card, CardContent } from "@/src/components/ui/card"
import { InlineLoading } from "@/src/components/ui/loading"
import { Switch } from "@/src/components/ui/switch"
import { RefreshCw, Edit, Trash2, FileText } from "lucide-react"
import { toast } from "sonner"
import { formatLastRunTime } from "@/src/utils"
import { StatusBadge } from "@/src/components/shared/status-badge"
import { formatSpeed } from "../utils"
import { useSubs, useDeleteSub, useRefreshSub, useUpdateSub } from "@/src/lib/queries/sub-queries"
import { useAlert } from "@/src/components/providers"
import type { SubResponse } from "@/src/types/sub"

interface SubscriptionListProps {
    onEdit: (subscription: SubResponse) => void
    onShowDetail: (subscription: SubResponse) => void
    onShowLogs: (subscription: SubResponse) => void
}

export function SubList({
    onEdit,
    onShowDetail,
    onShowLogs,
}: SubscriptionListProps) {
    const { data: subs = [], isLoading, error } = useSubs()
    const deleteSubMutation = useDeleteSub()
    const refreshSubMutation = useRefreshSub()
    const updateSubMutation = useUpdateSub()
    const { confirm } = useAlert()

    const handleDelete = useCallback(async (id: number, name: string) => {
        const confirmed = await confirm({
            title: '删除订阅',
            description: `确定要删除订阅 "${name}" 吗？`,
            confirmText: '删除',
            cancelText: '取消',
            variant: 'destructive'
        })

        if (confirmed) {
            try {
                await deleteSubMutation.mutateAsync(id)
                toast.success('删除成功')
            } catch (error) {
                console.error('Failed to delete subscription:', error)
                toast.error('删除失败')
            }
        }
    }, [confirm, deleteSubMutation])

    const handleToggleEnable = useCallback(async (subscription: SubResponse, enable: boolean) => {
        try {
            await updateSubMutation.mutateAsync({
                id: subscription.id,
                data: {
                    name: subscription.name,
                    tags: subscription.tags || [],
                    enable,
                    cron_expr: subscription.cron_expr,
                    config: {
                        url: subscription.config.url,
                        proxy: subscription.config.proxy || false,
                        timeout: subscription.config.timeout || 10,
                        protocol_filter_enable: subscription.config.protocol_filter_enable || false,
                        protocol_filter_mode: subscription.config.protocol_filter_mode || false,
                        protocol_filter: subscription.config.protocol_filter || [],
                    },
                },
            })
        } catch (error) {
            toast.error('更新订阅启用状态失败')
            console.error('Failed to update subscription enable state:', error)
        }
    }, [updateSubMutation])

    const handleRefresh = useCallback(async (id: number) => {
        try {
            await refreshSubMutation.mutateAsync(id)
            toast.info('订阅刷新已启动，测试中请稍候...')
        } catch (error) {
            console.error('Failed to refresh subscription:', error)
            toast.error('刷新失败')
        }
    }, [refreshSubMutation])

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

    if (subs.length === 0) {
        return (
            <Card>
                <CardContent>
                    <div className="text-center py-8 text-muted-foreground">
                        暂无订阅数据，点击上方按钮创建第一个订阅
                    </div>
                </CardContent>
            </Card>
        )
    }

    return (
        <div className="space-y-4">
            {subs.sort((a, b) => a.id - b.id).map((sub) => (
                <Card key={sub.id}>
                    <CardContent className="pl-10 pr-8 py-1">
                        <div className="grid gap-2 sm:grid-cols-[minmax(200px,320px)_10px_minmax(260px,1fr)_auto] sm:items-center sm:gap-x-2">
                            <div className="min-w-0">
                                <div
                                    className="text-sm font-medium cursor-pointer hover:text-blue-600 truncate"
                                    onClick={() => onShowDetail(sub)}
                                >
                                    {sub.name}
                                </div>
                                <div className="text-xs text-muted-foreground">{sub?.cron_expr || 'N/A'}</div>
                            </div>

                            <div className="flex items-center gap-3 sm:justify-self-center">
                                <Switch
                                    checked={sub.enable}
                                    onCheckedChange={(checked) => handleToggleEnable(sub, checked)}
                                    disabled={updateSubMutation.isPending}
                                />
                                <StatusBadge status={sub.status === 'running' ? 'running' : (sub.result?.last_status === 'error' || sub.status === 'pending' ? 'error' : (sub.enable ? sub.status : 'none'))} />
                            </div>

                            <div className="grid gap-1 text-xs sm:grid-cols-2 sm:gap-x-6 sm:pl-35 sm:pr-10">
                                <div className="space-y-1">
                                    <div>平均延迟: <span className="text-muted-foreground">{sub.info?.delay || 0}ms</span></div>
                                    <div className="text-muted-foreground">↑{formatSpeed(sub.info?.speed_up || 0)} ↓{formatSpeed(sub.info?.speed_down || 0)}</div>
                                </div>
                                <div className="space-y-1">
                                    <div>最后运行: <span className="text-muted-foreground">{formatLastRunTime(sub.result?.last_run)}</span></div>
                                    <div>执行时长: <span className="text-muted-foreground">{sub.result?.duration || 0}ms</span></div>
                                </div>
                            </div>

                            <div className="flex items-center gap-2 sm:justify-end">
                                <Button
                                    size="sm"
                                    variant="outline"
                                    onClick={() => onShowLogs(sub)}
                                >
                                    <FileText className="h-4 w-4" />
                                </Button>
                                {(() => {
                                    const isRefreshing = sub.status === 'running'
                                    return (
                                        <Button
                                            size="sm"
                                            variant="outline"
                                            onClick={() => handleRefresh(sub.id)}
                                            disabled={isRefreshing}
                                            className={isRefreshing ? 'opacity-50' : ''}
                                        >
                                            <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
                                        </Button>
                                    )
                                })()}
                                <Button
                                    size="sm"
                                    variant="outline"
                                    onClick={() => onEdit(sub)}
                                >
                                    <Edit className="h-4 w-4" />
                                </Button>
                                <Button
                                    size="sm"
                                    variant="outline"
                                    onClick={() => handleDelete(sub.id, sub.name)}
                                    disabled={deleteSubMutation.isPending && deleteSubMutation.variables === sub.id}
                                    className={deleteSubMutation.isPending && deleteSubMutation.variables === sub.id ? 'opacity-50' : ''}
                                >
                                    {deleteSubMutation.isPending && deleteSubMutation.variables === sub.id ? (
                                        <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                                    ) : (
                                        <Trash2 className="h-4 w-4" />
                                    )}
                                </Button>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            ))}
        </div>
    )
}
