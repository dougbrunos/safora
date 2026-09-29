import { useState, useEffect, useRef } from 'react';
import {
  Play, Plus, Terminal, Pencil, Trash2, Eraser, LayoutDashboard, CalendarClock, HardDrive,
  FolderInput, CheckCircle2, XCircle, TriangleAlert, LoaderCircle, Clock, ShieldCheck, Copy,
  Sun, Moon, Monitor, Server, History as HistoryIcon, ChevronDown, BookOpen,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog';
import { CreateJobForm } from '@/components/CreateJobForm';
import { Manual } from '@/components/Manual';
import { useI18n, setLang } from '@/lib/i18n';
import { useTheme, setTheme, type Theme } from '@/lib/theme';
import { describeSchedule } from '@/lib/schedule';
import { describeRetention } from '@/lib/retention';
import { cn } from '@/lib/utils';

type LogLine = { time: string; text: string };

const STATUS = {
  success: { tone: 'bg-success-soft text-success', Icon: CheckCircle2 },
  failed: { tone: 'bg-danger-soft text-danger', Icon: XCircle },
  warning: { tone: 'bg-warning-soft text-warning', Icon: TriangleAlert },
  running: { tone: 'bg-running-soft text-running', Icon: LoaderCircle },
} as const;

const fmtBytes = (n: number) => {
  if (!n) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.min(units.length - 1, Math.floor(Math.log(n) / Math.log(1024)));
  return `${(n / 1024 ** i).toFixed(i ? 1 : 0)} ${units[i]}`;
};

const fmtDuration = (s: number) => (s >= 60 ? `${Math.floor(s / 60)}m ${s % 60}s` : `${s}s`);

const fmtAgo = (iso: string, locale: string) => {
  const secs = Math.round((new Date(iso).getTime() - Date.now()) / 1000);
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });
  for (const [unit, size] of [['day', 86400], ['hour', 3600], ['minute', 60]] as const) {
    if (Math.abs(secs) >= size) return rtf.format(Math.round(secs / size), unit);
  }
  return rtf.format(secs, 'second');
};

function StatusBadge({ status }: { status: string }) {
  const { t } = useI18n();
  const s = STATUS[status as keyof typeof STATUS] ?? STATUS.warning;
  return (
    <span className={cn('inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium', s.tone)}>
      <s.Icon className={cn('h-3.5 w-3.5', status === 'running' && 'animate-spin')} />
      {t('status_' + status)}
    </span>
  );
}

type RunLog = { Level: string; Message: string; CreatedAt: string };

function RunRow({ run, name, locale, expandable, open, logs, onToggle }: {
  run: any; name: string; locale: string; expandable?: boolean; open?: boolean; logs?: RunLog[]; onToggle?: () => void;
}) {
  const { t } = useI18n();
  const row = (
    <div className="flex items-center gap-4 px-5 py-3.5">
      <div className="min-w-0 flex-1">
        <div className="truncate font-medium">{name}</div>
        <div className="text-xs text-muted-foreground">{new Date(run.StartedAt).toLocaleString(locale)}</div>
      </div>
      <dl className="hidden gap-8 text-sm tabular-nums sm:flex">
        <div>
          <dt className="text-xs text-muted-foreground">{t('duration')}</dt>
          <dd>{fmtDuration(run.DurationSeconds)}</dd>
        </div>
        <div>
          <dt className="text-xs text-muted-foreground">{t('transferred')}</dt>
          <dd>{fmtBytes(run.BytesTransferred)} <span className="text-muted-foreground">· {run.FilesProcessed} {t('files')}</span></dd>
        </div>
      </dl>
      <StatusBadge status={run.Status} />
      {expandable && <ChevronDown className={cn('h-4 w-4 shrink-0 text-muted-foreground transition-transform', open && 'rotate-180')} />}
    </div>
  );
  if (!expandable) return row;
  return (
    <div>
      <button type="button" aria-expanded={open} onClick={onToggle} className="block w-full text-left transition-colors hover:bg-muted/50">
        {row}
      </button>
      {open && (
        <div className="max-h-64 space-y-1 overflow-y-auto border-t bg-[#0B1220] px-5 py-3 font-mono text-xs leading-relaxed">
          {!logs ? <span className="text-slate-500">…</span>
            : logs.length === 0 ? <span className="italic text-slate-500">{t('no_logs')}</span>
            : logs.map((l, i) => (
              <div key={i} className="flex gap-3">
                <span className="shrink-0 text-slate-600 tabular-nums">{new Date(l.CreatedAt).toLocaleTimeString(locale)}</span>
                <span className={cn('w-16 shrink-0', l.Level === 'ERROR' ? 'text-red-400' : l.Level === 'WARNING' ? 'text-amber-400' : 'text-slate-500')}>{l.Level}</span>
                <span className={l.Level === 'ERROR' ? 'text-red-300' : l.Level === 'WARNING' ? 'text-amber-200' : 'text-slate-300'}>{l.Message}</span>
              </div>
            ))}
        </div>
      )}
    </div>
  );
}

function Stat({ label, value, Icon, tone }: { label: string; value: number; Icon: typeof Server; tone: string }) {
  return (
    <Card className="flex-row items-center gap-4 px-5 py-5 shadow-[var(--shadow-card)]">
      <div className={cn('flex h-11 w-11 shrink-0 items-center justify-center rounded-xl', tone)}>
        <Icon className="h-5 w-5" />
      </div>
      <div>
        <div className="text-3xl font-semibold leading-none tracking-tight tabular-nums">{value}</div>
        <div className="mt-1.5 text-sm text-muted-foreground">{label}</div>
      </div>
    </Card>
  );
}

function ThemeToggle() {
  const { t } = useI18n();
  const theme = useTheme();
  const options: { value: Theme; Icon: typeof Sun }[] = [
    { value: 'light', Icon: Sun },
    { value: 'system', Icon: Monitor },
    { value: 'dark', Icon: Moon },
  ];
  return (
    <div role="group" aria-label="Theme" className="inline-flex rounded-lg bg-muted p-0.5">
      {options.map(({ value, Icon }) => (
        <button
          key={value}
          type="button"
          title={t('theme_' + value)}
          aria-label={t('theme_' + value)}
          aria-pressed={theme === value}
          onClick={() => setTheme(value)}
          className={cn(
            'flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors',
            theme === value ? 'bg-card text-foreground shadow-sm' : 'hover:text-foreground',
          )}
        >
          <Icon className="h-4 w-4" />
        </button>
      ))}
    </div>
  );
}

export default function App() {
  const { t, lang, locale } = useI18n();
  const [activeTab, setActiveTab] = useState<'dashboard' | 'jobs' | 'history' | 'manual'>('dashboard');
  const [jobs, setJobs] = useState<any[]>([]);
  const [runs, setRuns] = useState<any[]>([]);
  const [liveLog, setLiveLog] = useState<LogLine[]>([]);
  const [progress, setProgress] = useState(0);
  const [online, setOnline] = useState(false);
  const [showImport, setShowImport] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [jobToEdit, setJobToEdit] = useState<any>(null);
  const [jobToDelete, setJobToDelete] = useState<any>(null);
  const [importText, setImportText] = useState('');
  const [notice, setNotice] = useState<string | null>(null);
  const [historyStatus, setHistoryStatus] = useState('all');
  const [historyJob, setHistoryJob] = useState('all');
  const [openRun, setOpenRun] = useState<number | null>(null);
  const [runLogs, setRunLogs] = useState<Record<number, RunLog[]>>({});

  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchJobs();
    fetchRuns();

    const eventSource = new EventSource('/api/stream');
    eventSource.onopen = () => setOnline(true);
    eventSource.onerror = () => setOnline(false);
    eventSource.onmessage = (event) => {
      // A new Run starts a fresh telemetry view.
      const starting = event.data.includes("Starting job");
      const line = { time: new Date().toLocaleTimeString(locale), text: event.data };
      setLiveLog(prev => [...(starting ? [] : prev.slice(-99)), line]);
      if (event.data.includes("finished with status")) {
        setProgress(100);
        setTimeout(() => setProgress(0), 3000);
        fetchRuns(); // refresh on finish
      } else if (event.data.includes("Copied") || event.data.includes("Starting job")) {
        setProgress(p => (p >= 90 ? 90 : p + 5));
      }
    };
    return () => eventSource.close();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!notice) return;
    const timer = setTimeout(() => setNotice(null), 7000);
    return () => clearTimeout(timer);
  }, [notice]);

  useEffect(() => {
    scrollRef.current?.scrollIntoView({ behavior: "smooth" });
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
    const res = await fetch(`/api/jobs/${id}/run`, { method: 'POST' }).catch(() => null);
    if (res?.status === 202) {
      // The Run's "Starting job" event clears the telemetry by itself.
      setActiveTab('dashboard');
    } else if (res?.status === 409) {
      setNotice(t('run_already'));
    } else {
      setNotice(`${t('run_failed')} ${res ? await res.text() : ''}`.trim());
    }
  };

  const handleDeleteJob = async (id: number) => {
    await fetch(`/api/jobs/${id}`, { method: 'DELETE' });
    setJobToDelete(null);
    fetchJobs();
    fetchRuns();
  };

  const handleImport = async () => {
    const res = await fetch('/api/importer/parse', {
      method: 'POST',
      body: importText,
      headers: { 'Content-Type': 'text/plain' }
    });
    if (!res.ok) return alert(t("import_failed"));
    // The endpoint only parses; persist the resulting Job.
    const save = await fetch('/api/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(await res.json()),
    });
    if (!save.ok) return alert((await save.text()) || t("import_failed"));
    await fetchJobs();
    setShowImport(false);
    setImportText('');
  };

  const toggleRun = async (id: number) => {
    if (openRun === id) return setOpenRun(null);
    setOpenRun(id);
    if (!runLogs[id]) {
      const res = await fetch(`/api/runs/${id}`);
      if (res.ok) {
        const logs = (await res.json()).logs || [];
        setRunLogs(prev => ({ ...prev, [id]: logs }));
      }
    }
  };

  const openCreate = (job: any = null) => {
    setJobToEdit(job);
    setShowCreate(true);
  };

  const jobName = (id: number) => jobs.find(j => j.ID === id)?.Name ?? `${t('job_n')}${id}`;
  const lastRun = (id: number) => runs.find(r => r.JobID === id);
  const filteredRuns = runs.filter(r =>
    (historyStatus === 'all' || r.Status === historyStatus) &&
    (historyJob === 'all' || String(r.JobID) === historyJob));
  const ok = runs.filter(r => r.Status === 'success').length;
  const failed = runs.filter(r => r.Status === 'failed').length;

  const navItems = [
    { id: 'dashboard' as const, label: t('nav_dashboard'), Icon: LayoutDashboard },
    { id: 'jobs' as const, label: t('nav_jobs'), Icon: CalendarClock },
    { id: 'history' as const, label: t('nav_history'), Icon: HistoryIcon },
    { id: 'manual' as const, label: t('nav_manual'), Icon: BookOpen },
  ];

  const brand = (
    <div className="flex items-center gap-2.5">
      <img src="/logo-solid.png" alt="" className="h-8 w-8 rounded-lg" />
      <span className="text-lg font-semibold tracking-tight">Safora</span>
    </div>
  );

  const controls = (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" className="h-8 w-10 px-0 font-medium" onClick={() => setLang(lang === 'pt' ? 'en' : 'pt')}>
        {lang === 'pt' ? 'EN' : 'PT'}
      </Button>
      <ThemeToggle />
    </div>
  );

  return (
    <div className="min-h-screen bg-background text-foreground md:grid md:grid-cols-[15rem_1fr]">
      {/* Sidebar (desktop) */}
      <aside className="sticky top-0 hidden h-screen flex-col border-r bg-card px-4 py-5 md:flex">
        {brand}
        <nav className="mt-8 flex flex-col gap-1">
          {navItems.map(({ id, label, Icon }) => (
            <button
              key={id}
              onClick={() => setActiveTab(id)}
              aria-current={activeTab === id ? 'page' : undefined}
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                activeTab === id
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-muted hover:text-foreground',
              )}
            >
              <Icon className="h-4 w-4" /> {label}
            </button>
          ))}
        </nav>
        <div className="mt-auto space-y-4">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span className={cn('h-2 w-2 rounded-full', online ? 'bg-success' : 'bg-muted-foreground/40')} />
            {online ? t('daemon_online') : t('daemon_offline')}
          </div>
          {controls}
        </div>
      </aside>

      <div className="min-w-0">
        {/* Top bar (mobile) */}
        <header className="sticky top-0 z-10 flex items-center justify-between gap-3 border-b bg-card/90 px-4 py-3 backdrop-blur md:hidden">
          {brand}
          <div className="flex items-center gap-2">
            {navItems.map(({ id, Icon, label }) => (
              <Button key={id} size="icon" variant={activeTab === id ? 'secondary' : 'ghost'} aria-label={label} onClick={() => setActiveTab(id)}>
                <Icon className="h-4 w-4" />
              </Button>
            ))}
            {controls}
          </div>
        </header>

        <main className="mx-auto max-w-5xl space-y-8 p-4 md:p-10">
          {activeTab === 'dashboard' && (
            <div className="space-y-8 animate-in fade-in duration-300">
              <div>
                <h1 className="text-3xl font-semibold tracking-tight">{t('overview')}</h1>
                <p className="mt-1 text-muted-foreground">{t('subtitle_overview')}</p>
              </div>

              <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
                <Stat label={t('successful_runs')} value={ok} Icon={CheckCircle2} tone="bg-success-soft text-success" />
                <Stat label={t('failed_runs')} value={failed} Icon={XCircle} tone="bg-danger-soft text-danger" />
                <Stat label={t('configured_jobs')} value={jobs.length} Icon={Server} tone="bg-running-soft text-running" />
              </div>

              {/* Live telemetry: always a dark terminal, in both themes */}
              <section className="overflow-hidden rounded-xl border border-slate-800 bg-[#0B1220] text-slate-200 shadow-[var(--shadow-card)]">
                <div className="flex items-center gap-3 px-4 py-3">
                  <span className="relative flex h-2.5 w-2.5">
                    {online && <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[#19E6D0] opacity-60" />}
                    <span className={cn('relative inline-flex h-2.5 w-2.5 rounded-full', online ? 'bg-[#19E6D0]' : 'bg-slate-600')} />
                  </span>
                  <Terminal className="h-4 w-4 text-slate-400" />
                  <h2 className="text-sm font-medium">{t('live_telemetry')}</h2>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="ml-auto gap-1.5 text-slate-300 hover:bg-white/10 hover:text-white"
                    onClick={() => { setLiveLog([]); setProgress(0); }}
                  >
                    <Eraser size={14} /> {t('clear')}
                  </Button>
                </div>
                <div className="h-0.5 bg-white/5">
                  <div className="h-full bg-gradient-to-r from-[#00D1B2] to-[#19E6D0] transition-[width] duration-500" style={{ width: `${progress}%` }} />
                </div>
                <ScrollArea className="h-72 w-full">
                  <div className="space-y-1 p-4 font-mono text-[13px] leading-relaxed">
                    {liveLog.length === 0 ? (
                      <div className="italic text-slate-500">{t('waiting')}</div>
                    ) : (
                      liveLog.map((line, i) => {
                        const m = line.text.match(/^\[(\w+)\]\s*(.*)$/);
                        const level = m?.[1] ?? 'INFO';
                        const color = level === 'ERROR' ? 'text-red-400' : level === 'WARNING' ? 'text-amber-400' : 'text-slate-500';
                        return (
                          <div key={i} className="flex gap-3">
                            <span className="shrink-0 text-slate-600 tabular-nums">{line.time}</span>
                            <span className={cn('w-16 shrink-0', color)}>{level}</span>
                            <span className={level === 'ERROR' ? 'text-red-300' : level === 'WARNING' ? 'text-amber-200' : 'text-slate-300'}>
                              {m?.[2] ?? line.text}
                            </span>
                          </div>
                        );
                      })
                    )}
                    <div ref={scrollRef} />
                  </div>
                </ScrollArea>
              </section>

              {/* Recent runs */}
              <section>
                <div className="mb-3 flex items-center justify-between">
                  <h2 className="text-lg font-semibold">{t('recent_runs')}</h2>
                  {runs.length > 5 && (
                    <Button variant="ghost" size="sm" onClick={() => setActiveTab('history')}>{t('view_all')}</Button>
                  )}
                </div>
                <Card className="gap-0 divide-y py-0 shadow-[var(--shadow-card)]">
                  {runs.slice(0, 5).map((run: any) => (
                    <RunRow key={run.ID} run={run} name={jobName(run.JobID)} locale={locale} />
                  ))}
                  {runs.length === 0 && (
                    <div className="flex flex-col items-center gap-1 px-5 py-12 text-center">
                      <Clock className="mb-2 h-6 w-6 text-muted-foreground" />
                      <div className="font-medium">{t('no_runs')}</div>
                      <div className="text-sm text-muted-foreground">{t('empty_runs_hint')}</div>
                    </div>
                  )}
                </Card>
              </section>
            </div>
          )}

          {activeTab === 'jobs' && (
            <div className="space-y-8 animate-in fade-in duration-300">
              <div className="flex flex-wrap items-end justify-between gap-4">
                <div>
                  <h1 className="text-3xl font-semibold tracking-tight">{t('configured_jobs')}</h1>
                  <p className="mt-1 text-muted-foreground">{t('subtitle_jobs')}</p>
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="outline" className="gap-2" onClick={() => setShowImport(true)}>
                    <Terminal size={16} /> {t('import_script')}
                  </Button>
                  <Button className="gap-2" onClick={() => openCreate()}>
                    <Plus size={16} /> {t('new_backup')}
                  </Button>
                </div>
              </div>

              <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
                {jobs.map((job: any) => {
                  const last = lastRun(job.ID);
                  return (
                    <Card key={job.ID} className="gap-4 px-5 py-5 shadow-[var(--shadow-card)] transition-shadow hover:shadow-md">
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <h3 className="truncate text-lg font-semibold leading-tight">{job.Name}</h3>
                          <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                            {last ? (
                              <>
                                <span className={cn('h-2 w-2 rounded-full', last.Status === 'success' ? 'bg-success' : last.Status === 'failed' ? 'bg-danger' : last.Status === 'running' ? 'bg-running' : 'bg-warning')} />
                                {t('last_run')}: {fmtAgo(last.StartedAt, locale)}
                              </>
                            ) : t('never_run')}
                          </div>
                        </div>
                        <div className="flex shrink-0 gap-1">
                          <Button variant="ghost" size="icon" className="h-8 w-8" aria-label={t('edit_job')} onClick={() => openCreate(job)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-danger" aria-label={t('delete')} onClick={() => setJobToDelete(job)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>

                      <div className="space-y-1.5 rounded-lg bg-muted/60 p-3 font-mono text-xs">
                        {job.Sources?.map((s: any) => (
                          <div key={s.ID} className="flex items-center gap-2 text-muted-foreground" title={s.Path}>
                            <FolderInput className="h-3.5 w-3.5 shrink-0" /> <span className="truncate">{s.Path}</span>
                          </div>
                        ))}
                        {job.Destinations?.map((d: any) => (
                          <div key={d.ID} className="flex items-center gap-2" title={d.Path}>
                            <HardDrive className="h-3.5 w-3.5 shrink-0 text-brand-text" /> <span className="truncate">{d.Path}</span>
                          </div>
                        ))}
                      </div>

                      <div className="flex flex-wrap gap-1.5 text-xs">
                        <span className="inline-flex items-center gap-1.5 rounded-full bg-muted px-2.5 py-1 text-muted-foreground">
                          <Clock className="h-3 w-3" /> {describeSchedule(job.Schedule, t, locale)}
                        </span>
                        <span className="rounded-full bg-muted px-2.5 py-1 text-muted-foreground">
                          {job.RetentionPolicy ? describeRetention(job.RetentionPolicy, t) : t('no_retention')}
                        </span>
                        {job.VerifyIntegrity && (
                          <span className="inline-flex items-center gap-1.5 rounded-full bg-running-soft px-2.5 py-1 text-running">
                            <ShieldCheck className="h-3 w-3" /> {t('badge_verify')}
                          </span>
                        )}
                        {job.SyncDeletions && (
                          <span className="inline-flex items-center gap-1.5 rounded-full bg-warning-soft px-2.5 py-1 text-warning">
                            <Copy className="h-3 w-3" /> {t('badge_sync')}
                          </span>
                        )}
                      </div>

                      <Button className="w-full gap-2" onClick={() => handleRunJob(job.ID)}>
                        <Play size={16} /> {t('run_now')}
                      </Button>
                    </Card>
                  );
                })}
              </div>

              {jobs.length === 0 && (
                <Card className="items-center gap-2 px-6 py-14 text-center shadow-[var(--shadow-card)]">
                  <div className="mb-2 flex h-12 w-12 items-center justify-center rounded-2xl bg-accent text-accent-foreground">
                    <HardDrive className="h-6 w-6" />
                  </div>
                  <div className="text-lg font-semibold">{t('empty_jobs_title')}</div>
                  <div className="text-sm text-muted-foreground">{t('empty_jobs_hint')}</div>
                  <Button className="mt-3 gap-2" onClick={() => openCreate()}>
                    <Plus size={16} /> {t('new_backup')}
                  </Button>
                </Card>
              )}
            </div>
          )}

          {activeTab === 'manual' && <Manual />}

          {activeTab === 'history' && (
            <div className="space-y-6 animate-in fade-in duration-300">
              <div>
                <h1 className="text-3xl font-semibold tracking-tight">{t('history_title')}</h1>
                <p className="mt-1 text-muted-foreground">{t('history_subtitle')}</p>
              </div>

              <div className="flex flex-wrap items-center gap-3">
                <div className="inline-flex rounded-lg bg-muted p-0.5">
                  {['all', 'success', 'warning', 'failed'].map((st) => (
                    <button
                      key={st}
                      type="button"
                      aria-pressed={historyStatus === st}
                      onClick={() => setHistoryStatus(st)}
                      className={cn(
                        'rounded-md px-3 py-1 text-sm transition-colors',
                        historyStatus === st ? 'bg-card font-medium text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground',
                      )}
                    >
                      {st === 'all' ? t('filter_all') : t('status_' + st)}
                    </button>
                  ))}
                </div>
                <Select
                  value={historyJob}
                  items={[{ value: 'all', label: t('filter_job_all') }, ...jobs.map((j: any) => ({ value: String(j.ID), label: j.Name }))]}
                  onValueChange={(v) => setHistoryJob(v as string)}
                >
                  <SelectTrigger className="w-56">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">{t('filter_job_all')}</SelectItem>
                    {jobs.map((j: any) => <SelectItem key={j.ID} value={String(j.ID)}>{j.Name}</SelectItem>)}
                  </SelectContent>
                </Select>
                <span className="ml-auto text-sm text-muted-foreground tabular-nums">{filteredRuns.length}</span>
              </div>

              <Card className="gap-0 divide-y py-0 shadow-[var(--shadow-card)]">
                {filteredRuns.map((run: any) => (
                  <RunRow
                    key={run.ID}
                    run={run}
                    name={jobName(run.JobID)}
                    locale={locale}
                    expandable
                    open={openRun === run.ID}
                    logs={runLogs[run.ID]}
                    onToggle={() => toggleRun(run.ID)}
                  />
                ))}
                {filteredRuns.length === 0 && (
                  <div className="flex flex-col items-center gap-1 px-5 py-12 text-center">
                    <Clock className="mb-2 h-6 w-6 text-muted-foreground" />
                    <div className="font-medium">{runs.length === 0 ? t('no_runs') : t('no_results')}</div>
                  </div>
                )}
              </Card>
            </div>
          )}
        </main>
      </div>

      {notice && (
        <div role="status" className="fixed bottom-4 right-4 z-50 flex max-w-sm items-start gap-3 rounded-lg border border-warning/40 bg-card px-4 py-3 text-sm shadow-lg">
          <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0 text-warning" />
          <span>{notice}</span>
          <button type="button" aria-label={t('close')} onClick={() => setNotice(null)} className="ml-1 text-muted-foreground hover:text-foreground">×</button>
        </div>
      )}

      {/* Create / edit */}
      <Dialog open={showCreate} onOpenChange={(val) => { setShowCreate(val); if (!val) setJobToEdit(null); }}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{jobToEdit ? t('edit_job') : t('create_job')}</DialogTitle>
          </DialogHeader>
          <div className="mt-4">
            <CreateJobForm
              key={jobToEdit?.ID ?? 'new'}
              jobToEdit={jobToEdit}
              onSuccess={() => { setShowCreate(false); setJobToEdit(null); fetchJobs(); }}
              onCancel={() => { setShowCreate(false); setJobToEdit(null); }}
            />
          </div>
        </DialogContent>
      </Dialog>

      {/* Import */}
      <Dialog open={showImport} onOpenChange={setShowImport}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t('import_title')}</DialogTitle>
          </DialogHeader>
          <div className="my-4">
            <p className="mb-4 text-sm text-muted-foreground">{t('import_help')}</p>
            <textarea
              className="h-48 w-full rounded-lg border bg-muted p-4 font-mono text-sm outline-none transition-all focus:ring-1 focus:ring-primary"
              placeholder="robocopy C:\Data D:\Backup /MIR ..."
              value={importText}
              onChange={(e) => setImportText(e.target.value)}
            />
          </div>
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => setShowImport(false)}>{t('cancel')}</Button>
            <Button onClick={handleImport}>{t('parse_save')}</Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete */}
      <AlertDialog open={!!jobToDelete} onOpenChange={(val) => !val && setJobToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('delete_named').replace('{name}', jobToDelete?.Name ?? '')}</AlertDialogTitle>
            <AlertDialogDescription>{t('confirm_delete_desc')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
            <AlertDialogAction className="bg-destructive text-destructive-foreground hover:bg-destructive/90" onClick={() => handleDeleteJob(jobToDelete.ID)}>
              {t('delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
