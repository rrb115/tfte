import { Play, Pause, SkipBack, SkipForward } from 'lucide-react';
import { clsx } from 'clsx';
import { useState, useEffect } from 'react';

interface TimelineControlProps {
    currentTimestamp: number;
    onTimestampChange: (ts: number) => void;
    live: boolean;
    setLive: (live: boolean) => void;
}

export default function TimelineControl({ currentTimestamp, onTimestampChange, live, setLive }: TimelineControlProps) {
    // Mock range: Last 1 hour
    const now = Date.now();
    const oneHourAgo = now - 60 * 60 * 1000;
    const [min, setMin] = useState(oneHourAgo);
    const [max, setMax] = useState(now);

    // Update range periodically if live
    useEffect(() => {
        if (live) {
            const interval = setInterval(() => {
                const n = Date.now();
                setMax(n);
                setMin(n - 60 * 60 * 1000); // Keep window moving
                onTimestampChange(n);
            }, 1000);
            return () => clearInterval(interval);
        }
    }, [live, onTimestampChange]);

    const formatTime = (ts: number) => new Date(ts).toLocaleTimeString();

    return (
        <div className="absolute bottom-6 left-1/2 -translate-x-1/2 w-[600px] h-20 bg-slate-900/90 backdrop-blur-xl border border-slate-700 shadow-2xl rounded-2xl flex items-center px-6 gap-4 z-50">

            {/* Play/Pause Controls */}
            <div className="flex items-center gap-2 border-r border-slate-700 pr-4">
                <button
                    onClick={() => setLive(!live)}
                    className={clsx(
                        "p-2 rounded-full transition-all",
                        live ? "bg-amber-500/10 text-amber-500 hover:bg-amber-500/20" : "bg-emerald-500 text-slate-900 hover:bg-emerald-400"
                    )}
                >
                    {live ? <Pause size={20} fill="currentColor" /> : <Play size={20} fill="currentColor" className="ml-0.5" />}
                </button>
                <div className="text-xs font-mono text-slate-500 w-16 text-center">
                    {live ? 'LIVE' : 'PAUSED'}
                </div>
            </div>

            {/* Slider */}
            <div className="flex-1 flex flex-col gap-1">
                <div className="flex justify-between text-[10px] text-slate-500 font-mono uppercase">
                    <span>{formatTime(min)}</span>
                    <span className="text-slate-300 font-bold">{formatTime(currentTimestamp)}</span>
                    <span>{formatTime(max)}</span>
                </div>
                <input
                    type="range"
                    min={min}
                    max={max}
                    step={1000} // 1 second steps
                    value={currentTimestamp}
                    onChange={(e) => {
                        setLive(false);
                        onTimestampChange(parseInt(e.target.value));
                    }}
                    className="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-indigo-500 hover:accent-indigo-400"
                />
            </div>

            {/* Step Controls */}
            <div className="flex items-center gap-1 border-l border-slate-700 pl-4">
                <button onClick={() => { setLive(false); onTimestampChange(currentTimestamp - 5000) }} className="p-1 text-slate-400 hover:text-white"><SkipBack size={16} /></button>
                <button onClick={() => { setLive(false); onTimestampChange(currentTimestamp + 5000) }} className="p-1 text-slate-400 hover:text-white"><SkipForward size={16} /></button>
            </div>
        </div>
    );
}
