import React, { useEffect, useMemo } from 'react';
import {
    ReactFlow,
    Background,
    Controls,
    useNodesState,
    useEdgesState,
    Position,
} from '@xyflow/react';
import type { Edge as FlowEdge, Node as FlowNode } from '@xyflow/react';
// Dagre import hack for ESM/CJS interop
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import dagre from 'dagre';
import '@xyflow/react/dist/style.css';

import ServiceNode from './ServiceNode';
import type { GraphSnapshot } from '../api';

interface GraphViewProps {
    data: GraphSnapshot | null;
}

// Layout helper
const getLayoutedElements = (nodes: FlowNode[], edges: FlowEdge[]) => {
    // Handle dagre default export mismatch
    const dagreGraph = (dagre as any).graphlib ? (dagre as any).graphlib : (dagre as any).default?.graphlib;

    if (!dagreGraph) {
        console.error("Dagre graphlib not found", dagre);
        return { nodes, edges };
    }

    const g = new dagreGraph.Graph();
    g.setGraph({ rankdir: 'LR' }); // Left to Right layout
    g.setDefaultEdgeLabel(() => ({}));

    // Set nodes
    nodes.forEach((node) => {
        // Approximate width/height for ServiceNode
        g.setNode(node.id, { width: 220, height: 80 });
    });

    edges.forEach((edge) => {
        g.setEdge(edge.source, edge.target);
    });

    dagre.layout(g);

    return {
        nodes: nodes.map((node) => {
            const nodeWithPosition = g.node(node.id);
            return {
                ...node,
                targetPosition: Position.Left,
                sourcePosition: Position.Right,
                position: {
                    x: nodeWithPosition.x - 110, // Center offset
                    y: nodeWithPosition.y - 40,
                },
            };
        }),
        edges,
    };
};

const GraphView: React.FC<GraphViewProps> = ({ data }) => {
    // Explicit generic types to avoid never[] inference
    const [nodes, setNodes, onNodesChange] = useNodesState<FlowNode>([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState<FlowEdge>([]);

    const nodeTypes = useMemo(() => ({
        service: ServiceNode,
    }), []);

    useEffect(() => {
        if (!data) return;

        // Transform GraphSnapshot to ReactFlow elements
        const initialNodes: FlowNode[] = (data.nodes || []).map(n => ({
            id: n.id,
            type: 'service',
            data: { ...n }, // Pass API node data directly
            position: { x: 0, y: 0 }, // Will be set by dagre
        }));

        const initialEdges: FlowEdge[] = (data.edges || []).map(e => ({
            id: `${e.source}_${e.target}`,
            source: e.source,
            target: e.target,
            animated: e.is_active,
            type: 'default',
            style: { stroke: '#64748b', strokeWidth: 2 },
        }));

        const layouted = getLayoutedElements(initialNodes, initialEdges);

        setNodes(layouted.nodes);
        setEdges(layouted.edges);

    }, [data, setNodes, setEdges]);

    return (
        <div className="w-full h-full bg-slate-900">
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                nodeTypes={nodeTypes}
                fitView
                minZoom={0.1}
            >
                <Background color="#1e293b" gap={16} />
                <Controls className="!bg-slate-800 !border-slate-700 !fill-slate-400" />
            </ReactFlow>
        </div>
    );
};

export default GraphView;
