#!/usr/bin/env python3
"""Create deterministic SPDX and in-toto release metadata without network access."""
import argparse, hashlib, json, os
from pathlib import Path

p = argparse.ArgumentParser()
p.add_argument("--directory", required=True)
p.add_argument("--version", required=True)
p.add_argument("--commit", required=True)
p.add_argument("--repository", required=True)
a = p.parse_args()
root = Path(a.directory)
files = sorted(x for x in root.iterdir() if x.is_file() and
               (x.name.startswith("junod-linux-") or x.name.endswith(".tar.gz")))
def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()
subjects = [{"name": x.name, "digest": {"sha256": sha(x)}} for x in files]

packages = []
for binary in (x for x in files if x.name.startswith("junod-linux-")):
    packages.append({"SPDXID": "SPDXRef-" + binary.name, "name": binary.name,
                     "versionInfo": a.version, "downloadLocation": "NOASSERTION",
                     "filesAnalyzed": False,
                     "licenseConcluded": "NOASSERTION",
                     "licenseDeclared": "NOASSERTION",
                     "copyrightText": "NOASSERTION",
                     "checksums": [{"algorithm": "SHA256", "checksumValue": sha(binary)}],
                     "comment": f"junod {a.version}, commit {a.commit}"})
sbom = {"spdxVersion":"SPDX-2.3", "dataLicense":"CC0-1.0", "SPDXID":"SPDXRef-DOCUMENT",
        "name":f"juno-{a.version}", "documentNamespace":f"{a.repository}/releases/{a.version}/sbom",
        "creationInfo":{"created":"1970-01-01T00:00:00Z", "creators":["Tool: scripts/release/metadata.py"]},
        "packages":packages}
provenance = {"_type":"https://in-toto.io/Statement/v1", "subject":subjects,
 "predicateType":"https://slsa.dev/provenance/v1",
 "predicate":{"buildDefinition":{"buildType":f"{a.repository}/.github/workflows/release.yml",
 "externalParameters":{"version":a.version,"commit":a.commit},
 "resolvedDependencies":[{"uri":f"git+{a.repository}@refs/tags/{a.version}","digest":{"gitCommit":a.commit}},
 {"uri":"docker.io/library/golang:1.25.10-alpine3.22","digest":{"sha256":"26b4d7113039cd51356bd7930ecafd1031d2975dc3b6940ec8ed09457e17cf95"}}]},
 "runDetails":{"builder":{"id":f"{a.repository}/.github/workflows/release.yml@{a.commit}"},"metadata":{"invocationId":os.getenv("GITHUB_RUN_ID", "local")}}}}
for name, value in (("SBOM.spdx.json", sbom), ("provenance.intoto.jsonl", provenance)):
    (root/name).write_text(json.dumps(value, sort_keys=True, separators=(",", ":"))+"\n")