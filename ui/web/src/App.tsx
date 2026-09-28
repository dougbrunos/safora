import { useState, useEffect, useRef } from 'react';
import { Play, Plus, Server, CheckCircle, XCircle, Terminal, Pencil, Trash2 } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog";
import { CreateJobForm } from '@/components/CreateJobForm';

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [jobs, setJobs] = useState([]);
  const [runs, setRuns] = useState([]);
  const [liveLog, setLiveLog] = useState<string[]>([]);
  const [progress, setProgress] = useState(0);
  const [showImport, setShowImport] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [jobToEdit, setJobToEdit] = useState<any>(null);
  const [importText, setImportText] = useState('');
  
  const scrollRef = useRef<HTMLDivElement>(null);
  
  useEffect(() => {
    fetchJobs();
    fetchRuns();

    const eventSource = new EventSource('/api/stream');
    eventSource.onmessage = (event) => {
      setLiveLog(prev => [...prev.slice(-99), event.data]);
      if (event.data.includes("finished with status")) {
        setProgress(100);
        setTimeout(() => setProgress(0), 3000);
        fetchRuns(); // refresh on finish
      } else if (event.data.includes("Copied") || event.data.includes("Starting job")) {
        setProgress(p => (p >= 90 ? 90 : p + 5));
      }
    };
    return () => eventSource.close();
  }, []);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [liveLog]);

  const fetchJobs = async () => {
    const res = await fetch('/api/jobs');
    if (res.ok) setJobs(await res.json() || []);
  };

  const fetchRuns = async () => {
    const res = await fetch('/api/runs');
    if (res.ok) setRuns(await res.json() || []);
  };

  const handleRunJob = async (id: number) => {
    setProgress(0);
    setActiveTab('dashboard');
    await fetch(`/api/jobs/${id}/run`, { method: 'POST' });
    setLiveLog(prev => [...prev, `[INFO] Triggered Job ID ${id}...`]);
  };

  const handleDeleteJob = async (id: number) => {
    await fetch(`/api/jobs/${id}`, { method: 'DELETE' });
    fetchJobs();
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
    <div className="min-h-screen font-sans bg-background text-foreground selection:bg-safora-teal-500/30">
      {/* Navbar */}
      <nav className="border-b bg-card px-6 py-4 flex items-center justify-between sticky top-0 z-10 shadow-sm">
        <div className="flex items-center gap-4">
          <img src="/logo.jpeg" alt="Safora Logo" className="w-8 h-8 rounded" />
          <span className="text-xl font-semibold tracking-tight">Safora</span>
        </div>
        <div className="flex gap-2">
          <Button variant={activeTab === 'dashboard' ? 'secondary' : 'ghost'} onClick={() => setActiveTab('dashboard')}>
            Dashboard
          </Button>
          <Button variant={activeTab === 'jobs' ? 'secondary' : 'ghost'} onClick={() => setActiveTab('jobs')}>
            Jobs
          </Button>
        </div>
      </nav>

      <main className="p-8 max-w-6xl mx-auto space-y-8 animate-in fade-in duration-500">
        {activeTab === 'dashboard' && (
          <div className="space-y-6">
            <h1 className="text-3xl font-bold tracking-tight">Overview</h1>
            
            {/* Stats */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Successful Runs</CardTitle>
                  <CheckCircle className="h-4 w-4 text-emerald-500" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{runs.filter((r:any) => r.Status === 'success').length}</div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Failed Runs</CardTitle>
                  <XCircle className="h-4 w-4 text-destructive" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{runs.filter((r:any) => r.Status === 'failed').length}</div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Configured Jobs</CardTitle>
                  <Server className="h-4 w-4 text-safora-cyan-400" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{jobs.length}</div>
                </CardContent>
              </Card>
            </div>

            {/* Live Progress */}
            <Card className="border-border overflow-hidden">
              <CardHeader className="bg-muted/50 border-b pb-4 flex flex-col gap-3">
                <div className="flex flex-row items-center gap-2">
                  <Terminal className="h-5 w-5 text-primary" />
                  <CardTitle>Live Telemetry</CardTitle>
                </div>
                {progress > 0 && <Progress value={progress} className="h-2" />}
              </CardHeader>
              <CardContent className="p-0">
                <ScrollArea className="h-72 w-full bg-black/60 rounded-b-lg">
                  <div className="p-4 font-mono text-sm">
                    {liveLog.length === 0 ? (
                      <div className="text-muted-foreground italic">Waiting for events...</div>
                    ) : (
                      liveLog.map((log, i) => {
                        const isErr = log.includes("[ERROR]");
                        const isWarn = log.includes("[WARNING]");
                        return (
                          <div key={i} className={`mb-1 ${isErr ? 'text-destructive' : isWarn ? 'text-amber-400' : 'text-gray-300'}`}>
                            {log}
                          </div>
                        )
                      })
                    )}
                    <div ref={scrollRef} />
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
            
            {/* History */}
            <h2 className="text-xl font-semibold mt-8 mb-4">Recent Runs</h2>
            <div className="space-y-4">
              {runs.map((run: any) => (
                <Card key={run.ID} className="bg-card">
                  <CardContent className="p-4 flex items-center justify-between">
                    <div>
                      <div className="font-semibold">Job #{run.JobID}</div>
                      <div className="text-sm text-muted-foreground">{new Date(run.StartedAt).toLocaleString()}</div>
                    </div>
                    <div className="flex items-center gap-6 text-sm">
                      <div><span className="text-muted-foreground mr-2">Duration</span> {run.DurationSeconds}s</div>
                      <div><span className="text-muted-foreground mr-2">Transferred</span> {(run.BytesTransferred / 1024 / 1024).toFixed(2)} MB</div>
                      <Badge variant={run.Status === 'success' ? 'default' : run.Status === 'failed' ? 'destructive' : 'secondary'}
                             className={run.Status === 'success' ? 'bg-emerald-500/20 text-emerald-400 hover:bg-emerald-500/30 border-emerald-500/20' : ''}>
                        {run.Status}
                      </Badge>
                    </div>
                  </CardContent>
                </Card>
              ))}
              {runs.length === 0 && <div className="text-muted-foreground">No recent runs</div>}
            </div>
          </div>
        )}

        {activeTab === 'jobs' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="text-3xl font-bold tracking-tight">Configured Jobs</h1>
              
              <div className="flex items-center gap-3">
                <Dialog open={showCreate} onOpenChange={(val) => {
                  setShowCreate(val);
                  if (!val) setJobToEdit(null);
                }}>
                  <DialogTrigger>
                    <Button variant="outline" className="gap-2" onClick={() => setJobToEdit(null)}>
                      <Plus size={16} /> New Backup
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                      <DialogTitle>{jobToEdit ? "Edit Backup Job" : "Create Backup Job"}</DialogTitle>
                    </DialogHeader>
                    <div className="mt-4">
                      <CreateJobForm 
                        jobToEdit={jobToEdit}
                        onSuccess={() => { setShowCreate(false); setJobToEdit(null); fetchJobs(); }} 
                        onCancel={() => { setShowCreate(false); setJobToEdit(null); }} 
                      />
                    </div>
                  </DialogContent>
                </Dialog>

                <Dialog open={showImport} onOpenChange={setShowImport}>
                  <DialogTrigger>
                    <Button className="gap-2">
                      <Terminal size={16} /> Import Script
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="sm:max-w-2xl">
                    <DialogHeader>
                      <DialogTitle>Import Legacy Script</DialogTitle>
                    </DialogHeader>
                    <div className="my-4">
                    <p className="text-sm text-muted-foreground mb-4">
                      Paste your Windows <code>.bat</code> or <code>robocopy</code> command. Safora will automatically extract sources, destinations, exclusions, and retention dates.
                    </p>
                    <textarea 
                      className="w-full h-48 bg-muted border rounded-lg p-4 font-mono text-sm outline-none focus:ring-1 focus:ring-primary transition-all"
                      placeholder="robocopy C:\Data D:\Backup /MIR ..."
                      value={importText}
                      onChange={(e) => setImportText(e.target.value)}
                    />
                  </div>
                  <div className="flex justify-end gap-3">
                    <Button variant="ghost" onClick={() => setShowImport(false)}>Cancel</Button>
                    <Button onClick={handleImport}>Parse & Save Job</Button>
                  </div>
                </DialogContent>
              </Dialog>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {jobs.map((job: any) => (
                <Card key={job.ID} className="flex flex-col">
                  <CardHeader>
                    <CardTitle className="text-xl flex items-center justify-between">
                      {job.Name}
                      <div className="flex items-center gap-2">
                        <Badge variant="outline">ID: {job.ID}</Badge>
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => {
                          setJobToEdit(job);
                          setShowCreate(true);
                        }}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <AlertDialog>
                          <AlertDialogTrigger>
                            <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive">
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                              <AlertDialogDescription>
                                This will permanently delete the backup job. The files copied so far will remain untouched.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90" onClick={() => handleDeleteJob(job.ID)}>Delete</AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    </CardTitle>
                    <CardDescription>
                      {job.StorageStrategy} &bull; {job.RetentionPolicy || 'No Retention'}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="flex-1">
                    <div className="bg-muted p-3 rounded-lg text-sm font-mono text-muted-foreground space-y-1">
                      <div>Retries: <span className="text-foreground">{job.RetryCount}</span></div>
                      <div>Wait: <span className="text-foreground">{job.RetryWait}s</span></div>
                      <div>Log: <span className="text-foreground">{job.LogOutput}</span></div>
                    </div>
                  </CardContent>
                  <CardContent className="pt-0">
                    <Button 
                      variant="secondary" 
                      className="w-full gap-2 hover:bg-primary/20 hover:text-primary transition-all"
                      onClick={() => handleRunJob(job.ID)}
                    >
                      <Play size={16} /> Run Now
                    </Button>
                  </CardContent>
                </Card>
              ))}
              {jobs.length === 0 && <div className="text-muted-foreground col-span-2">No jobs configured.</div>}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
