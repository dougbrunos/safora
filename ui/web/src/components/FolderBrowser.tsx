import { useEffect, useState } from "react"
import { ArrowUp, Folder, HardDrive } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useI18n } from "@/lib/i18n"

interface Listing {
  Path: string
  Parent: string
  Entries: { Name: string; Path: string }[]
}

// Starts at the typed directory; for templated paths, at the part before the first {variable}.
const startDir = (value: string) => value.split("{")[0].replace(/(?<=.)[\\/]+$/, "")

export function FolderBrowser({ initial, onSelect, onClose }: {
  initial: string
  onSelect: (path: string) => void
  onClose: () => void
}) {
  const { t } = useI18n()
  const [listing, setListing] = useState<Listing | null>(null)
  const [error, setError] = useState(false)

  const load = async (path: string, fallback = true) => {
    const res = await fetch(`/api/fs/list?path=${encodeURIComponent(path)}`)
    if (res.ok) {
      setListing(await res.json())
      setError(false)
    } else if (fallback && path) {
      load("", false) // typed path doesn't exist: start from the roots
    } else {
      setError(true)
    }
  }

  useEffect(() => {
    load(startDir(initial))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const atRoots = !listing?.Path

  return (
    <div className="rounded-lg border bg-muted/40">
      <div className="flex items-center gap-2 border-b px-3 py-2">
        <Button type="button" variant="ghost" size="icon" className="h-7 w-7" title={t("browse_up")} aria-label={t("browse_up")}
          disabled={atRoots} onClick={() => load(listing?.Parent ?? "")}>
          <ArrowUp className="h-4 w-4" />
        </Button>
        <Button type="button" variant="ghost" size="icon" className="h-7 w-7" title={t("browse_roots")} aria-label={t("browse_roots")}
          disabled={atRoots} onClick={() => load("")}>
          <HardDrive className="h-4 w-4" />
        </Button>
        <span className="truncate font-mono text-xs text-muted-foreground" title={listing?.Path}>
          {listing?.Path || t("browse_roots")}
        </span>
      </div>

      <div className="max-h-56 overflow-y-auto p-1">
        {error && <p className="px-3 py-4 text-sm text-danger">{t("browse_error")}</p>}
        {!error && listing?.Entries.length === 0 && (
          <p className="px-3 py-4 text-sm text-muted-foreground">{t("browse_empty")}</p>
        )}
        {listing?.Entries.map((e) => (
          <button key={e.Path} type="button" onClick={() => load(e.Path, false)}
            className="flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-left text-sm transition-colors hover:bg-muted">
            <Folder className="h-4 w-4 shrink-0 text-brand-text" />
            <span className="truncate">{e.Name}</span>
          </button>
        ))}
      </div>

      <div className="flex justify-end gap-2 border-t px-3 py-2">
        <Button type="button" variant="ghost" size="sm" onClick={onClose}>{t("cancel")}</Button>
        <Button type="button" size="sm" disabled={atRoots} onClick={() => { onSelect(listing!.Path); onClose() }}>
          {t("browse_select")}
        </Button>
      </div>
    </div>
  )
}
