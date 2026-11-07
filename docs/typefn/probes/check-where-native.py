"""Constraint-design research against the existing fork; no proposed syntax.

Run from the repository root with the rebuilt compiler path as the argument.
Diagnostics are expected experimental observations, not where conformance tests.
"""

import json
from pathlib import Path
import re
import subprocess
import sys
import tempfile

compiler = Path(sys.argv[1] if len(sys.argv) > 1 else "built/local/tsgo").resolve()
baseline = Path(__file__).with_name("where-baseline.ts").resolve()
cases = [
    ("dependent-bound-and-result", """
typefn Select<T extends object, K extends keyof T> { return T[K]; }
type Selected = Select<{id:42}, "id">;
const correct: Selected = 42;
const incorrect: Selected = 43;
""", [2322]),
    ("invalid-dependent-key", """
typefn Select<T extends object, K extends keyof T> { return T[K]; }
type Invalid = Select<{id:42}, "missing">;
""", [2344]),
    ("symbolic-bound-replays", """
typefn Accept<T extends string> { return "ok"; }
type Deferred<U> = Accept<U>;
type Good = Deferred<"hello">;
const correct: Good = "ok";
type Invalid = Deferred<number>;
""", [2344]),
    ("bounded-caller", """
typefn Accept<T extends string> { return "ok"; }
typefn Forward<T extends string> { return Accept<T>; }
type Deferred<T extends string> = Forward<T>;
const correct: Deferred<"hello"> = "ok";
const incorrect: Deferred<"hello"> = "wrong";
""", [2322]),
    ("any-and-never-meet-native-bound", """
typefn Accept<T extends string> { return "ok"; }
const a: Accept<any> = "ok";
const b: Accept<never> = "ok";
""", []),
    ("numeric-bound-is-not-literal-readiness", """
typefn Twice<N extends number> { return N * 2; }
type Invalid = Twice<number>;
""", [-1]),
    ("discarded-bound-check-replays", """
typefn Accept<T extends string> { return 1; }
typefn Discard<T> { type unused = Accept<T>; return "ok"; }
type Deferred<T> = Discard<T>;
type Invalid = Deferred<number>;
""", [2344]),
    ("unreached-application-does-not-check-bound", """
typefn Accept<T extends string> { return 0; }
typefn Choose<> { if (false) { return Accept<number>; } return 1; }
const result: Choose<> = 1;
""", []),
    ("native-bound-checked-at-definition", """
typefn Unused<T extends Uppercase<number>> { return 0; }
""", [2344]),
    ("dsl-in-bound-executes-at-definition", """
typefn Twice<N> { return N * 2; }
typefn Unused<T extends Twice<"bad">> { return 0; }
""", [-1]),
    ("later-parameter-bound-currently-accepted", """
typefn Select<K extends keyof T, T> { return T[K]; }
const correct: Select<"id", {id:42}> = 42;
const incorrect: Select<"id", {id:42}> = 43;
""", [2322]),
    ("blocked-earlier-bound-delays-later-failure", """
typefn Gate<T extends string, U extends number> { return 0; }
type Pending<X> = Gate<X, string>;
""", []),
    ("ready-earlier-bound-reports-failure", """
typefn Gate<U extends number, T extends string> { return 0; }
type Pending<X> = Gate<string, X>;
""", [2344]),
]


def run(directory, name, source_file, dsl):
    options = {
        "strict": True, "exactOptionalPropertyTypes": True, "noEmit": True,
        "types": [], "target": "es2022", "module": "esnext",
        "ignoreDeprecations": "6.0",
    }
    if dsl:
        options["typefn"] = {"version": 1}
    config = directory / f"{name}.json"
    config.write_text(json.dumps({"compilerOptions": options, "files": [str(source_file)]}))
    result = subprocess.run(
        [str(compiler), "--project", str(config), "--pretty", "false"],
        capture_output=True, text=True, timeout=20,
    )
    if result.returncode not in (0, 2) or result.stderr:
        raise AssertionError((name, result.returncode, result.stdout, result.stderr))
    diagnostics = [
        {"line": int(line), "code": int(code)}
        for line, code in re.findall(r"\((\d+),\d+\): error TS(-?\d+):", result.stdout)
    ]
    if result.returncode != (2 if diagnostics else 0):
        raise AssertionError((name, "unrecognized compiler output", result.stdout))
    return diagnostics, result.stdout


observations = []
with tempfile.TemporaryDirectory(prefix="typefn-where-native-") as temporary:
    directory = Path(temporary)
    expected = []
    for index, line in enumerate(baseline.read_text().splitlines()):
        match = re.search(r"@probe-error (\d+)(?: native=(\d+))?", line)
        if match:
            expected.append({"line": index + 2, "code": int(match[2] or match[1])})
    actual, output = run(directory, "native-baseline", baseline, False)
    if actual != expected:
        raise AssertionError(("native-baseline", expected, actual, output))
    observations.append({"case": "native-baseline", "expectedRejections": len(actual)})
    for name, source, codes in cases:
        path = directory / f"{name}.ts"
        path.write_text(source)
        actual, output = run(directory, name, path, True)
        if sorted(d["code"] for d in actual) != sorted(codes):
            raise AssertionError((name, codes, actual, output))
        if -1 in codes and "scalar operator requires single unbranded literals" not in output:
            raise AssertionError((name, "expected scalar-domain diagnostic", output))
        observations.append({"case": name, "diagnostics": actual})

version = subprocess.run(
    [str(compiler), "--version"], capture_output=True, text=True, check=True, timeout=20,
).stdout.strip()
print(json.dumps({"compiler": version, "observations": observations}, indent=2))
