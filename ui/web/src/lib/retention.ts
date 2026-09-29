// Retention policies are stored as "keep <n> <runs|copies|days>"; older dashboards saved "KEEP 5".
// "none" is the form's value for an empty policy: never delete anything.
export const NO_RETENTION = "none"

export const normalizeRetention = (p?: string) => {
  if (!p) return NO_RETENTION
  const parts = p.toLowerCase().trim().split(/\s+/)
  return (parts.length === 2 ? [...parts, "runs"] : parts).join(" ")
}

// Human-readable, translated summary ("Keep last 5 copies"); falls back to the raw text.
export function describeRetention(policy: string, t: (k: string) => string): string {
  const [verb, n, unit] = normalizeRetention(policy).split(" ")
  if (verb !== "keep" || !/^\d+$/.test(n ?? "")) return policy
  if (unit === "days") return t("keep_days").replace("{n}", n)
  if (unit === "runs" || unit === "copies") return t("keep_last").replace("{n}", n)
  return policy
}
