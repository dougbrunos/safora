import { useState } from "react"
import { useForm, useFieldArray } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { FolderOpen, Plus, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { FolderBrowser } from "@/components/FolderBrowser"
import { ScheduleField } from "@/components/ScheduleField"
import { useI18n } from "@/lib/i18n"
import { buildRules, parseRules } from "@/lib/rules"
import { normalizeRetention } from "@/lib/retention"
import { cn } from "@/lib/utils"

interface Props {
  onSuccess: () => void;
  onCancel: () => void;
  jobToEdit?: any;
}

type Tab = "general" | "paths" | "options"

const PATH_VARIABLES = ["{today}", "{yesterday}", "{hostname}"]

// Appends v to a path, adding the separator style the path already uses.
const appendVariable = (path: string, v: string) => {
  if (!path || /[\\/]$/.test(path)) return path + v
  return path + (path.includes("\\") ? "\\" : "/") + v
}

// Shown only while the row has focus, so idle rows stay clean.
function VariableChips({ onPick }: { onPick: (v: string) => void }) {
  return (
    <div className="hidden flex-wrap gap-1 group-focus-within:flex">
      {PATH_VARIABLES.map((v) => (
        <button key={v} type="button" onClick={() => onPick(v)}
          className="rounded-md bg-muted px-2 py-0.5 font-mono text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground">
          {v}
        </button>
      ))}
    </div>
  )
}

function OptionRow({ id, title, desc, checked, disabled, onChange }: {
  id: string; title: string; desc: string; checked: boolean; disabled?: boolean; onChange: (v: boolean) => void
}) {
  return (
    <label htmlFor={id} className={cn("flex items-center justify-between gap-4 rounded-lg border p-3", disabled && "opacity-60")}>
      <div>
        <div className="text-sm font-medium">{title}</div>
        <p className="mt-0.5 text-xs text-muted-foreground">{desc}</p>
      </div>
      <Switch id={id} checked={checked} disabled={disabled} onCheckedChange={onChange} />
    </label>
  )
}

export function CreateJobForm({ onSuccess, onCancel, jobToEdit }: Props) {
  const { t } = useI18n()
  const [loading, setLoading] = useState(false)
  const [tab, setTab] = useState<Tab>("general")
  const [browsing, setBrowsing] = useState<string | null>(null) // field being browsed, e.g. "sources.0.path"
  const [exclOpen, setExclOpen] = useState<Record<string, boolean>>({})

  const jobSchema = z.object({
    name: z.string().min(2, t("err_name")),
    sources: z.array(z.object({
      path: z.string().min(1, t("err_source")),
      dirs: z.string().optional(),
      files: z.string().optional(),
    })).min(1),
    destinations: z.array(z.object({
      path: z.string().min(1, t("err_destination")),
    })).min(1),
    enableVSS: z.boolean().default(false),
    retention: z.string().default("keep 5 runs"),
    schedule: z.string().default(""),
    verifyIntegrity: z.boolean().default(false),
    syncDeletions: z.boolean().default(false),
  })

  const {
    register,
    control,
    getValues,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(jobSchema),
    defaultValues: {
      name: jobToEdit?.Name || "",
      sources: jobToEdit?.Sources?.length
        ? jobToEdit.Sources.map((s: any) => ({ path: s.Path, ...parseRules(s.ExclusionRules) }))
        : [{ path: "", dirs: "", files: "" }],
      destinations: jobToEdit?.Destinations?.length
        ? jobToEdit.Destinations.map((d: any) => ({ path: d.Path }))
        : [{ path: "" }],
      enableVSS: jobToEdit?.Description === "VSS Enabled" || false,
      retention: normalizeRetention(jobToEdit?.RetentionPolicy),
      schedule: jobToEdit?.Schedule || "",
      verifyIntegrity: jobToEdit?.VerifyIntegrity || false,
      syncDeletions: jobToEdit?.SyncDeletions || false,
    },
  })

  const sources = useFieldArray({ control, name: "sources" })
  const destinations = useFieldArray({ control, name: "destinations" })

  const enableVSS = watch("enableVSS")
  const verifyIntegrity = watch("verifyIntegrity")
  const syncDeletions = watch("syncDeletions")
  const retention = watch("retention")
  const schedule = watch("schedule")
  const sourceValues = watch("sources")
  const hasDateTemplate = (watch("destinations") || []).some((d: any) => d?.path?.includes("{"))

  const onSubmit = async (data: any) => {
    setLoading(true)
    try {
      const payload = {
        Name: data.name,
        StorageStrategy: jobToEdit?.StorageStrategy || "Date-Stamped Mirroring",
        RetentionPolicy: data.retention,
        Description: data.enableVSS ? "VSS Enabled" : "",
        RetryCount: jobToEdit?.RetryCount || 3,
        RetryWait: jobToEdit?.RetryWait || 30,
        LogOutput: jobToEdit?.LogOutput || "",
        Schedule: data.schedule || "",
        VerifyIntegrity: !!data.verifyIntegrity,
        SyncDeletions: !!data.syncDeletions,
        Sources: data.sources.map((s: any) => ({ Path: s.path, ExclusionRules: buildRules(s.dirs || "", s.files || "") })),
        Destinations: data.destinations.map((d: any) => ({ Path: d.path })),
      }

      const url = jobToEdit ? `/api/jobs/${jobToEdit.ID}` : "/api/jobs"
      const method = jobToEdit ? "PUT" : "POST"

      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      })

      if (res.ok) {
        onSuccess()
      } else {
        alert((await res.text()) || t("save_failed"))
      }
    } catch (e) {
      console.error(e)
      alert(t("save_error"))
    } finally {
      setLoading(false)
    }
  }

  // Jump to the first tab that has a validation error.
  const onInvalid = (errs: any) => setTab(errs.name ? "general" : errs.sources || errs.destinations ? "paths" : "options")

  const tabs: { id: Tab; label: string; hasError: boolean }[] = [
    { id: "general", label: t("tab_general"), hasError: !!errors.name },
    { id: "paths", label: t("tab_paths"), hasError: !!(errors.sources || errors.destinations) },
    { id: "options", label: t("tab_options"), hasError: false },
  ]

  const pathRow = (kind: "sources" | "destinations", i: number, onRemove: () => void, canRemove: boolean) => {
    const name = `${kind}.${i}.path` as const
    return (
      <div className="group space-y-1.5">
        <div className="flex gap-2">
          <Input placeholder={kind === "sources" ? "C:\\Data\\{yesterday}" : "D:\\Backup\\{today}"} {...register(name)} />
          <Button type="button" variant="outline" size="icon" title={t("browse")} aria-label={t("browse")}
            onClick={() => setBrowsing(browsing === name ? null : name)}>
            <FolderOpen className="h-4 w-4" />
          </Button>
          <Button type="button" variant="ghost" size="icon" title={t("remove")} aria-label={t("remove")}
            disabled={!canRemove} onClick={onRemove}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
        <VariableChips onPick={(v) => setValue(name, appendVariable(getValues(name) || "", v))} />
        {browsing === name && (
          <FolderBrowser initial={getValues(name) || ""} onClose={() => setBrowsing(null)}
            onSelect={(p) => setValue(name, p, { shouldValidate: true })} />
        )}
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit, onInvalid)} className="space-y-5">
      <div role="tablist" className="flex gap-1 border-b">
        {tabs.map(({ id, label, hasError }) => (
          <button key={id} type="button" role="tab" aria-selected={tab === id} onClick={() => setTab(id)}
            className={cn(
              "-mb-px border-b-2 px-3 py-2 text-sm font-medium transition-colors",
              tab === id ? "border-primary text-foreground" : "border-transparent text-muted-foreground hover:text-foreground",
            )}>
            {label}
            {hasError && <span className="ml-1.5 inline-block h-1.5 w-1.5 rounded-full bg-destructive align-middle" />}
          </button>
        ))}
      </div>

      <div className="min-h-[19rem] space-y-5">
        {tab === "general" && (
          <>
            <div className="space-y-2">
              <Label htmlFor="name">{t("job_name")}</Label>
              <Input id="name" placeholder={t("job_name_ph")} {...register("name")} />
              {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
            </div>
            <div className="space-y-2">
              <Label>{t("schedule")}</Label>
              <ScheduleField value={schedule || ""} onChange={(v) => setValue("schedule", v)} />
            </div>
          </>
        )}

        {tab === "paths" && (
          <>
            <section className="space-y-2">
              <Label>{t("sources_title")}</Label>
              {sources.fields.map((field, i) => {
                const v = sourceValues?.[i]
                const showExcl = exclOpen[field.id] || !!(v?.dirs || v?.files)
                return (
                  <div key={field.id} className="space-y-2">
                    {pathRow("sources", i, () => sources.remove(i), sources.fields.length > 1)}
                    {errors.sources?.[i]?.path && <p className="text-sm text-destructive">{errors.sources[i]?.path?.message}</p>}
                    {showExcl ? (
                      <div className="grid grid-cols-2 gap-2">
                        <Input placeholder={t("excl_dirs_ph")} aria-label={t("excl_dirs")} {...register(`sources.${i}.dirs`)} />
                        <Input placeholder={t("excl_files_ph")} aria-label={t("excl_files")} {...register(`sources.${i}.files`)} />
                      </div>
                    ) : (
                      <button type="button" className="text-xs text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
                        onClick={() => setExclOpen({ ...exclOpen, [field.id]: true })}>
                        + {t("exclusions_toggle")}
                      </button>
                    )}
                  </div>
                )
              })}
              <Button type="button" variant="ghost" size="sm" className="gap-1 text-muted-foreground"
                onClick={() => sources.append({ path: "", dirs: "", files: "" })}>
                <Plus size={14} /> {t("add_source")}
              </Button>
            </section>

            <section className="space-y-2 border-t pt-5">
              <Label>{t("destinations_title")}</Label>
              {destinations.fields.map((field, i) => (
                <div key={field.id} className="space-y-1.5">
                  {pathRow("destinations", i, () => destinations.remove(i), destinations.fields.length > 1)}
                  {errors.destinations?.[i]?.path && <p className="text-sm text-destructive">{errors.destinations[i]?.path?.message}</p>}
                </div>
              ))}
              <Button type="button" variant="ghost" size="sm" className="gap-1 text-muted-foreground"
                onClick={() => destinations.append({ path: "" })}>
                <Plus size={14} /> {t("add_destination")}
              </Button>
            </section>
          </>
        )}

        {tab === "options" && (
          <>
            <div className="space-y-2">
              <Label>{t("retention_policy")}</Label>
              <Select value={retention as string} onValueChange={(val) => setValue("retention", val as string)}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={t("select_retention")} />
                </SelectTrigger>
                <SelectContent>
                  {[3, 5, 10, 30].map((n) => (
                    <SelectItem key={`r${n}`} value={`keep ${n} runs`}>{t("keep_last").replace("{n}", String(n))}</SelectItem>
                  ))}
                  {[7, 30, 90].map((n) => (
                    <SelectItem key={`d${n}`} value={`keep ${n} days`}>{t("keep_days").replace("{n}", String(n))}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {!hasDateTemplate && <p className="text-xs text-amber-600 dark:text-amber-500">{t("retention_help")}</p>}
            </div>

            <div className="space-y-2">
              <OptionRow id="verify" title={t("verify_integrity")} desc={t("verify_help")}
                checked={!!verifyIntegrity} onChange={(v) => setValue("verifyIntegrity", v)} />
              <OptionRow id="sync" title={t("sync_deletions")} desc={t("sync_help")}
                checked={!!syncDeletions} onChange={(v) => setValue("syncDeletions", v)} />
              <OptionRow id="vss" title={t("enable_vss")} desc={t("vss_help")}
                checked={!!enableVSS} disabled onChange={(v) => setValue("enableVSS", v)} />
            </div>
          </>
        )}
      </div>

      <div className="flex justify-end gap-3 border-t pt-4">
        <Button type="button" variant="ghost" onClick={onCancel}>{t("cancel")}</Button>
        <Button type="submit" disabled={loading}>
          {loading ? t("saving") : jobToEdit ? t("save_changes") : t("create_btn")}
        </Button>
      </div>
    </form>
  )
}
