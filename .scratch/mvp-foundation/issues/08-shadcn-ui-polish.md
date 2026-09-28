# 08: Shadcn UI Integration & Visual Polish

**What to build:**
Upgrade the current MVP embedded React dashboard by integrating **shadcn/ui**. Replace basic HTML/CSS with polished shadcn components (Cards, Buttons, Progress Bars, Badges, Dialogs, etc.) styled exactly to match the `ui/SAFORA_DESIGN_SYSTEM.md` specifications. Additionally, integrate the official project logo (`images/logo.jpeg`) into the application layout (e.g., header/sidebar).

**Blocked by:** 06 (Embedded Web Dashboard)

**Status:** done

- [x] Initialize shadcn/ui in the `ui/web/` directory (`npx shadcn@latest init`) and configure it for a dark-first Vite/React/TypeScript project.
- [x] Configure `tailwind.config.ts` and `index.css` to use Safora's specific color tokens: Background `#0B1220`, Surface `#0F172A`, Primary/Teal `#00D1B2`, Accent `#19E6D0`.
- [x] Install required shadcn components (`card`, `button`, `progress`, `badge`, `dialog`, `table`).
- [x] Copy `images/logo.jpeg` to `ui/web/public/` (or import it as an asset) and add it to the main navigation header/sidebar of the dashboard.
- [x] Refactor the existing Dashboard layout, Jobs list, Run history, and the "Import Script" modal to use the newly installed shadcn components.
- [x] Ensure the SSE Live Progress stream updates correctly within a shadcn `Progress` component and `ScrollArea` for logs.
