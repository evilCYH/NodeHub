export function getRiskLabel(risk: number): string {
    if (risk >= 7) return '高风险'
    if (risk >= 4) return '中风险'
    return '低风险'
}

export function getRiskClass(risk: number): string {
    if (risk >= 7) return 'text-red-600'
    if (risk >= 4) return 'text-yellow-600'
    return 'text-green-600'
}
