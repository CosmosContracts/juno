#!/usr/bin/env python3
"""Create deterministic dependency-bearing SPDX and SLSA release metadata."""

import argparse
import hashlib
import json
import os
import re
from pathlib import Path

parser = argparse.ArgumentParser()
parser.add_argument("--directory", required=True)
parser.add_argument("--version", required=True)
parser.add_argument("--commit", required=True)
parser.add_argument("--repository", required=True)
parser.add_argument("--workflow-sha", required=True)
args = parser.parse_args()
root = Path(args.directory)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def spdx_id(value: str) -> str:
    return "SPDXRef-" + re.sub(r"[^A-Za-z0-9.-]", "-", value)


binary_files = sorted(
    path for path in root.iterdir() if path.is_file() and re.fullmatch(r"junod-linux-(amd64|arm64)", path.name)
)
files = sorted(binary_files + [path for path in root.iterdir() if path.is_file() and path.name.endswith(".tar.gz")])
subjects = [{"name": path.name, "digest": {"sha256": sha(path)}} for path in files]

modules: dict[tuple[str, str], str] = {}
for module_file in sorted(root.glob("junod-linux-*.modules")):
    for line in module_file.read_text().splitlines():
        fields = line.lstrip().split("\t")
        if len(fields) < 3 or fields[0] != "dep":
            continue
        name, version = fields[1], fields[2]
        checksum = fields[3] if len(fields) > 3 and fields[3].startswith("h1:") else ""
        modules[(name, version)] = checksum

apk_packages: set[str] = set()
for package_file in sorted(root.glob("build-dependencies-linux-*.txt")):
    apk_packages.update(line.strip() for line in package_file.read_text().splitlines() if line.strip())

packages = []
relationships = []
for binary in binary_files:
    binary_id = spdx_id(binary.name)
    packages.append(
        {
            "SPDXID": binary_id,
            "name": binary.name,
            "versionInfo": args.version,
            "downloadLocation": "NOASSERTION",
            "filesAnalyzed": False,
            "licenseConcluded": "NOASSERTION",
            "licenseDeclared": "NOASSERTION",
            "copyrightText": "NOASSERTION",
            "checksums": [{"algorithm": "SHA256", "checksumValue": sha(binary)}],
            "comment": f"junod {args.version}, commit {args.commit}",
        }
    )
    relationships.append(
        {"spdxElementId": "SPDXRef-DOCUMENT", "relationshipType": "DESCRIBES", "relatedSpdxElement": binary_id}
    )

for (name, version), checksum in sorted(modules.items()):
    package_id = spdx_id(f"go-{name}-{version}")
    package = {
        "SPDXID": package_id,
        "name": name,
        "versionInfo": version,
        "downloadLocation": f"https://proxy.golang.org/{name}/@v/{version}.zip",
        "filesAnalyzed": False,
        "licenseConcluded": "NOASSERTION",
        "licenseDeclared": "NOASSERTION",
        "copyrightText": "NOASSERTION",
        "externalRefs": [
            {
                "referenceCategory": "PACKAGE-MANAGER",
                "referenceType": "purl",
                "referenceLocator": f"pkg:golang/{name}@{version}",
            }
        ],
    }
    if checksum:
        package["comment"] = f"Go module checksum {checksum}"
    packages.append(package)
    for binary in binary_files:
        relationships.append(
            {"spdxElementId": spdx_id(binary.name), "relationshipType": "DEPENDS_ON", "relatedSpdxElement": package_id}
        )

for package_version in sorted(apk_packages):
    package_id = spdx_id(f"apk-{package_version}")
    packages.append(
        {
            "SPDXID": package_id,
            "name": package_version,
            "versionInfo": package_version,
            "downloadLocation": "https://dl-cdn.alpinelinux.org/alpine/v3.22/main",
            "filesAnalyzed": False,
            "licenseConcluded": "NOASSERTION",
            "licenseDeclared": "NOASSERTION",
            "copyrightText": "NOASSERTION",
            "comment": "Package present in the release builder image",
        }
    )

sbom = {
    "spdxVersion": "SPDX-2.3",
    "dataLicense": "CC0-1.0",
    "SPDXID": "SPDXRef-DOCUMENT",
    "name": f"juno-{args.version}",
    "documentNamespace": f"{args.repository}/releases/{args.version}/{args.commit}/sbom",
    "creationInfo": {"created": "1970-01-01T00:00:00Z", "creators": ["Tool: scripts/release/metadata.py"]},
    "packages": packages,
    "relationships": relationships,
}

resolved_dependencies = [
    {"uri": f"git+{args.repository}@refs/tags/{args.version}", "digest": {"gitCommit": args.commit}},
    {
        "uri": "docker.io/library/golang:1.25.10-alpine3.22",
        "digest": {"sha256": "26b4d7113039cd51356bd7930ecafd1031d2975dc3b6940ec8ed09457e17cf95"},
    },
]
for (name, version), checksum in sorted(modules.items()):
    dependency: dict[str, object] = {"uri": f"pkg:golang/{name}@{version}"}
    if checksum:
        dependency["digest"] = {"goSum": checksum}
    resolved_dependencies.append(dependency)
for package_version in sorted(apk_packages):
    resolved_dependencies.append({"uri": f"pkg:apk/alpine/{package_version}?distro=alpine-3.22"})

provenance = {
    "_type": "https://in-toto.io/Statement/v1",
    "subject": subjects,
    "predicateType": "https://slsa.dev/provenance/v1",
    "predicate": {
        "buildDefinition": {
            "buildType": f"{args.repository}/.github/workflows/release.yml",
            "externalParameters": {"version": args.version, "commit": args.commit},
            "resolvedDependencies": resolved_dependencies,
        },
        "runDetails": {
            "builder": {"id": f"{args.repository}/.github/workflows/release.yml@{args.workflow_sha}"},
            "metadata": {"invocationId": os.getenv("GITHUB_RUN_ID", "local")},
        },
    },
}

for name, value in (("SBOM.spdx.json", sbom), ("provenance.intoto.jsonl", provenance)):
    (root / name).write_text(json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n")
