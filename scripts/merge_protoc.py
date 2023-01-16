#!/usr/bin/env python3

# Call this from the ./scripts/protoc_swagger_openapi_gen.sh script

# merged protoc definitions together into 1 JSOB file without duplicate keys
# this is done AFTER swagger-merger has been run, merging the multiple name-#.json files into 1.

import json
import os

current_dir = os.path.dirname(os.path.realpath(__file__))
project_root = os.path.dirname(current_dir)

all_dir = os.path.join(project_root, "tmp-swagger-gen", "_all")

# What we will save when all combined
output = {
    "swagger": "2.0",
    "info": {"title": "Juno network", "version": "v12"},
    "consumes": ["application/json"],
    "produces": ["application/json"],
    "paths": {},
    "definitions": {},
}

# combine all individual files calls into 1 massive file.
# Duplicate key references are removed.
for file in os.listdir(all_dir):
    if not file.endswith(".json"):
        continue

    # read file all_dir / file
    with open(os.path.join(all_dir, file), "r") as f:
        data = json.load(f)

    for key in data["paths"]:
        output["paths"][key] = data["paths"][key]

    for key in data["definitions"]:
        output["definitions"][key] = data["definitions"][key]


# save output into 1 big json file
with open(
    os.path.join(project_root, "tmp-swagger-gen", "_all", "FINAL.json"), "w"
) as f:
    json.dump(output, f, indent=2)
