import { useState } from 'react';
import { Network, Clock, FileJson } from 'lucide-react';
import { clsx } from 'clsx';
import useSWR from 'swr';
import { ReactFlowProvider, useOnSelectionChange } from '@xyflow/react';
import type { Node as FlowNode, Edge as FlowEdge } from '@xyflow/react';

import GraphView from './components/GraphView';
import InspectorPanel from './components/InspectorPanel';
import TimelineControl from './components/TimelineControl';
import { fetchGraph } from './api';

function App() {
  return (
    <ReactFlowProvider>
      <AppContent />
    </ReactFlowProvider>
  )
}

function AppContent() {
  const [activeTab, setActiveTab] = useState<'graph' | 'timeline'>('graph');

  // Timeline State
  const [timestamp, setTimestamp] = useState(Date.now());
  const [live, setLive] = useState(true);

  // Selection State
  const [selectedNode, setSelectedNode] = useState<FlowNode | null>(null);
  const [selectedEdge, setSelectedEdge] = useState<FlowEdge | null>(null);

  // Data Fetching
  // If live, we rely on revalidation. If paused, we fetch specific ts.
  const { data: graphData, error } = useSWR(['graph', live ? 'live' : timestamp], () => fetchGraph(timestamp), {
    refreshInterval: live ? 2000 : 0,
    keepPreviousData: true
  });

  useOnSelectionChange({
    onChange: ({ nodes, edges }) => {
      setSelectedNode(nodes[0] || null);
      setSelectedEdge(edges[0] || null);
      if (nodes.length > 0 || edges.length > 0) {
        // Auto open inspector if closed? Or just update state
      }
    },
  });

  return (
    <div className="flex h-screen w-full bg-slate-900 text-slate-50 overflow-hidden font-sans">
      {/* Sidebar */}
      <aside className="w-16 flex flex-col items-center py-6 bg-slate-950 border-r border-slate-800 gap-6 z-10 shrink-0">
        <div className="w-10 h-10 bg-indigo-600 rounded-xl flex items-center justify-center shadow-lg shadow-indigo-500/20 mb-4">
          <Network className="w-6 h-6 text-white" />
        </div>

        <NavButton
          active={activeTab === 'graph'}
          onClick={() => setActiveTab('graph')}
          icon={<Network size={24} />}
          label="Graph"
        />
        <NavButton
          active={activeTab === 'timeline'}
          onClick={() => setActiveTab('timeline')}
          icon={<Clock size={24} />}
          label="Timeline"
        />

        <div className="mt-auto">
          <NavButton
            active={false}
            onClick={() => { }}
            icon={<FileJson size={24} />}
            label="Logs"
          />
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 relative flex flex-col h-full overflow-hidden">
        {/* Header */}
        <header className="h-14 border-b border-slate-800 bg-slate-900/50 backdrop-blur-md flex items-center px-6 justify-between shrink-0 z-10">
          <h1 className="font-semibold text-lg text-slate-200">TFTE Causality Engine</h1>
          <div className="flex items-center gap-4">
            {live && <span className="bg-red-500/20 text-red-500 px-2 py-0.5 rounded text-xs font-bold animate-pulse">LIVE</span>}
            {!live && <span className="bg-slate-700/50 text-slate-400 px-2 py-0.5 rounded text-xs font-bold">HISTORICAL</span>}

            <div className="h-4 w-px bg-slate-700 mx-2" />

            {error && <span className="text-red-400 text-sm">Connection Error</span>}
            <div className="flex items-center gap-2">
              <span className={clsx("w-2 h-2 rounded-full", error ? "bg-red-500" : "bg-emerald-500 animate-pulse")}></span>
              <span className={clsx("text-xs font-medium", error ? "text-red-500" : "text-emerald-500")}>
                {error ? "Offline" : "System Online"}
              </span>
            </div>
          </div>
        </header>

        {/* Viewport */}
        <div className="flex-1 relative w-full h-full bg-slate-900 overflow-hidden">
          {activeTab === 'graph' && (
            <>
              <GraphView data={graphData || null} />

              {/* Overlays */}
              <InspectorPanel
                selectedNode={selectedNode}
                selectedEdge={selectedEdge}
                onClose={() => { setSelectedNode(null); setSelectedEdge(null); }}
              />

              <TimelineControl
                currentTimestamp={timestamp}
                onTimestampChange={setTimestamp}
                live={live}
                setLive={setLive}
              />
            </>
          )}
          {activeTab === 'timeline' && <div className="p-10 text-slate-500">Timeline View (Advanced Histogram Coming Soon)</div>}
        </div>
      </main>
    </div>
  );
}

function NavButton({ active, onClick, icon, label }: { active: boolean, onClick: () => void, icon: React.ReactNode, label: string }) {
  return (
    <button
      onClick={onClick}
      className={clsx(
        "p-3 rounded-xl transition-all duration-200 group relative",
        active ? "bg-indigo-500/10 text-indigo-400" : "text-slate-500 hover:text-slate-200 hover:bg-slate-800"
      )}
      title={label}
    >
      {icon}
      {active && (
        <span className="absolute left-0 top-1/2 -translate-y-1/2 -ml-0.5 w-1 h-8 bg-indigo-500 rounded-r-full" />
      )}
    </button>
  )
}

export default App;
