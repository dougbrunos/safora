import { test } from "node:test"
import assert from "node:assert/strict"
import { NO_RETENTION, describeRetention, normalizeRetention } from "../src/lib/retention.ts"

const t = (key: string) => ({ keep_last: "Keep last {n} copies", keep_days: "Keep last {n} days" })[key] ?? key

test("an empty policy means never delete", () => {
  assert.equal(normalizeRetention(""), NO_RETENTION)
  assert.equal(normalizeRetention(undefined), NO_RETENTION)
})

test("older dashboards saved KEEP 5", () => {
  assert.equal(normalizeRetention("KEEP 5"), "keep 5 runs")
  assert.equal(normalizeRetention("keep 30 days"), "keep 30 days")
})

test("describeRetention translates through t", () => {
  assert.equal(describeRetention("keep 5 runs", t), "Keep last 5 copies")
  assert.equal(describeRetention("keep 7 copies", t), "Keep last 7 copies")
  assert.equal(describeRetention("keep 30 days", t), "Keep last 30 days")
  assert.equal(describeRetention("KEEP 10", t), "Keep last 10 copies")
  assert.equal(describeRetention("something else", t), "something else")
})
