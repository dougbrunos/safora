import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useI18n } from "@/lib/i18n"
import { buildCron, parseCron, type Mode, type Sched } from "@/lib/schedule"

const MODES: Mode[] = ["manual", "minutes", "hours", "daily", "weekly", "monthly", "cron"]

export function ScheduleField({ value, onChange }: { value: string; onChange: (cron: string) => void }) {
  const { t, locale } = useI18n()
  const [s, setS] = useState<Sched>(() => parseCron(value))

  const update = (patch: Partial<Sched>) => {
    const next = { ...s, ...patch }
    // Switching to "advanced" starts from the expression the current choice produces.
    if (patch.mode === "cron" && s.mode !== "cron") next.cron = buildCron(s)
    setS(next)
    onChange(buildCron(next))
  }

  const weekday = (d: number) => new Date(2024, 0, 7 + d).toLocaleDateString(locale, { weekday: "short" })
  const cron = buildCron(s)

  return (
    <div className="space-y-3">
      <Select value={s.mode} items={MODES.map((m) => ({ value: m, label: t(`sched_${m}`) }))} onValueChange={(v) => update({ mode: v as Mode })}>
        <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
        <SelectContent>
          {MODES.map((m) => <SelectItem key={m} value={m}>{t(`sched_${m}`)}</SelectItem>)}
        </SelectContent>
      </Select>

      {(s.mode === "minutes" || s.mode === "hours") && (
        <div className="flex items-center gap-2 text-sm">
          {t("sched_every")}
          <Select value={String(s.every)} onValueChange={(v) => update({ every: Number(v) })}>
            <SelectTrigger className="w-24"><SelectValue /></SelectTrigger>
            <SelectContent>
              {(s.mode === "minutes" ? [5, 10, 15, 20, 30] : [1, 2, 3, 4, 6, 8, 12]).map((n) => (
                <SelectItem key={n} value={String(n)}>{n}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          {t(s.mode === "minutes" ? "sched_minutes_unit" : "sched_hours_unit")}
        </div>
      )}

      {s.mode === "weekly" && (
        <div className="flex flex-wrap gap-1">
          {[0, 1, 2, 3, 4, 5, 6].map((d) => (
            <Button key={d} type="button" size="sm" variant={s.days.includes(d) ? "default" : "outline"}
              onClick={() => update({ days: s.days.includes(d) ? s.days.filter((x) => x !== d) : [...s.days, d] })}>
              {weekday(d)}
            </Button>
          ))}
        </div>
      )}

      {s.mode === "monthly" && (
        <div className="flex items-center gap-2 text-sm">
          {t("sched_on_day")}
          <Input type="number" min={1} max={28} className="w-20" value={s.dom}
            onChange={(e) => update({ dom: Math.min(28, Math.max(1, Number(e.target.value) || 1)) })} />
        </div>
      )}

      {(s.mode === "daily" || s.mode === "weekly" || s.mode === "monthly") && (
        <div className="flex items-center gap-2 text-sm">
          {t("sched_at")}
          <Input type="time" className="w-32" value={s.time} onChange={(e) => update({ time: e.target.value || "02:00" })} />
        </div>
      )}

      {s.mode === "cron" && (
        <Input placeholder="0 2 * * *" value={s.cron} onChange={(e) => update({ cron: e.target.value })} />
      )}

      <p className="text-xs text-muted-foreground">
        {s.mode === "manual" ? t("sched_manual_help") : <>{t("sched_cron_is")} <code>{cron || "—"}</code></>}
      </p>
    </div>
  )
}
