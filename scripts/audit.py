#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-only
"""Check source/distribution contents for local configuration and common secrets.

Additional installation-specific strings can be supplied via --deny-file. Keep
that file outside the distribution. This tool never prints matching secret text.
"""
import argparse
import io
import pathlib
import re
import tarfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("root", type=pathlib.Path)
parser.add_argument("--deny-file", type=pathlib.Path)
args = parser.parse_args()
deny = []
if args.deny_file:
    deny = [line.strip().lower().encode() for line in args.deny_file.read_text().splitlines() if line.strip()]
patterns = [
    re.compile(rb"(?:gh[pousr]_[A-Za-z0-9]{30,}|github_pat_[A-Za-z0-9_]{40,}|sk-proj-[A-Za-z0-9_-]{30,})"),
    re.compile(rb"-----BEGIN (?:OPENSSH |RSA |EC |DSA )?PRIVATE KEY-----"),
    re.compile(rb"/home/(?!root(?:/|\b))[a-zA-Z0-9_-]+/"),
    re.compile(rb"\b192\.168\.[0-9]{1,3}\.[0-9]{1,3}\b"),
]
issues = []
count = 0

def check(name, data):
    global count
    count += 1
    base = pathlib.PurePosixPath(name).name
    if base.endswith(".pyc") or "__pycache__" in pathlib.PurePosixPath(name).parts:
        issues.append((name, "compiled Python cache"))
    if base in {"config.json", "state.json", "nextcloud.env", "dropbox.env"} or base.endswith((".env", ".pem", ".key")):
        issues.append((name, "runtime/secret file"))
    if any(value in data.lower() for value in deny):
        issues.append((name, "installation-specific data"))
    if any(pattern.search(data) for pattern in patterns):
        issues.append((name, "secret/private-path pattern"))
    if name.endswith((".tar.gz", ".tgz")):
        try:
            with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as archive:
                for member in archive.getmembers():
                    path = pathlib.PurePosixPath(member.name)
                    if path.is_absolute() or ".." in path.parts or member.issym() or member.islnk():
                        issues.append((name, "unsafe archive member"))
                    elif member.isfile():
                        check(name + ":" + member.name, archive.extractfile(member).read())
        except (tarfile.TarError, OSError):
            issues.append((name, "unreadable archive"))

for path in sorted(args.root.rglob("*")):
    rel = path.relative_to(args.root)
    if any(part in {".git", "__pycache__"} for part in rel.parts):
        continue
    if path.is_symlink():
        issues.append((str(rel), "unexpected source symlink"))
    elif path.is_file():
        check(str(rel), path.read_bytes())
if issues:
    for name, reason in issues:
        print(f"FAIL {name}: {reason}")
    raise SystemExit(1)
print(f"Privacy/content checks passed for {count} files and archive members.")
