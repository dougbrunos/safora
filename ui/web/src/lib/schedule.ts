// Turns friendly schedule choices into a 5-field cron expression and back.
export type Mode = "manual" | "minutes" | "hours" | "daily" | "weekly" | "monthly" | "cron"

export interface Sched {
  mode: Mode
  every: number // minutes or hours
  time: string // "HH:MM"
  days: number[] // 0 = Sunday
  dom: number // day of month
  cron: string // used by mode "cron"
}

export const defaults: Sched = { mode: "manual", every: 30, time: "02:00", days: [1], dom: 1, cron: "" }

const hhmm = (h: string, m: string) => `${h.padStart(2, "0")}:${m.padStart(2, "0")}`

export function buildCron(s: Sched): string {
  const [h, m] = s.time.split(":").map(Number)
  switch (s.mode) {
    case "manual": return ""
    case "minutes": return `*/${s.every} * * * *`
    case "hours": return `0 */${s.every} * * *`
    case "daily": return `${m} ${h} * * *`
    case "weekly": return `${m} ${h} * * ${[...s.days].sort().join(",") || "*"}`
    case "monthly": return `${m} ${h} ${s.dom} * *`
    case "cron": return s.cron.trim()
  }
}

export function parseCron(cron: string): Sched {
  const c = (cron || "").trim()
  if (!c) return defaults
  let m: RegExpMatchArray | null
  if ((m = c.match(/^\*\/(\d+) \* \* \* \*$/))) return { ...defaults, mode: "minutes", every: +m[1] }
  if ((m = c.match(/^0 \*\/(\d+) \* \* \*$/))) return { ...defaults, mode: "hours", every: +m[1] }
  if ((m = c.match(/^(\d+) (\d+) \* \* \*$/))) return { ...defaults, mode: "daily", time: hhmm(m[2], m[1]) }
  if ((m = c.match(/^(\d+) (\d+) \* \* ([0-6](?:,[0-6])*)$/)))
    return { ...defaults, mode: "weekly", time: hhmm(m[2], m[1]), days: m[3].split(",").map(Number) }
  if ((m = c.match(/^(\d+) (\d+) (\d+) \* \*$/)))
    return { ...defaults, mode: "monthly", time: hhmm(m[2], m[1]), dom: +m[3] }
  return { ...defaults, mode: "cron", cron: c }
}

// One-line summary of a cron expression for cards ("Every day at 02:00").
export function describeSchedule(cron: string, t: (k: string) => string, locale: string): string {
  const s = parseCron(cron)
  const day = (d: number) => new Date(2024, 0, 7 + d).toLocaleDateString(locale, { weekday: "short" })
  switch (s.mode) {
    case "manual": return t("sched_manual_short")
    case "minutes": return `${t("sched_every")} ${s.every} ${t("sched_minutes_unit")}`
    case "hours": return `${t("sched_every")} ${s.every} ${t("sched_hours_unit")}`
    case "daily": return `${t("sched_daily")} ${t("sched_at")} ${s.time}`
    case "weekly": return `${s.days.map(day).join(", ")} ${t("sched_at")} ${s.time}`
    case "monthly": return `${t("sched_on_day")} ${s.dom} ${t("sched_at")} ${s.time}`
    default: return s.cron
  }
}
