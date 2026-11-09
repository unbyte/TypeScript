#!/usr/bin/env python3
"""Compare two compiled checker test binaries on the native compiler fixture."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import random
import re
import statistics
import subprocess
import sys

RESULT = re.compile(r"^BenchmarkCheck-\d+\s+\d+\s+(\d+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$", re.M)


def positive(value):
    number = int(value)
    if number < 1:
        raise argparse.ArgumentTypeError("must be positive")
    return number


def digest(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("baseline", type=Path)
    parser.add_argument("candidate", type=Path)
    parser.add_argument("--pairs", type=positive, default=8)
    parser.add_argument("--iterations", type=positive, default=3)
    parser.add_argument("--procs", type=positive, default=4)
    args = parser.parse_args()
    binaries = {name: getattr(args, name).resolve(strict=True) for name in ("baseline", "candidate")}
    environment = dict(os.environ, GOMAXPROCS=str(args.procs))
    samples = []
    for pair in range(args.pairs):
        order = ("baseline", "candidate") if pair % 2 == 0 else ("candidate", "baseline")
        sample = {"pair": pair + 1, "order": order}
        for name in order:
            result = subprocess.run([
                str(binaries[name]), "-test.run=^$", "-test.bench=^BenchmarkCheck$",
                f"-test.benchtime={args.iterations}x", "-test.count=1",
            ], env=environment, text=True, capture_output=True)
            if result.returncode:
                raise RuntimeError(f"Benchmark {name} exited {result.returncode}:\n{result.stdout}\n{result.stderr}")
            matches = RESULT.findall(result.stdout)
            if len(matches) != 1:
                raise RuntimeError(f"Expected one benchmark result from {name}:\n{result.stdout}\n{result.stderr}")
            sample[name] = dict(zip(("ns_per_op", "bytes_per_op", "allocs_per_op"), map(int, matches[0])))
        samples.append(sample)
        print(f"Completed pair {pair + 1}/{args.pairs}", file=sys.stderr, flush=True)
    summary = {
        name: {metric: statistics.median(s[name][metric] for s in samples)
               for metric in ("ns_per_op", "bytes_per_op", "allocs_per_op")}
        for name in binaries
    }
    changes = [100 * (s["candidate"]["ns_per_op"] / s["baseline"]["ns_per_op"] - 1) for s in samples]
    summary["paired_median_change_percent"] = statistics.median(changes)
    # A descriptive interval over these pairs, not a claim about other workloads
    # or hosts. The fixed seed makes analysis of a saved sample set reproducible.
    rng = random.Random(0)
    bootstrap = sorted(statistics.median(rng.choices(changes, k=len(changes))) for _ in range(10000))
    summary["paired_bootstrap_95_percent"] = [bootstrap[250], bootstrap[9749]]
    print(json.dumps({
        "binaries": {name: {"path": str(path), "sha256": digest(path)} for name, path in binaries.items()},
        "iterations_per_sample": args.iterations,
        "gomaxprocs": args.procs,
        "samples": samples,
        "summary": summary,
    }, indent=2))


if __name__ == "__main__":
    main()
