import { assertSnapshot } from "jsr:@std/testing@0.225.3/snapshot";
import { assertEquals, assertStringIncludes } from "jsr:@std/assert";
import { $ } from "jsr:@david/dax"

/*
 * TODO
 * - Add tests for compatibility mode
 * - Add tests for tome_ignore behavior
*/
const env = {
  TOME_ROOT: "examples",
  PATH: "bin:" + Deno.env.get("PATH"),
}

const envWrapper = {
  TOME_ROOT: "examples",
  TOME_EXECUTABLE: "wrapper.sh",
  PATH: "test/bin:bin:" + Deno.env.get("PATH"),
}

const wrapper = async (args: string) => {
  const v = await $.raw`wrapper.sh ${args}`.env(envWrapper).stderr("inherit").stdout("inherit").captureCombined().noThrow()
  const r = v.stdout.trimEnd()
  return { code: v.code, out: r, lines: r.split("\n"), asKeyVal: r.split("\n").map((l) => l.split("\t").map(t => t.trim())), executable: "wrapper.sh" }
}

const tome = async (args: string) => {
  const v = await $.raw`tome-cli ${args}`.env(env).stderr("inherit").stdout("inherit").captureCombined().noThrow()
  const r = v.stdout.trimEnd()
  return { code: v.code, out: r, lines: r.split("\n"), asKeyVal: r.split("\n").map((l) => l.split("\t").map(t => t.trim())), executable: "tome-cli" }
}

for await (let [executable, fn] of [["tome-cli", tome], ["wrapper.sh", wrapper]]) {
  executable = executable as string
  fn = fn as () => any
  Deno.test(`top level -h`, async function (t): Promise<void> {
    const { out } = await fn("-h");
    await assertSnapshot(t, out);
  });

  Deno.test(`top level help`, async function (t): Promise<void> {
    const { lines } = await fn("help");
    assertEquals(lines, [
      "folder bar: <arg1> <arg2>",
      "folder test-env-injection: ",
      "foo: <arg1> <arg2>",
    ]);
  });

  Deno.test(`completion`, async function (t): Promise<void> {
    const { out, asKeyVal } = await fn("__complete exec fo");
    assertEquals(asKeyVal.slice(0, 2), [
      ["folder", "directory"],
      ["foo", "<arg1> <arg2>"],
    ]);
  });

  Deno.test(`completion folder`, async function (t): Promise<void> {
    const { out } = await fn("__complete exec fold");
    await assertSnapshot(t, out);
  });

  Deno.test(`completion nested script`, async function (t): Promise<void> {
    const { out } = await fn("__complete exec folder bar");
    await assertSnapshot(t, out);
  });

  Deno.test(`completion nested script arguments`, async function (t): Promise<void> {
    const { out } = await fn(`__complete exec foo ""`);
    await assertStringIncludes(out, "--help");
  });

  Deno.test(`completion for nested scripts in directory`, async function (t): Promise<void> {
    const { out } = await fn(`__complete exec folder ""`);
    await assertStringIncludes(out, "test-env-injection")
    await assertStringIncludes(out, "bar")
  });

  Deno.test(`completion nested script arguments if they implement --completion`, async function (t): Promise<void> {
    const { lines } = await fn(`__complete exec foo ""`);
    await assertEquals(lines, [
      "--help\tHelp message for foo",
      "--query\tQuery message for foo",
      "an-argument\tArgument",
      ":4",
      "Completion ended with directive: ShellCompDirectiveNoFileComp",
    ]);
  });

  // Testing this because executing scripts is dangerous if they're not opted in
  Deno.test(`completion nested script arguments does not execute script if not implementing --completion`, async function (t): Promise<void> {
    const { code, lines } = await fn(`__complete exec folder bar ""`);
    assertEquals(code, 0);
    await assertEquals(lines, [
      ":4",
      "Completion ended with directive: ShellCompDirectiveNoFileComp",
    ]);
  });

  Deno.test(`${executable}: injects TOME_ROOT and TOME_EXECUTABLE into environment of script`, async function (t): Promise<void> {
    const { code, lines, executable } = await fn(`exec folder test-env-injection`);
    assertEquals(code, 0);
    await assertEquals(lines.filter(l => l.startsWith("TOME_ROOT=") || l.startsWith("TOME_EXECUTABLE=")).sort(), [
      `TOME_ROOT=${Deno.env.get("PWD")}/examples`,
      `TOME_EXECUTABLE=${executable}`,
    ].sort());
  });
}

Deno.test(`tome-cli: TOME_COMPLETION passed through as env`, async function (t): Promise<void> {
  const { code, lines, executable } = await tome(`__complete exec folder test-env-injection "a"`);
  assertEquals(code, 0);
  await assertEquals(lines.filter(l => l.startsWith("TOME_COMPLETION")).sort(), [
    'TOME_COMPLETION={"args":["folder","test-env-injection"],"last_arg":"test-env-injection","current_word":"a"}',
  ].sort());
})
