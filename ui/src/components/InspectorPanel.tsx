import { X, Activity, AlertTriangle, CheckCircle, Info } from 'lucide-react';
import { clsx } from 'clsx';
import type { Node as FlowNode, Edge as FlowEdge } from '@xyflow/react';
import type { Node as ApiNode } from '../api';

interface InspectorPanelProps {
    selectedNode: FlowNode | null;
    selectedEdge: FlowEdge | null;
    onClose: () => void;
}

export default function InspectorPanel({ selectedNode, selectedEdge, onClose }: InspectorPanelProps) {
    if (!selectedNode && !selectedEdge) return null;

    return (
        <div className="absolute right-4 top-4 bottom-4 w-96 bg-slate-900/90 backdrop-blur-xl border border-slate-700 shadow-2xl rounded-2xl flex flex-col overflow-hidden transition-all duration-300 z-50">
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-slate-700 bg-slate-800/50">
                <div className="flex items-center gap-2 text-slate-100 font-semibold">
                    <Info size={18} className="text-indigo-400" />
                    <span>Details</span>
                </div>
                <button onClick={onClose} className="p-1 hover:bg-slate-700 rounded-lg text-slate-400 hover:text-white transition-colors">
                    <X size={18} />
                </button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-4 space-y-6">
                {selectedNode && <NodeDetails node={selectedNode} />}
                {selectedEdge && <EdgeDetails edge={selectedEdge} />}
            </div>
        </div>
    );
}

function NodeDetails({ node }: { node: FlowNode }) {
    const data = node.data as unknown as ApiNode;

    return (
        <div className="space-y-4">
            <div>
                <h3 className="text-xl font-bold text-white">{data.service_name}</h3>
                <code className="text-xs text-slate-500 font-mono">{data.id}</code>
            </div>

            {/* Health Status */}
            <div className="p-3 bg-slate-800 rounded-xl border border-slate-700">
                <div className="text-xs uppercase tracking-wider text-slate-500 font-bold mb-2">Health Status</div>
                <div className="flex items-center gap-2">
                    {data.health_status === 0 && <CheckCircle className="text-emerald-500" size={20} />}
                    {data.health_status === 1 && <AlertTriangle className="text-amber-500" size={20} />}
                    {data.health_status === 2 && <Activity className="text-red-500" size={20} />}

                    <span className={clsx("font-medium", {
                        "text-emerald-400": data.health_status === 0,
                        "text-amber-400": data.health_status === 1,
                        "text-red-400": data.health_status === 2,
                    })}>
                        {data.health_status === 0 ? "Healthy / Operational" :
                            data.health_status === 1 ? "Degraded Performance" : "Critical Outage"}
                    </span>
                </div>
            </div>

            {/* Metrics */}
            <div className="grid grid-cols-2 gap-3">
                <div className="p-3 bg-slate-800 rounded-xl border border-slate-700">
                    <div className="text-xs text-slate-500 mb-1">Amplification</div>
                    <div className="text-lg font-mono text-purple-400">{data.amplification_score.toFixed(2)}x</div>
                </div>
                <div className="p-3 bg-slate-800 rounded-xl border border-slate-700">
                    <div className="text-xs text-slate-500 mb-1">Downstream Fails</div>
                    <div className="text-lg font-mono text-indigo-400">{data.downstream_failures}</div>
                </div>
            </div>
        </div>
    );
}

function EdgeDetails({ edge }: { edge: FlowEdge }) {
    return (
        <div className="space-y-4">
            <div>
                <h3 className="text-lg font-bold text-white flex items-center gap-2">
                    <span className="text-slate-400">{edge.source}</span>
                    <span className="text-slate-600">→</span>
                    <span className="text-slate-400">{edge.target}</span>
                </h3>
                <div className="text-xs text-slate-500 mt-1">Causal Dependency</div>
            </div>

            <div className="p-4 bg-slate-800/50 rounded-xl border border-slate-700">
                <div className="text-sm text-slate-300">
                    Evidence suggests a strong causal link between these services.
                    Failures in <b>{edge.source}</b> propagate to <b>{edge.target}</b>.
                </div>
            </div>
        </div>
    );
}
