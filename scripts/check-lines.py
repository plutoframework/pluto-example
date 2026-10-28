#!/usr/bin/env python3
"""Source limits; dependencies, compiled assets and binary resources are not code."""
import os
from pathlib import Path

IGNORED = {".git", "node_modules", "dist", ".cert", "coverage", "test-results", "playwright-report", "__pycache__"}
for directory, directories, files in os.walk("."):
    directories[:] = [name for name in directories if name not in IGNORED]
    for filename in files:
        path = Path(directory) / filename
        # npm owns this resolution graph; preserve its standard format and audit it
        # with npm tooling. It is the sole source-tree line-count exception.
        if path == Path("web/package-lock.json"):
            continue
        try:
            source = path.read_text()
        except UnicodeDecodeError:
            continue
        if len(source.splitlines()) > 1000:
            raise SystemExit(f"File exceeds 1,000 lines: {path}")
print("Source files and generated API modules stay within 1,000 lines")
