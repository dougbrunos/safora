# 09: Manual Job Creation Form (UI)

**What to build:**
A robust, user-friendly form in the Web Dashboard for creating Jobs manually without relying on the Script Importer. The form should leverage `shadcn/ui` Form components (powered by `react-hook-form` and `zod` for validation). It must include fields for Job Name, Source Path, Destination Path, Exclusions, VSS toggle, and Retention Policy rules, securely submitting a `POST` request to `/api/jobs`.

**Blocked by:** 08 (Shadcn UI Integration & Visual Polish)

**Status:** done

- [x] Install shadcn form components: `npx shadcn@latest add form input select checkbox switch`.
- [x] Create a Zod schema validating the Job creation payload (matching the Go backend schema).
- [x] Build a `CreateJobForm` component with fields:
  - Name (text)
  - Source Path (text, with helper text about `{today}`/`{yesterday}` templates)
  - Destination Path (text)
  - Enable VSS (switch/toggle)
  - Exclusions (text, comma-separated)
- [x] Render the form inside a dedicated page (`/jobs/new`) or a sliding `Sheet` / `Dialog` component.
- [x] Connect the form submission to the `POST /api/jobs` REST endpoint, displaying success/error toast notifications upon completion.
