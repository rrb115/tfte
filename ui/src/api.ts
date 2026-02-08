import axios from 'axios';

// Interfaces matching our Proto definitions roughly
export interface GraphSnapshot {
    timestamp: number;
    nodes: Node[];
    edges: Edge[];
}

export interface Node {
    id: string;
    service_name: string;
    health_status: number; // 0=OK, 1=DEGRADED, 2=DOWN
    amplification_score: number;
    downstream_failures: number;
}

export interface Edge {
    source: string;
    target: string;
    causal_confidence: number;
    is_active: boolean;
}

const API_BASE = 'http://localhost:8081/api';

export const fetchGraph = async (ts: number): Promise<GraphSnapshot> => {
    // Match backend contract: /api/graph?timestamp=<ms>
    const response = await axios.get<GraphSnapshot>(`${API_BASE}/graph`, {
        params: { timestamp: ts },
    });
    return response.data;
};
