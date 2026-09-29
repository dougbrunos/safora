import { test } from "node:test"
import assert from "node:assert/strict"
import { buildCron, parseCron, describeSchedule, defaults, type Sched } from "../src/lib/schedule.ts"

const sched = (patch: Partial<Sched>): Sched => ({ ...defaults, ...patch })

test("each mode builds the expected cron expression", () => {
  assert.equal(buildCron(sched({ mode: "manual" })), "")
  assert.equal(buildCron(sched({ mode: "minutes", every: 15 })), "*/15 * * * *")
  assert.equal(buildCron(sched({ mode: "hours", every: 6 })), "0 */6 * * *")
  assert.equal(buildCron(sched({ mode: "daily", time: "02:30" })), "30 2 * * *")
  assert.equal(buildCron(sched({ mode: "weekly", time: "22:05", days: [5, 1, 3] })), "5 22 * * 1,3,5")
  assert.equal(buildCron(sched({ mode: "monthly", time: "06:00", dom: 15 })), "0 6 15 * *")
  assert.equal(buildCron(sched({ mode: "cron", cron: "  0 0 * * 0  " })), "0 0 * * 0")
})

test("parsing a built expression gives the same choices back", () => {
  const cases: Sched[] = [
    sched({ mode: "minutes", every: 30 }),
    sched({ mode: "hours", every: 12 }),
    sched({ mode: "daily", time: "23:59" }),
    sched({ mode: "weekly", time: "08:15", days: [0, 6] }),
    sched({ mode: "monthly", time: "01:00", dom: 28 }),
  ]
  for (const s of cases) {
    const back = parseCron(buildCron(s))
    assert.equal(back.mode, s.mode)
    assert.equal(buildCron(back), buildCron(s))
  }
})

test("empty and unknown expressions", () => {
  assert.equal(parseCron("").mode, "manual")
  const odd = parseCron("0 9-17 * * 1-5")
  assert.equal(odd.mode, "cron")
  assert.equal(odd.cron, "0 9-17 * * 1-5")
})

const t = (key: string) =>
  ({ sched_manual_short: "Manual", sched_daily: "Every day", sched_at: "at", sched_every: "Every", sched_hours_unit: "hours", sched_on_day: "On day" })[key] ?? key

test("describeSchedule reads like a sentence", () => {
  assert.equal(describeSchedule("", t, "en-US"), "Manual")
  assert.equal(describeSchedule("0 2 * * *", t, "en-US"), "Every day at 02:00")
  assert.equal(describeSchedule("0 */6 * * *", t, "en-US"), "Every 6 hours")
  assert.equal(describeSchedule("0 6 15 * *", t, "en-US"), "On day 15 at 06:00")
  assert.match(describeSchedule("0 8 * * 1,3", t, "en-US"), /^Mon, Wed at 08:00$/)
  assert.equal(describeSchedule("0 9-17 * * 1-5", t, "en-US"), "0 9-17 * * 1-5")
})
