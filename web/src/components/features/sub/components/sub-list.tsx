import { useCallback, useEffect, useMemo, useState } from "react"
import {
    DndContext,
    type DragEndEvent,
    PointerSensor,
    useSensor,
    useSensors,
} from "@dnd-kit/core"
import {
    SortableContext,
    arrayMove,
    rectSortingStrategy,
    useSortable,
} from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import { Button } from "@/src/components/ui/button"
import { Card, CardContent } from "@/src/components/ui/card"
import { InlineLoading } from "@/src/components/ui/loading"
import { Switch } from "@/src/components/ui/switch"
import { RefreshCw, Edit, Trash2, FileText, GripVertical } from "lucide-react"
import { toast } from "sonner"
import { cn, formatExpireStatus, formatLastRunTime, formatTrafficSummary, hasSubscriptionInfo } from "@/src/utils"
import { StatusBadge } from "@/src/components/shared/status-badge"
import { formatSpeed } from "../utils"
import { useSubs, useDeleteSub, useRefreshSub, useUpdateSub, useUpdateSubOrder } from "@/src/lib/queries/sub-queries"
import { useAlert } from "@/src/components/providers"
import type { SubOrderItem, SubResponse } from "@/src/types/sub"

interface SubscriptionListProps {
    onEdit: (subscription: SubResponse) => void
    onShowDetail: (subscription: SubResponse) => void
    onShowLogs: (subscription: SubResponse) => void
}

interface SortableSubCardProps {
    sub: SubResponse
    onEdit: (subscription: SubResponse) => void
    onShowDetail: (subscription: SubResponse) => void
    onShowLogs: (subscription: SubResponse) => void
    onToggleEnable: (subscription: SubResponse, enable: boolean) => void
    onRefresh: (id: number) => void
    onDelete: (id: number, name: string) => void
    isUpdating: boolean
    isDeleting: boolean
}

function SortableSubCard({
    sub,
    onEdit,
    onShowDetail,
    onShowLogs,
    onToggleEnable,
    onRefresh,
    onDelete,
    isUpdating,
    isDeleting,
}: SortableSubCardProps) {
    const {
        attributes,
        listeners,
        setNodeRef,
        setActivatorNodeRef,
        transform,
        transition,
        isDragging,
    } = useSortable({ id: sub.id })

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
    }

    const hasInfo = hasSubscriptionInfo(sub.info_updated_at)
    const traffic = formatTrafficSummary(sub.upload, sub.download, sub.total)
    const expire = formatExpireStatus(sub.expire)
    const trafficClass = traffic.isOverLimit ? 'text-red-600' : 'text-green-600'
    const expireClass = expire.state === 'expired'
        ? 'text-red-600'
        : expire.state === 'warning'
            ? 'text-yellow-600'
            : expire.state === 'permanent'
                ? 'text-green-600'
                : 'text-muted-foreground'

    return (
        <Card
            ref={setNodeRef}
            style={style}
            className={cn("gap-0 py-0", isDragging && "ring-2 ring-primary")}
        >
            <CardContent className="flex p-0">
                <button
                    ref={setActivatorNodeRef}
                    type="button"
                    aria-label="拖拽排序"
                    className="flex w-10 items-center justify-center border-r bg-muted/30 text-muted-foreground transition hover:bg-muted hover:text-foreground cursor-grab active:cursor-grabbing"
                    {...attributes}
                    {...listeners}
                >
                    <GripVertical className="h-4 w-4" />
                </button>

                <div className="flex min-w-0 flex-1 flex-col gap-2 p-3">
                    <div className="flex items-start justify-between gap-2">
                        <div className="min-w-0">
                            <div
                                className="truncate text-sm font-medium cursor-pointer hover:text-blue-600"
                                onClick={() => onShowDetail(sub)}
                            >
                                {sub.name}
                            </div>
                            <div className="text-xs text-muted-foreground">{sub.cron_expr || 'N/A'}</div>
                        </div>

                        <div className="flex items-center gap-2">
                            <Switch
                                checked={sub.enable}
                                onCheckedChange={(checked) => onToggleEnable(sub, checked)}
                                disabled={isUpdating}
                            />
                            <StatusBadge status={sub.status === 'running'
                                ? 'running'
                                : sub.status === 'pending'
                                    ? 'pending'
                                    : (sub.result?.last_status === 'error'
                                        ? 'error'
                                        : (sub.enable ? sub.status : 'none'))} />
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
                        <div>最后运行: <span className="text-muted-foreground">{formatLastRunTime(sub.result?.last_run)}</span></div>
                        <div>执行时长: <span className="text-muted-foreground">{sub.result?.duration || 0}ms</span></div>
                        <div>平均延迟: <span className="text-muted-foreground">{sub.info?.delay || 0}ms</span></div>
                        <div className="text-muted-foreground">↑{formatSpeed(sub.info?.speed_up || 0)} ↓{formatSpeed(sub.info?.speed_down || 0)}</div>
                        {!hasInfo ? (
                            <>
                                <div>流量: <span className="text-muted-foreground">未知</span></div>
                                <div>到期: <span className="text-muted-foreground">未知</span></div>
                            </>
                        ) : (
                            <>
                                <div>
                                    流量:
                                    <span className={`ml-1 ${trafficClass}`}>
                                        {traffic.usedText} / {traffic.totalText}
                                    </span>
                                </div>
                                <div>到期: <span className={expireClass}>{expire.label}</span></div>
                            </>
                        )}
                    </div>

                    <div className="flex items-center gap-2 justify-end">
                        <Button
                            size="sm"
                            variant="outline"
                            onClick={() => onShowLogs(sub)}
                        >
                            <FileText className="h-4 w-4" />
                        </Button>
                        <Button
                            size="sm"
                            variant="outline"
                            onClick={() => onRefresh(sub.id)}
                            disabled={sub.status === 'running'}
                            className={sub.status === 'running' ? 'opacity-50' : ''}
                        >
                            <RefreshCw className={`h-4 w-4 ${sub.status === 'running' ? 'animate-spin' : ''}`} />
                        </Button>
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
                            onClick={() => onDelete(sub.id, sub.name)}
                            disabled={isDeleting}
                            className={isDeleting ? 'opacity-50' : ''}
                        >
                            {isDeleting ? (
                                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                            ) : (
                                <Trash2 className="h-4 w-4" />
                            )}
                        </Button>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
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
    const updateSubOrderMutation = useUpdateSubOrder()
    const { confirm } = useAlert()
    const sensors = useSensors(useSensor(PointerSensor))

    const [items, setItems] = useState<SubResponse[]>([])
    const [isDragging, setIsDragging] = useState(false)

    useEffect(() => {
        if (!isDragging) setItems(subs)
    }, [subs, isDragging])

    const itemIds = useMemo(() => items.map(item => item.id), [items])

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

    const handleDragEnd = useCallback(async (event: DragEndEvent) => {
        const { active, over } = event
        if (!over || active.id === over.id) {
            setIsDragging(false)
            return
        }
        if (updateSubOrderMutation.isPending) {
            setIsDragging(false)
            return
        }

        const oldIndex = items.findIndex(item => item.id === Number(active.id))
        const newIndex = items.findIndex(item => item.id === Number(over.id))
        if (oldIndex < 0 || newIndex < 0) {
            setIsDragging(false)
            return
        }

        const previousItems = items
        const newItems = arrayMove(items, oldIndex, newIndex)
        setItems(newItems)

        const orders: SubOrderItem[] = newItems.map((item, index) => ({
            id: item.id,
            sort_order: index,
        }))

        try {
            await updateSubOrderMutation.mutateAsync(orders)
        } catch (error) {
            setItems(previousItems)
            toast.error('顺序修改失败')
            console.error('Failed to update subscription order:', error)
        } finally {
            setIsDragging(false)
        }
    }, [items, updateSubOrderMutation])

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

    if (items.length === 0) {
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
        <DndContext
            sensors={sensors}
            onDragStart={() => setIsDragging(true)}
            onDragCancel={() => setIsDragging(false)}
            onDragEnd={handleDragEnd}
        >
            <SortableContext items={itemIds} strategy={rectSortingStrategy}>
                <div className="grid grid-cols-2 gap-4">
                    {items.map((sub) => (
                        <SortableSubCard
                            key={sub.id}
                            sub={sub}
                            onEdit={onEdit}
                            onShowDetail={onShowDetail}
                            onShowLogs={onShowLogs}
                            onToggleEnable={handleToggleEnable}
                            onRefresh={handleRefresh}
                            onDelete={handleDelete}
                            isUpdating={updateSubMutation.isPending}
                            isDeleting={deleteSubMutation.isPending && deleteSubMutation.variables === sub.id}
                        />
                    ))}
                </div>
            </SortableContext>
        </DndContext>
    )
}
