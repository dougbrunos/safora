// Exclusion Rules are stored as "DIR:a,b;FILE:*.x,*.y"; the form edits the two lists separately.
export function parseRules(rules: string): { dirs: string; files: string } {
  const dirs: string[] = []
  const files: string[] = []
  let isDir = false
  for (let item of (rules || "").split(/[;,]/)) {
    item = item.trim()
    if (item.startsWith("DIR:")) { isDir = true; item = item.slice(4) }
    else if (item.startsWith("FILE:")) { isDir = false; item = item.slice(5) }
    if (item) (isDir ? dirs : files).push(item)
  }
  return { dirs: dirs.join(", "), files: files.join(", ") }
}

export function buildRules(dirs: string, files: string): string {
  const list = (s: string) => s.split(",").map((x) => x.trim()).filter(Boolean).join(",")
  return [list(dirs) && `DIR:${list(dirs)}`, list(files) && `FILE:${list(files)}`].filter(Boolean).join(";")
}
