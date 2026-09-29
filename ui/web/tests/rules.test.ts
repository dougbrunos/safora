import { test } from "node:test"
import assert from "node:assert/strict"
import { buildRules, parseRules } from "../src/lib/rules.ts"

test("parses the importer format", () => {
  assert.deepEqual(parseRules("DIR:Cache,temp;FILE:*.tmp,*.bak"), { dirs: "Cache, temp", files: "*.tmp, *.bak" })
})

test("items without a prefix follow the last group", () => {
  assert.deepEqual(parseRules("DIR:a,b,FILE:*.x,*.y"), { dirs: "a, b", files: "*.x, *.y" })
  assert.deepEqual(parseRules("DIR:only"), { dirs: "only", files: "" })
})

test("empty rules", () => {
  assert.deepEqual(parseRules(""), { dirs: "", files: "" })
  assert.equal(buildRules("", ""), "")
  assert.equal(buildRules("  ,  ", ""), "")
})

test("builds the stored format and round-trips", () => {
  assert.equal(buildRules("node_modules, temp", "*.log"), "DIR:node_modules,temp;FILE:*.log")
  assert.equal(buildRules("", "*.tmp"), "FILE:*.tmp")
  const rules = "DIR:a,b;FILE:*.c"
  const { dirs, files } = parseRules(rules)
  assert.equal(buildRules(dirs, files), rules)
})
