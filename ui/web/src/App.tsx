import { useState, useEffect } from 'react';
import { Play, Plus, Server, CheckCircle, XCircle, X, Terminal } from 'lucide-react';

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [jobs, setJobs] = useState([]);
  const [runs, setRuns] = useState([]);
  const [liveLog, setLiveLog] = useState<string[]>([]);
  const [showImport, setShowImport] = useState(false);
  const [importText, setImportText] = useState('');
  
  useEffect(() => {
    fetchJobs();
    fetchRuns();

    const eventSource = new EventSource('/api/stream');
    eventSource.onmessage = (event) => {
      setLiveLog(prev => [...prev.slice(-49), event.data]);
      if (event.data.includes("finished with status")) {
        fetchRuns(); // refresh on finish
      }
    };
    return () => eventSource.close();
  }, []);

  const fetchJobs = async () => {
    const res = await fetch('/api/jobs');
    if (res.ok) {
      setJobs(await res.json() || []);
    }
  };

  const fetchRuns = async () => {
    const res = await fetch('/api/runs');
    if (res.ok) {
      setRuns(await res.json() || []);
    }
  };

  const handleRunJob = async (id: number) => {
    await fetch(`/api/jobs/${id}/run`, { method: 'POST' });
    setLiveLog(prev => [...prev, `Triggered Job ID ${id}...`]);
  };

  const handleImport = async () => {
    const res = await fetch('/api/importer/parse', {
      method: 'POST',
      body: importText,
      headers: { 'Content-Type': 'text/plain' }
    });
    if (res.ok) {
      await fetchJobs();
      setShowImport(false);
      setImportText('');
    } else {
      alert("Failed to import script");
    }
  };

  return (
    <div className="min-h-screen font-sans bg-safora-navy-950 text-[#F8FAFC]">
      {/* Navbar */}
      <nav className="border-b border-[#263247] bg-[#0F172A] px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[#19E6D0] to-[#009B88] flex items-center justify-center font-bold">S</div>
          <span className="text-xl font-semibold">Safora</span>
        </div>
        <div className="flex gap-4">
          <button onClick={() => setActiveTab('dashboard')} className={`px-3 py-1 rounded-md ${activeTab === 'dashboard' ? 'bg-[#172033] text-safora-teal-500' : 'text-gray-400 hover:text-white'}`}>Dashboard</button>
          <button onClick={() => setActiveTab('jobs')} className={`px-3 py-1 rounded-md ${activeTab === 'jobs' ? 'bg-[#172033] text-safora-teal-500' : 'text-gray-400 hover:text-white'}`}>Jobs</button>
        </div>
      </nav>

      <main className="p-8 max-w-6xl mx-auto">
        {activeTab === 'dashboard' && (
          <div className="space-y-6">
            <h1 className="text-3xl font-semibold mb-6">Dashboard</h1>
            
            {/* Stats */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-[#0F172A] border border-[#263247] p-5 rounded-xl flex items-center gap-4">
                <div className="p-3 bg-green-500/20 text-green-500 rounded-lg"><CheckCircle /></div>
                <div>
                  <div className="text-sm text-gray-400">Successful Runs</div>
                  <div className="text-2xl font-semibold">{runs.filter((r:any) => r.Status === 'success').length}</div>
                </div>
              </div>
              <div className="bg-[#0F172A] border border-[#263247] p-5 rounded-xl flex items-center gap-4">
                <div className="p-3 bg-red-500/20 text-red-500 rounded-lg"><XCircle /></div>
                <div>
                  <div className="text-sm text-gray-400">Failed Runs</div>
                  <div className="text-2xl font-semibold">{runs.filter((r:any) => r.Status === 'failed').length}</div>
                </div>
              </div>
              <div className="bg-[#0F172A] border border-[#263247] p-5 rounded-xl flex items-center gap-4">
                <div className="p-3 bg-blue-500/20 text-blue-500 rounded-lg"><Server /></div>
                <div>
                  <div className="text-sm text-gray-400">Configured Jobs</div>
                  <div className="text-2xl font-semibold">{jobs.length}</div>
                </div>
              </div>
            </div>

            {/* Live Progress */}
            <div className="bg-[#0F172A] border border-[#263247] rounded-xl overflow-hidden">
              <div className="px-5 py-4 border-b border-[#263247] bg-[#172033] flex items-center gap-2">
                <Terminal size={18} className="text-safora-teal-500" />
                <h2 className="font-semibold">Live Telemetry</h2>
              </div>
              <div className="p-5 font-mono text-sm h-64 overflow-y-auto bg-black/40">
                {liveLog.length === 0 ? (
                  <div className="text-gray-500 italic">Waiting for events...</div>
                ) : (
                  liveLog.map((log, i) => <div key={i} className="mb-1 text-gray-300">{log}</div>)
                )}
              </div>
            </div>
            
            {/* History */}
            <h2 className="text-xl font-semibold mt-8 mb-4">Recent Runs</h2>
            <div className="space-y-3">
              {runs.map((run: any) => (
                <div key={run.ID} className="bg-[#0F172A] border border-[#263247] p-4 rounded-xl flex items-center justify-between">
                  <div>
                    <div className="font-medium">Job #{run.JobID}</div>
                    <div className="text-sm text-gray-400">{new Date(run.StartedAt).toLocaleString()}</div>
                  </div>
                  <div className="flex gap-6 text-sm">
                    <div><span className="text-gray-500">Duration:</span> {run.DurationSeconds}s</div>
                    <div><span className="text-gray-500">Transferred:</span> {(run.BytesTransferred / 1024 / 1024).toFixed(2)} MB</div>
                    <div className={`font-semibold capitalize ${run.Status === 'success' ? 'text-green-500' : run.Status === 'failed' ? 'text-red-500' : 'text-amber-500'}`}>
                      {run.Status}
                    </div>
                  </div>
                </div>
              ))}
              {runs.length === 0 && <div className="text-gray-500">No recent runs</div>}
            </div>
          </div>
        )}

        {activeTab === 'jobs' && (
          <div>
            <div className="flex items-center justify-between mb-6">
              <h1 className="text-3xl font-semibold">Configured Jobs</h1>
              <button 
                onClick={() => setShowImport(true)}
                className="bg-safora-teal-500 hover:bg-[#009B88] text-white px-4 py-2 rounded-md flex items-center gap-2 font-medium transition-colors"
              >
                <Plus size={18} />
                Import Script
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {jobs.map((job: any) => (
                <div key={job.ID} className="bg-[#0F172A] border border-[#263247] p-5 rounded-xl">
                  <h3 className="text-lg font-semibold mb-1">{job.Name} (ID: {job.ID})</h3>
                  <div className="text-sm text-gray-400 mb-4">{job.StorageStrategy} • {job.RetentionPolicy}</div>
                  
                  <div className="space-y-2 text-sm text-gray-300 font-mono bg-[#172033] p-3 rounded-lg mb-4">
                    <div>Retries: {job.RetryCount}</div>
                    <div>Wait: {job.RetryWait}s</div>
                    <div>Log: {job.LogOutput}</div>
                  </div>

                  <button 
                    onClick={() => handleRunJob(job.ID)}
                    className="w-full bg-[#172033] hover:bg-[#263247] border border-[#263247] text-white px-4 py-2 rounded-md flex items-center justify-center gap-2 transition-colors"
                  >
                    <Play size={16} className="text-safora-teal-500" />
                    Run Now
                  </button>
                </div>
              ))}
              {jobs.length === 0 && <div className="text-gray-500">No jobs configured.</div>}
            </div>
          </div>
        )}
      </main>

      {/* Import Modal */}
      {showImport && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50">
          <div className="bg-[#0F172A] border border-[#263247] rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl">
            <div className="px-6 py-4 border-b border-[#263247] flex items-center justify-between">
              <h3 className="text-lg font-semibold">Import Legacy Script</h3>
              <button onClick={() => setShowImport(false)} className="text-gray-400 hover:text-white"><X size={20} /></button>
            </div>
            <div className="p-6">
              <p className="text-sm text-gray-400 mb-4">
                Paste your Windows <code>.bat</code> or <code>robocopy</code> command. Safora will automatically extract sources, destinations, exclusions, and retention dates.
              </p>
              <textarea 
                className="w-full h-48 bg-[#172033] border border-[#263247] rounded-lg p-4 font-mono text-sm text-gray-200 outline-none focus:border-safora-teal-500"
                placeholder="robocopy C:\Data D:\Backup /MIR ..."
                value={importText}
                onChange={(e) => setImportText(e.target.value)}
              />
            </div>
            <div className="px-6 py-4 border-t border-[#263247] bg-[#0B1220] flex justify-end gap-3">
              <button onClick={() => setShowImport(false)} className="px-4 py-2 text-gray-400 hover:text-white">Cancel</button>
              <button onClick={handleImport} className="bg-safora-teal-500 hover:bg-[#009B88] text-white px-6 py-2 rounded-md font-medium">Parse & Save Job</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
