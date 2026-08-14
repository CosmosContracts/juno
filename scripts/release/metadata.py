#!/usr/bin/env python3
"""Create deterministic dependency-bearing SPDX and SLSA release metadata."""

import argparse
import base64
import hashlib
import json
import os
import re
from pathlib import Path
from typing import Any

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


def go_sum_dirhash1(checksum: str) -> str:
    """Convert Go's h1: directory Hash1 to in-toto dirHash1 lowercase hex."""
    if not checksum.startswith("h1:"):
        raise ValueError(f"unsupported Go module checksum: {checksum}")
    try:
        digest = base64.b64decode(checksum[3:], validate=True)
    except ValueError as exc:
        raise ValueError(f"invalid Go module checksum: {checksum}") from exc
    if len(digest) != hashlib.sha256().digest_size:
        raise ValueError(f"invalid Go module checksum length: {checksum}")
    return digest.hex()


def spdx_id(value: str) -> str:
    readable = re.sub(r"[^A-Za-z0-9.-]", "-", value).strip("-.") or "item"
    suffix = hashlib.sha256(value.encode()).hexdigest()[:16]
    return f"SPDXRef-{readable[:80]}-{suffix}"


def module_package(identity: tuple[str, str, str], package_id: str, comment: str = "") -> dict[str, Any]:
    name, version, checksum = identity
    package: dict[str, Any] = {
        "SPDXID": package_id,
        "name": name,
        "downloadLocation": "NOASSERTION",
        "filesAnalyzed": False,
        "licenseConcluded": "NOASSERTION",
        "licenseDeclared": "NOASSERTION",
        "copyrightText": "NOASSERTION",
    }
    if version:
        package["versionInfo"] = version
        package["downloadLocation"] = f"https://proxy.golang.org/{name}/@v/{version}.zip"
        package["externalRefs"] = [{
            "referenceCategory": "PACKAGE-MANAGER",
            "referenceType": "purl",
            "referenceLocator": f"pkg:golang/{name}@{version}",
        }]
    notes = [comment] if comment else []
    if checksum:
        notes.append(f"Go module checksum {checksum}")
    if notes:
        package["comment"] = "; ".join(notes)
    return package


binary_files = sorted(
    path for path in root.iterdir() if path.is_file() and re.fullmatch(r"junod-linux-(amd64|arm64)", path.name)
)
files = sorted(binary_files + [path for path in root.iterdir() if path.is_file() and path.name.endswith(".tar.gz")])
subjects = [{"name": path.name, "digest": {"sha256": sha(path)}} for path in files]

# Each record is original name/version/checksum followed by the selected
# replacement name/version/checksum (empty when there is no replacement).
modules: set[tuple[str, str, str, str, str, str]] = set()
for module_file in sorted(root.glob("junod-linux-*.modules")):
    pending: tuple[str, str, str] | None = None
    for line in module_file.read_text().splitlines():
        fields = line.lstrip().split("\t")
        if len(fields) >= 3 and fields[0] == "dep":
            if pending:
                modules.add((*pending, "", "", ""))
            pending = (fields[1], fields[2], fields[3] if len(fields) > 3 and fields[3].startswith("h1:") else "")
        elif len(fields) >= 2 and fields[0] == "=>":
            if pending is None:
                raise ValueError(f"replacement without dependency in {module_file}: {line}")
            replacement_version = fields[2] if len(fields) > 2 else ""
            replacement_checksum = fields[3] if len(fields) > 3 and fields[3].startswith("h1:") else ""
            if replacement_version == "(devel)" or fields[1].startswith((".", "/")):
                raise ValueError(f"local Go module replacement is not release-verifiable in {module_file}: {line}")
            modules.add((*pending, fields[1], replacement_version, replacement_checksum))
            pending = None
        elif pending:
            modules.add((*pending, "", "", ""))
            pending = None
    if pending:
        modules.add((*pending, "", "", ""))

apk_packages: set[str] = set()
for package_file in sorted(root.glob("build-dependencies-linux-*.txt")):
    apk_packages.update(line.strip() for line in package_file.read_text().splitlines() if line.strip())

package_map: dict[str, dict[str, Any]] = {}
relationship_map: dict[tuple[str, str, str], dict[str, str]] = {}


def add_package(package: dict[str, Any]) -> None:
    package_id = package["SPDXID"]
    previous = package_map.setdefault(package_id, package)
    if previous != package:
        raise ValueError(f"SPDX identifier collision: {package_id}")


def add_relationship(source: str, kind: str, target: str) -> None:
    key = (source, kind, target)
    relationship_map[key] = {
        "spdxElementId": source,
        "relationshipType": kind,
        "relatedSpdxElement": target,
    }


for binary in binary_files:
    binary_id = spdx_id(binary.name)
    add_package({
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
    })
    add_relationship("SPDXRef-DOCUMENT", "DESCRIBES", binary_id)

for name, version, checksum, replacement_name, replacement_version, replacement_checksum in sorted(modules):
    original = (name, version, checksum)
    original_id = spdx_id(f"go-original-{name}@{version}#{checksum}")
    add_package(module_package(original, original_id, "Original dependency identity" if replacement_name else ""))
    selected_id = original_id
    if replacement_name:
        replacement = (replacement_name, replacement_version, replacement_checksum)
        selected_id = spdx_id(f"go-replacement-{replacement_name}@{replacement_version}#{replacement_checksum}")
        add_package(module_package(replacement, selected_id, "Selected Go module replacement"))
        add_relationship(selected_id, "VARIANT_OF", original_id)
    for binary in binary_files:
        add_relationship(spdx_id(binary.name), "DEPENDS_ON", selected_id)

for package_version in sorted(apk_packages):
    package_id = spdx_id(f"apk-{package_version}")
    add_package({
        "SPDXID": package_id,
        "name": package_version,
        "versionInfo": package_version,
        "downloadLocation": "https://dl-cdn.alpinelinux.org/alpine/v3.22/main",
        "filesAnalyzed": False,
        "licenseConcluded": "NOASSERTION",
        "licenseDeclared": "NOASSERTION",
        "copyrightText": "NOASSERTION",
        "comment": "Package present in the release builder image",
    })

sbom = {
    "spdxVersion": "SPDX-2.3",
    "dataLicense": "CC0-1.0",
    "SPDXID": "SPDXRef-DOCUMENT",
    "name": f"juno-{args.version}",
    "documentNamespace": f"{args.repository}/releases/{args.version}/{args.commit}/sbom",
    "creationInfo": {"created": "1970-01-01T00:00:00Z", "creators": ["Tool: scripts/release/metadata.py"]},
    "packages": [package_map[key] for key in sorted(package_map)],
    "relationships": [relationship_map[key] for key in sorted(relationship_map)],
}

resolved_dependencies: list[dict[str, Any]] = [
    {"uri": f"git+{args.repository}@refs/tags/{args.version}", "digest": {"gitCommit": args.commit}},
    {
        "uri": "docker.io/library/golang:1.25.10-alpine3.22",
        "digest": {"sha256": "26b4d7113039cd51356bd7930ecafd1031d2975dc3b6940ec8ed09457e17cf95"},
    },
]
for name, version, checksum, replacement_name, replacement_version, replacement_checksum in sorted(modules):
    original_dependency: dict[str, Any] = {"name": "original Go dependency", "uri": f"pkg:golang/{name}@{version}"}
    if checksum:
        original_dependency["digest"] = {"dirHash1": go_sum_dirhash1(checksum)}
    if replacement_name:
        original_dependency["annotations"] = {"selectedReplacement": f"{replacement_name}@{replacement_version}"}
    resolved_dependencies.append(original_dependency)
    if replacement_name:
        replacement_dependency: dict[str, Any] = {
            "name": f"selected replacement for {name}@{version}",
            "uri": f"pkg:golang/{replacement_name}" + (f"@{replacement_version}" if replacement_version else ""),
            "annotations": {"goOriginal": f"{name}@{version}", "goReplacement": f"{replacement_name}@{replacement_version}"},
        }
        if replacement_checksum:
            replacement_dependency["digest"] = {"dirHash1": go_sum_dirhash1(replacement_checksum)}
        resolved_dependencies.append(replacement_dependency)
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

for output_name, value in (("SBOM.spdx.json", sbom), ("provenance.intoto.jsonl", provenance)):
    (root / output_name).write_text(json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n")
