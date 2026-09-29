import { Fragment, useEffect, useState } from "react"
import { useI18n } from "@/lib/i18n"
import { manuals, type Block } from "@/lib/manual"
import { cn } from "@/lib/utils"

// Renders `code` spans written with backticks.
function Inline({ text }: { text: string }) {
  return (
    <>
      {text.split("`").map((part, i) =>
        i % 2 ? (
          <code key={i} className="rounded bg-muted px-1.5 py-0.5 font-mono text-[0.85em] text-foreground">{part}</code>
        ) : (
          <Fragment key={i}>{part}</Fragment>
        ),
      )}
    </>
  )
}

function BlockView({ block }: { block: Block }) {
  if ("p" in block) return <p className="leading-relaxed text-foreground/90"><Inline text={block.p} /></p>
  if ("ul" in block)
    return (
      <ul className="list-disc space-y-1.5 pl-5 leading-relaxed text-foreground/90 marker:text-muted-foreground">
        {block.ul.map((li, i) => <li key={i}><Inline text={li} /></li>)}
      </ul>
    )
  if ("ol" in block)
    return (
      <ol className="list-decimal space-y-1.5 pl-5 leading-relaxed text-foreground/90 marker:font-medium marker:text-muted-foreground">
        {block.ol.map((li, i) => <li key={i}><Inline text={li} /></li>)}
      </ol>
    )
  if ("code" in block)
    return <pre className="overflow-x-auto rounded-lg bg-[#0B1220] p-4 font-mono text-xs text-slate-200">{block.code}</pre>
  if ("note" in block)
    return (
      <div className="rounded-lg border-l-2 border-warning bg-warning-soft px-4 py-3 text-sm leading-relaxed">
        <Inline text={block.note} />
      </div>
    )
  return (
    <div className="overflow-x-auto rounded-lg border">
      <table className="w-full text-left text-sm">
        <thead className="bg-muted/60 text-xs uppercase tracking-wide text-muted-foreground">
          <tr>{block.table.head.map((h) => <th key={h} className="px-4 py-2.5 font-medium">{h}</th>)}</tr>
        </thead>
        <tbody className="divide-y">
          {block.table.rows.map((row, i) => (
            <tr key={i} className="align-top">
              {row.map((cell, j) => (
                <td key={j} className={cn("px-4 py-2.5 leading-relaxed", j === 0 && "font-medium")}>
                  <Inline text={cell} />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function Manual() {
  const { lang } = useI18n()
  const manual = manuals[lang]
  const [active, setActive] = useState(`manual-${manual.sections[0].id}`)

  // Highlight the section nearest the top of the viewport.
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting).sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top)
        if (visible[0]) setActive(visible[0].target.id)
      },
      { rootMargin: "0px 0px -70% 0px" },
    )
    manual.sections.forEach((s) => {
      const el = document.getElementById(`manual-${s.id}`)
      if (el) observer.observe(el)
    })
    return () => observer.disconnect()
  }, [manual])

  const go = (id: string) => document.getElementById(`manual-${id}`)?.scrollIntoView({ behavior: "smooth", block: "start" })

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      <div>
        <h1 className="text-3xl font-semibold tracking-tight">{manual.title}</h1>
        <p className="mt-1 text-muted-foreground">{manual.subtitle}</p>
      </div>

      <div className="grid gap-10 lg:grid-cols-[13rem_1fr]">
        <nav aria-label={manual.toc} className="lg:sticky lg:top-8 lg:self-start">
          <div className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">{manual.toc}</div>
          <ol className="flex flex-wrap gap-x-4 gap-y-1 lg:flex-col lg:gap-0.5">
            {manual.sections.map((s, i) => (
              <li key={s.id}>
                <button
                  type="button"
                  onClick={() => go(s.id)}
                  aria-current={active === `manual-${s.id}` ? "true" : undefined}
                  className={cn(
                    "w-full rounded-md px-2 py-1 text-left text-sm transition-colors",
                    active === `manual-${s.id}` ? "bg-accent font-medium text-accent-foreground" : "text-muted-foreground hover:text-foreground",
                  )}
                >
                  <span className="mr-2 tabular-nums text-muted-foreground">{i + 1}.</span>
                  {s.title}
                </button>
              </li>
            ))}
          </ol>
        </nav>

        <div className="min-w-0 space-y-12">
          {manual.sections.map((s, i) => (
            <section key={s.id} id={`manual-${s.id}`} className="scroll-mt-8 space-y-4">
              <h2 className="border-b pb-2 text-xl font-semibold tracking-tight">
                <span className="mr-2 tabular-nums text-muted-foreground">{i + 1}.</span>
                {s.title}
              </h2>
              {s.blocks.map((b, k) => <BlockView key={k} block={b} />)}
            </section>
          ))}
        </div>
      </div>
    </div>
  )
}
