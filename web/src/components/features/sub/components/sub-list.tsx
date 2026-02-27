import { useCallback, useEffect, useMemo, useState } from "react"
import {
    DndContext,
    type DragEndEvent,
    PointerSensor,
    useSensor,
    useSensors,
    DragOverlay,
    defaultDropAnimationSideEffects,
    MeasuringStrategy,
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
import { RefreshCw, Edit, Trash2, FileText, GripVertical, Zap, Activity } from "lucide-react"
import { toast } from "sonner"
import { cn, formatExpireStatus, formatTrafficSummary, hasSubscriptionInfo, formatRelativeTime } from "@/src/utils"
import { StatusBadge } from "@/src/components/shared/status-badge"
import { useSubs, useDeleteSub, useRefreshSub, useUpdateSub, useUpdateSubOrder } from "@/src/lib/queries/sub-queries"
import { useAlert } from "@/src/components/providers"
import type { SubOrderItem, SubResponse } from "@/src/types/sub"

interface SubscriptionListProps {
    onEdit: (subscription: SubResponse) => void
    onShowDetail: (subscription: SubResponse) => void
    onShowLogs: (subscription: SubResponse) => void
}

interface SubCardInnerProps {
    sub: SubResponse
    isUpdating?: boolean
    isDeleting?: boolean
    onEdit?: (s: SubResponse) => void
    onShowDetail?: (s: SubResponse) => void
    onShowLogs?: (s: SubResponse) => void
    onToggleEnable?: (s: SubResponse, e: boolean) => void
    onRefresh?: (id: number) => void
    onDelete?: (id: number, n: string) => void
    isOverlay?: boolean
}

function SubCardInner({
    sub,
    isUpdating,
    isDeleting,
    onEdit,
    onShowDetail,
    onShowLogs,
    onToggleEnable,
    onRefresh,
    onDelete,
    isOverlay
}: SubCardInnerProps) {
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
        <CardContent className="flex p-0">
            {/* Drag Handle */}
            <div
                className={cn(
                    "flex w-6 items-center justify-center transition-colors shrink-0 relative",
                    isOverlay ? "bg-muted/40 text-muted-foreground" : "bg-muted/20 text-muted-foreground/30 hover:bg-muted hover:text-foreground cursor-grab active:cursor-grabbing"
                )}
            >
                <GripVertical className="h-3.5 w-3.5" />
                <div className="absolute right-0 top-4 bottom-4 w-[1px] bg-foreground/5" />
            </div>

            <div className="flex min-w-0 flex-1 flex-col gap-2 pl-3 pr-4 py-5">
                {/* Header: Name (Line 1) */}
                <div className="flex items-center justify-between gap-2">
                    <div
                        className="truncate text-[15px] font-bold text-foreground cursor-pointer hover:text-primary transition-colors flex-1"
                        onClick={() => onShowDetail?.(sub)}
                        title={sub.name}
                    >
                        {sub.name}
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                        <StatusBadge status={sub.status === 'running'
                            ? 'running'
                            : sub.status === 'pending'
                                ? 'pending'
                                : (sub.result?.last_status === 'error'
                                    ? 'error'
                                    : (sub.enable ? sub.status : 'none'))} />
                        <Switch
                            checked={sub.enable}
                            onCheckedChange={(checked) => onToggleEnable?.(sub, checked)}
                            disabled={isUpdating || isOverlay}
                            className="scale-75"
                        />
                    </div>
                </div>

                {/* Middle Section: Last Run and Delay */}
                <div className="flex items-center justify-between gap-2">
                    <div className="flex flex-col gap-2 min-w-0 flex-1">
                        <div className="flex items-center gap-2 text-[11px] text-muted-foreground truncate">
                            <Activity className="h-3 w-3 text-blue-500/80 shrink-0" />
                            <span className="truncate">上次运行: {formatRelativeTime(sub.result?.last_run) || '从未运行'}</span>
                        </div>
                        <div className="flex items-center gap-2 text-[11px] text-muted-foreground truncate">
                            <Zap className="h-3 w-3 text-amber-500/80 shrink-0" />
                            <span className="truncate">平均延迟: {sub.info?.delay || 0}ms</span>
                        </div>
                    </div>

                    <div className="flex items-center gap-0 shrink-0 self-center">
                        <Button
                            size="sm" variant="ghost" className="h-7 w-7 p-0 text-muted-foreground hover:text-primary"
                            onClick={() => onShowLogs?.(sub)} disabled={isOverlay}
                        >
                            <FileText className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                            size="sm" variant="ghost" className={cn(
                                "h-7 w-7 p-0 text-muted-foreground hover:text-primary",
                                sub.status === 'running' && "text-primary"
                            )}
                            onClick={() => onRefresh?.(sub.id)} disabled={sub.status === 'running' || isOverlay}
                        >
                            <RefreshCw className={cn("h-3.5 w-3.5", sub.status === 'running' && "animate-spin")} />
                        </Button>
                        <Button
                            size="sm" variant="ghost" className="h-7 w-7 p-0 text-muted-foreground hover:text-primary"
                            onClick={() => onEdit?.(sub)} disabled={isOverlay}
                        >
                            <Edit className="h-3.5 w-3.5" />
                        </Button>
                        <Button
                            size="sm" variant="ghost" className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
                            onClick={() => onDelete?.(sub.id, sub.name)} disabled={isDeleting || isOverlay}
                        >
                            {isDeleting ? (
                                <div className="h-3 w-3 animate-spin rounded-full border-2 border-current border-t-transparent" />
                            ) : (
                                <Trash2 className="h-3.5 w-3.5" />
                            )}
                        </Button>
                    </div>
                </div>

                {/* Bottom Section: Traffic and Progress Bar */}
                <div className="flex flex-col gap-1">
                    <div className="flex items-center justify-between text-[11px] text-muted-foreground/80 px-0.5 tabular-nums">
                        <span className={cn("font-medium", trafficClass)}>
                            {traffic.usedText} / {traffic.totalText}
                        </span>
                        <span className={cn("font-medium text-right", expireClass)}>{expire.label}</span>
                    </div>
                    {hasInfo && traffic.usagePercent !== null && (
                        <div className="h-1 w-full bg-zinc-200 dark:bg-zinc-800 rounded-full overflow-hidden">
                            <div
                                className={cn(
                                    "h-full bg-primary transition-all duration-500",
                                    traffic.isOverLimit ? "bg-red-500" : (traffic.usagePercent > 80 ? "bg-amber-500" : "")
                                )}
                                style={{ width: `${Math.min(100, traffic.usagePercent)}%` }}
                            />
                        </div>
                    )}
                </div>
            </div>
        </CardContent>
    )
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
        zIndex: isDragging ? 50 : undefined,
    }

    return (
        <Card
            ref={setNodeRef}
            style={style}
            className={cn(
                "group relative overflow-hidden border-muted/50 p-0 hover:border-primary/30",
                // Only apply our own transitions when not being moved by dnd-kit
                !transform && "transition-all duration-200 hover:shadow-md",
                isDragging ? "opacity-30 invisible" : "opacity-100",
            )}
        >
            <div ref={setActivatorNodeRef} {...attributes} {...listeners} className="h-full">
                <SubCardInner
                    sub={sub}
                    isUpdating={isUpdating}
                    isDeleting={isDeleting}
                    onEdit={onEdit}
                    onShowDetail={onShowDetail}
                    onShowLogs={onShowLogs}
                    onToggleEnable={onToggleEnable}
                    onRefresh={onRefresh}
                    onDelete={onDelete}
                />
            </div>
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
    const sensors = useSensors(
        useSensor(PointerSensor, {
            activationConstraint: {
                distance: 8,
            },
        })
    )

    const [items, setItems] = useState<SubResponse[]>([])
    const [activeId, setActiveId] = useState<number | null>(null)
    const isDragging = activeId !== null

    useEffect(() => {
        if (!activeId && !updateSubOrderMutation.isPending) {
            setItems(subs)
        }
    }, [subs, activeId, updateSubOrderMutation.isPending])

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

    const handleDragStart = (event: { active: { id: any } }) => {
        setActiveId(Number(event.active.id))
    }

    const handleDragEnd = useCallback(async (event: DragEndEvent) => {
        const { active, over } = event

        if (!over || active.id === over.id) {
            setActiveId(null)
            return
        }

        const oldIndex = items.findIndex(item => item.id === Number(active.id))
        const newIndex = items.findIndex(item => item.id === Number(over.id))

        if (oldIndex < 0 || newIndex < 0) {
            setActiveId(null)
            return
        }

        const previousItems = items
        const newItems = arrayMove(items, oldIndex, newIndex)

        // 1. First update local UI state to the new position
        setItems(newItems)
        // 2. Then clear the active drag ID to allow the drop animation to finish at the new spot
        setActiveId(null)

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
            measuring={{
                droppable: {
                    strategy: MeasuringStrategy.Always,
                },
            }}
            onDragStart={handleDragStart}
            onDragCancel={() => setActiveId(null)}
            onDragEnd={handleDragEnd}
        >
            <SortableContext items={itemIds} strategy={rectSortingStrategy}>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
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

            <DragOverlay adjustScale={true} dropAnimation={{
                sideEffects: defaultDropAnimationSideEffects({
                    styles: {
                        active: {
                            opacity: '0.5',
                        },
                    },
                }),
            }}>
                {activeId ? (
                    <div className="w-[calc(100vw-2rem)] md:w-[350px] max-w-full lg:w-[400px]">
                        {(() => {
                            const sub = items.find(i => i.id === activeId);
                            if (!sub) return null;
                            return (
                                <Card className="relative overflow-hidden border-primary shadow-2xl ring-2 ring-primary/20 p-0 scale-[1.02] transition-transform duration-200 cursor-grabbing bg-card/95 backdrop-blur-sm">
                                    <SubCardInner sub={sub} isOverlay />
                                </Card>
                            )
                        })()}
                    </div>
                ) : null}
            </DragOverlay>
        </DndContext>
    )
}
