import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps, Node } from '@xyflow/react';
import { Database, Server, Smartphone, Globe } from 'lucide-react';
import { clsx } from 'clsx';
import type { Node as ApiNode } from '../api';

// Create a composite type for our ReactFlow node
type ServiceNodeData = ApiNode & Record<string, unknown>;
type ServiceNode = Node<ServiceNodeData>;

const ServiceNode = ({ data, selected }: NodeProps<ServiceNode>) => {
    // Determine icon based on service name (heuristic)
    let Icon = Server;
    if (data.service_name.includes('db') || data.service_name.includes('redis')) Icon = Database;
    if (data.service_name.includes('mobile') || data.service_name.includes('ios')) Icon = Smartphone;
    if (data.service_name.includes('frontend') || data.service_name.includes('web')) Icon = Globe;

    // Health styling
    const isDown = data.health_status === 2;
    const isDegraded = data.health_status === 1;

    const borderColor = isDown ? 'border-red-500' : isDegraded ? 'border-amber-500' : 'border-emerald-500';
    const glowColor = isDown ? 'shadow-red-500/50' : isDegraded ? 'shadow-amber-500/50' : 'shadow-emerald-500/20';
    const iconColor = isDown ? 'text-red-400' : isDegraded ? 'text-amber-400' : 'text-emerald-400';

    return (
        <div className={clsx(
            "relative flex items-center gap-3 px-4 py-3 rounded-xl bg-slate-900 border-2 transition-all duration-300 min-w-[180px]",
            borderColor,
            selected ? `shadow-[0_0_20px_0px] ${glowColor} scale-105` : "shadow-md hover:shadow-lg",
        )}>
            {/* Input Handle */}
            <Handle type="target" position={Position.Top} className="!bg-slate-500 !w-3 !h-3" />

            <div className={clsx("p-2 rounded-lg bg-slate-800", iconColor)}>
                <Icon size={20} />
            </div>

            <div className="flex flex-col">
                <span className="text-sm font-semibold text-slate-100">{data.service_name}</span>
                <span className={clsx("text-xs font-medium", iconColor)}>
                    {isDown ? 'DOWN' : isDegraded ? 'DEGRADED' : 'HEALTHY'}
                </span>
            </div>

            {/* Amplification Badge */}
            {data.amplification_score > 0 && (
                <div className="absolute -top-3 -right-2 bg-purple-600 text-white text-[10px] font-bold px-2 py-0.5 rounded-full shadow-lg border border-purple-400">
                    {data.amplification_score.toFixed(1)}x
                </div>
            )}

            {/* Downstream Failure Badge */}
            {data.downstream_failures > 0 && (
                <div className="absolute -bottom-3 -right-2 bg-red-600 text-white text-[10px] font-bold px-2 py-0.5 rounded-full shadow-lg border border-red-400">
                    {data.downstream_failures} Fails
                </div>
            )}

            {/* Output Handle */}
            <Handle type="source" position={Position.Bottom} className="!bg-slate-500 !w-3 !h-3" />
        </div>
    );
};

export default memo(ServiceNode);
