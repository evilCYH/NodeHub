import { useState } from "react"
import { Button } from "@/src/components/ui/button"
import { Plus, Upload } from "lucide-react"
import { SubForm } from "./sub-form"
import { SubDetail } from "./sub-detail"
import { SubList } from "./sub-list"
import { SubLogDialog } from "./sub-log-dialog"
import { BatchSubForm } from "./batch-sub-form"
import { useSubs } from "@/src/lib/queries/sub-queries"
import type { SubResponse } from "@/src/types/sub"

export function SubPage() {
    const { data: subs = [] } = useSubs()
    const [detailSubscriptionId, setDetailSubscriptionId] = useState<number | null>(null)
    const [isDetailDialogOpen, setIsDetailDialogOpen] = useState(false)
    const [isFormDialogOpen, setIsFormDialogOpen] = useState(false)
    const [isBatchFormDialogOpen, setIsBatchFormDialogOpen] = useState(false)
    const [editingSubscription, setEditingSubscription] = useState<SubResponse | null>(null)
    const [logSubscription, setLogSubscription] = useState<SubResponse | null>(null)
    const [isLogDialogOpen, setIsLogDialogOpen] = useState(false)
    const detailSubscription = detailSubscriptionId == null
        ? null
        : subs.find(item => item.id === detailSubscriptionId) ?? null

    const handleEdit = (subscription: SubResponse) => {
        setEditingSubscription(subscription)
        setIsFormDialogOpen(true)
    }

    const handleCreate = () => {
        setEditingSubscription(null)
        setIsFormDialogOpen(true)
    }

    const handleBatchCreate = () => {
        setIsBatchFormDialogOpen(true)
    }

    const handleFormSuccess = () => {
        setIsFormDialogOpen(false)
        setEditingSubscription(null)
    }

    const handleBatchFormSuccess = () => {
        setIsBatchFormDialogOpen(false)
    }


    const showDetail = (subscription: SubResponse) => {
        setDetailSubscriptionId(subscription.id)
        setIsDetailDialogOpen(true)
    }

    const showLogs = (subscription: SubResponse) => {
        setLogSubscription(subscription)
        setIsLogDialogOpen(true)
    }

    return (
        <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
            <div className="flex items-center justify-between px-4 lg:px-6">
                <div>
                    <h1 className="text-2xl font-bold">订阅管理</h1>
                </div>

                <div className="flex gap-2">
                    <Button variant="outline" onClick={handleBatchCreate}>
                        <Upload className="h-4 w-4 mr-2" />
                        批量添加
                    </Button>
                    <Button onClick={handleCreate}>
                        <Plus className="h-4 w-4 mr-2" />
                        添加订阅
                    </Button>
                </div>
            </div>

            <SubForm
                initialData={editingSubscription ? {
                    name: editingSubscription.name,
                    tags: editingSubscription.tags || [],
                    enable: editingSubscription.enable,
                    cron_expr: editingSubscription.cron_expr,
                    config: {
                        url: editingSubscription.config.url || '',
                        proxy: editingSubscription.config.proxy || false,
                        timeout: editingSubscription.config.timeout || 10,
                    }
                } : undefined}
                formTitle={editingSubscription ? "编辑订阅" : "添加订阅"}
                isOpen={isFormDialogOpen}
                onClose={handleFormSuccess}
                editingSubId={editingSubscription?.id}
            />

            <BatchSubForm
                isOpen={isBatchFormDialogOpen}
                onClose={handleBatchFormSuccess}
            />

            <div className="px-4 lg:px-6">
                <SubList
                    onEdit={(sub) => handleEdit(sub)}
                    onShowDetail={showDetail}
                    onShowLogs={showLogs}
                />
            </div>

            <SubDetail
                subscription={detailSubscription}
                isOpen={isDetailDialogOpen}
                onOpenChange={(open) => {
                    setIsDetailDialogOpen(open)
                    if (!open) {
                        setDetailSubscriptionId(null)
                    }
                }}
            />

            <SubLogDialog
                subscription={logSubscription}
                isOpen={isLogDialogOpen}
                onOpenChange={setIsLogDialogOpen}
            />
        </div>
    )
}
