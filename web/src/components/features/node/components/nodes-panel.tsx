import { Card, CardContent } from "@/src/components/ui/card"
import { Button } from "@/src/components/ui/button"
import { useState } from "react"
import { NodeLogDialog } from "./node-log-dialog"
import type { SubResponse } from "@/src/types"
import type { NodeResponse } from "@/src/types"
import { NodesTable } from "./nodes-table"

interface NodesPanelProps {
    selectedSub: SubResponse | null
    nodes: NodeResponse[]
    isLoading: boolean
    error: Error | null
}

export function NodesPanel({ selectedSub, nodes, isLoading, error }: NodesPanelProps) {
    const [isLogOpen, setIsLogOpen] = useState(false)

    if (!selectedSub) {
        return (
            <Card className="min-w-0">
                <CardContent>
                    <div className="text-center py-10 text-muted-foreground">
                        请选择左侧订阅以查看节点
                    </div>
                </CardContent>
            </Card>
        )
    }

    return (
        <div className="space-y-4 min-w-0">
            <div className="flex items-center justify-between">
                <div className="text-sm text-muted-foreground">
                    当前订阅: {selectedSub.name}
                </div>
                <Button size="sm" variant="outline" onClick={() => setIsLogOpen(true)}>
                    节点日志
                </Button>
            </div>
            <NodesTable nodes={nodes} isLoading={isLoading} error={error} />
            <NodeLogDialog
                subscription={selectedSub}
                isOpen={isLogOpen}
                onOpenChange={setIsLogOpen}
            />
        </div>
    )
}
