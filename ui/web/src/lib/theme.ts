import { useSyncExternalStore } from "react"

export type Theme = "system" | "light" | "dark"

const KEY = "safora_theme"
const media = window.matchMedia("(prefers-color-scheme: dark)")
const listeners = new Set<() => void>()

const read = (): Theme => {
  try {
    const v = localStorage.getItem(KEY)
    if (v === "light" || v === "dark" || v === "system") return v
  } catch {
    // storage unavailable: follow the system
  }
  return "system"
}

let theme = read()

function apply() {
  const dark = theme === "dark" || (theme === "system" && media.matches)
  document.documentElement.classList.toggle("dark", dark)
}

media.addEventListener("change", () => theme === "system" && apply())
apply()

export function setTheme(next: Theme) {
  theme = next
  try {
    localStorage.setItem(KEY, next)
  } catch {
    // preference just won't persist
  }
  apply()
  listeners.forEach((l) => l())
}

export function useTheme() {
  return useSyncExternalStore(
    (cb) => {
      listeners.add(cb)
      return () => listeners.delete(cb)
    },
    () => theme,
  )
}
