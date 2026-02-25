import { useState } from "react"
import { Card, CardContent } from "@/src/components/ui/card"
import { Table, TableBody, TableHead, TableHeader, TableRow, TableCell } from "@/src/components/ui/table"
import { InlineLoading } from "@/src/components/ui/loading"
import { formatSpeed } from "@/src/components/features/sub/utils"
import { formatTime } from "@/src/utils"
import { getRiskClass, getRiskLabel } from "../utils"
import { NODE_STATUS } from "../constants"
import type { NodeResponse } from "@/src/types"
import { NodeDetail } from "./node-detail"

interface NodesTableProps {
    nodes: NodeResponse[]
    isLoading: boolean
    error: Error | null
}

export function NodesTable({ nodes, isLoading, error }: NodesTableProps) {
    const [detailNode, setDetailNode] = useState<NodeResponse | null>(null)
    const [isDetailOpen, setIsDetailOpen] = useState(false)

    if (isLoading) {
        return (
            <Card>
                <CardContent>
                    <InlineLoading message="加载节点列表..." />
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

    if (nodes.length === 0) {
        return (
            <Card>
                <CardContent>
                    <div className="text-center py-8 text-muted-foreground">
                        暂无节点数据
                    </div>
                </CardContent>
            </Card>
        )
    }

    const orderedNodes = nodes.slice().sort((a, b) => {
        const aAlive = (a.alive_status & NODE_STATUS.ALIVE) !== 0
        const bAlive = (b.alive_status & NODE_STATUS.ALIVE) !== 0
        if (aAlive === bAlive) return 0
        return aAlive ? -1 : 1
    })

    return (
        <>
            <Card className="min-w-0">
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>节点名称</TableHead>
                                <TableHead>类型</TableHead>
                                {orderedNodes.some((node) => node.alive_status & NODE_STATUS.ALIVE) ? (
                                    <>
                                        <TableHead>延迟</TableHead>
                                        <TableHead>上下行</TableHead>
                                        <TableHead>风险</TableHead>
                                        <TableHead>国家</TableHead>
                                        <TableHead>最近检测</TableHead>
                                    </>
                                ) : (
                                    <>
                                        <TableHead>初测状态</TableHead>
                                        <TableHead>失败原因</TableHead>
                                        <TableHead>最近检测</TableHead>
                                        <TableHead>来源</TableHead>
                                    </>
                                )}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {orderedNodes.map((node) => {
                                const isAlive = (node.alive_status & NODE_STATUS.ALIVE) !== 0
                                const lastCheck = formatTime(node.last_check_at) || '未知'
                                const lastSource = node.last_check_source || '未知'
                                return (
                                    <TableRow key={`${node.sub_id}-${node.unique_key}`}>
                                        <TableCell className="font-medium">
                                            <button
                                                type="button"
                                                className="cursor-pointer hover:text-blue-600 text-left"
                                                onClick={() => {
                                                    setDetailNode(node)
                                                    setIsDetailOpen(true)
                                                }}
                                            >
                                                {node.name || '未命名节点'}
                                            </button>
                                            {isAlive ? (
                                                node.reason ? (
                                                    <div className="text-xs text-muted-foreground">原因: {node.reason}</div>
                                                ) : null
                                            ) : null}
                                        </TableCell>
                                        <TableCell>{node.type || 'N/A'}</TableCell>
                                        {isAlive ? (
                                            <>
                                                <TableCell>{node.delay ? `${node.delay}ms` : 'N/A'}</TableCell>
                                                <TableCell className="text-xs text-muted-foreground">
                                                    ↑{formatSpeed(node.speed_up || 0)} ↓{formatSpeed(node.speed_down || 0)}
                                                </TableCell>
                                                <TableCell>
                                                    <span className={`text-xs font-medium ${getRiskClass(node.risk || 0)}`}>
                                                        {getRiskLabel(node.risk || 0)}
                                                    </span>
                                                </TableCell>
                                                <TableCell>{node.country || '未知'}</TableCell>
                                                <TableCell className="text-xs text-muted-foreground">{lastCheck}</TableCell>
                                            </>
                                        ) : (
                                            <>
                                                <TableCell>{node.init_status || 'unknown'}</TableCell>
                                                <TableCell className="text-xs text-muted-foreground">{node.reason || node.last_fail_reason || '未知'}</TableCell>
                                                <TableCell className="text-xs text-muted-foreground">{lastCheck}</TableCell>
                                                <TableCell className="text-xs text-muted-foreground">{lastSource}</TableCell>
                                            </>
                                        )}
                                    </TableRow>
                                )
                            })}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <NodeDetail
                node={detailNode}
                isOpen={isDetailOpen}
                onOpenChange={(open) => {
                    setIsDetailOpen(open)
                    if (!open) {
                        setDetailNode(null)
                    }
                }}
            />
        </>
    )
}
