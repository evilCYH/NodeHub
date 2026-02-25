import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/src/components/ui/dialog"
import { InlineLoading } from "@/src/components/ui/loading"
import type { NodeResponse } from "@/src/types"
import { useNodeDetail } from "@/src/lib/queries/node-queries"

const COMMON_KEYS = [
    'name', 'server', 'port', 'type', 'uuid', 'password', 'cipher',
    'network', 'tls', 'udp', 'sni', 'alpn', 'host', 'path'
]

function getSortWeight(key: string): number {
    const idx = COMMON_KEYS.indexOf(key.toLowerCase())
    return idx === -1 ? 999 : idx
}

interface NodeDetailProps {
    node: NodeResponse | null
    isOpen: boolean
    onOpenChange: (open: boolean) => void
}

export function NodeDetail({ node, isOpen, onOpenChange }: NodeDetailProps) {
    const { data: detail, isLoading, error } = useNodeDetail(
        node?.sub_id ?? null,
        node?.unique_key_str ?? node?.unique_key ?? null,
        "registry",
        isOpen && node !== null
    )

    if (!node) return null

    const current = detail ?? node
    const rawConfig = detail?.raw ?? {}

    return (
        <Dialog open={isOpen} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>节点详情 - {current.name || "未命名节点"}</DialogTitle>
                </DialogHeader>

                {isLoading ? (
                    <InlineLoading message="加载节点详情..." />
                ) : error ? (
                    <div className="text-sm text-destructive">加载节点详情失败: {(error as Error).message}</div>
                ) : (
                    <div className="space-y-6">
                        <div>
                            <h3 className="font-semibold mb-3">基本信息</h3>
                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-3 text-sm px-1">
                                <div className="sm:col-span-2 flex flex-col sm:flex-row sm:items-start gap-1 sm:gap-4">
                                    <span className="text-muted-foreground shrink-0 sm:w-[80px]">唯一键:</span>
                                    <span className="break-all font-mono">{String(current.unique_key)}</span>
                                </div>
                                <div className="flex gap-4">
                                    <span className="text-muted-foreground shrink-0 sm:w-[80px]">初测状态:</span>
                                    <span>{current.init_status || "unknown"}</span>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h3 className="font-semibold mb-3">原始配置</h3>
                            <div className="border rounded-md overflow-hidden text-sm bg-muted/10">
                                {Object.keys(rawConfig).length === 0 ? (
                                    <div className="p-4 text-muted-foreground text-center">无原始配置数据</div>
                                ) : (
                                    <ul className="divide-y">
                                        {Object.entries(rawConfig)
                                            .sort(([a], [b]) => {
                                                const weightA = getSortWeight(a)
                                                const weightB = getSortWeight(b)
                                                if (weightA !== weightB) {
                                                    return weightA - weightB
                                                }
                                                return a.localeCompare(b)
                                            })
                                            .map(([key, value]) => (
                                                <li
                                                    key={key}
                                                    className="flex flex-col sm:flex-row sm:items-start gap-1 sm:gap-4 p-3 hover:bg-muted/50 transition-colors"
                                                >
                                                    <span className="text-muted-foreground shrink-0 sm:w-[90px] font-medium">{key}</span>
                                                    <span className="break-all font-mono">{formatRawValue(value)}</span>
                                                </li>
                                            ))}
                                    </ul>
                                )}
                            </div>
                        </div>
                    </div>
                )}
            </DialogContent>
        </Dialog>
    )
}

function formatRawValue(value: unknown): string {
    if (value === null || value === undefined) return "null"
    if (typeof value === "string") return value
    if (typeof value === "number" || typeof value === "boolean") return String(value)
    try {
        return JSON.stringify(value)
    } catch {
        return String(value)
    }
}
