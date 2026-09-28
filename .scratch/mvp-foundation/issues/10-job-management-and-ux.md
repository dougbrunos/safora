# 10: Job Management & UX Navigation

**What to build:**
Complete the Job lifecycle management by adding Edit and Delete capabilities to the UI, and improve the user experience by automatically navigating to the Live Telemetry view when a Job is triggered.

**Blocked by:** 09 (Manual Job Creation Form)

**Status:** done

- [x] Backend (if not present): Ensure `PUT` or `PATCH /api/jobs/:id` is implemented in the Go API to update existing Jobs.
- [x] UI - Delete: Add a "Delete" action to the Job list (preferably in a Dropdown Menu or action column) with an Alert Dialog (shadcn `alert-dialog`) confirming the destructive action before calling `DELETE /api/jobs/:id`.
- [x] UI - Edit: Add an "Edit" action that reuses the `CreateJobForm` logic but pre-fills it with the existing Job data and calls the update endpoint.
- [x] UX - Auto-navigation: Update the "Run Now" action handler. When a Job is manually started, automatically use the router to navigate the user to the Dashboard / Live Telemetry screen so they can immediately see the SSE progress stream.
