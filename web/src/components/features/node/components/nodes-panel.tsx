import { Card, CardContent } from "@/src/components/ui/card"
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
        <div className="min-w-0">
            <NodesTable nodes={nodes} isLoading={isLoading} error={error} />
        </div>
    )
}
