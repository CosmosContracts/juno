'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const YAML = require('yaml');

const { buildDispatchRequest, validateReleaseTag } = require('./release-dispatch');

const root = path.resolve(__dirname, '../..');
const workflowPath = path.join(root, '.github/workflows/release-dispatch.yml');
const ciPath = path.join(root, '.github/workflows/release-dispatch-ci.yml');
const workflow = YAML.parse(fs.readFileSync(workflowPath, 'utf8'));

const dependencyEnv = {
  JUNO_REPO: 'https://github.com/CosmosContracts/juno.git',
  JUNO_DIR: 'proto',
  COSMOS_SDK_REPO: 'https://github.com/cosmos/cosmos-sdk.git',
  COSMOS_SDK_REV: 'v0.53.7',
  COSMOS_SDK_DIR: 'proto',
  WASMD_REPO: 'https://github.com/CosmWasm/wasmd.git',
  WASMD_REV: 'v0.61.11',
  WASMD_DIR: 'proto',
  COMETBFT_REPO: 'https://github.com/cometbft/cometbft.git',
  COMETBFT_REV: 'v0.38.23',
  COMETBFT_DIR: 'proto',
  IBC_GO_REPO: 'https://github.com/cosmos/ibc-go.git',
  IBC_GO_REV: 'v10.6.0',
  IBC_GO_DIR: 'proto',
  ICS23_REPO: 'https://github.com/cosmos/ics23.git',
  ICS23_REV: 'go/v0.11.0',
  ICS23_DIR: 'proto',
};

function request(payload) {
  return buildDispatchRequest(payload, dependencyEnv);
}

test('builds a payload from a published release event', () => {
  const result = request({
    action: 'published',
    release: { tag_name: 'v31.0.0', draft: false, prerelease: false },
  });

  assert.equal(result.owner, 'CosmosContracts');
  assert.equal(result.repo, 'juno-std');
  assert.equal(result.event_type, 'juno-release');
  assert.deepEqual(result.client_payload, {
    is_draft: false,
    is_prerelease: false,
    release_tag: 'v31.0.0',
    repos: {
      juno: {
        name: 'juno', repo: dependencyEnv.JUNO_REPO, rev: 'v31.0.0', dir: 'proto', exclude_mods: [],
      },
      cosmos_sdk: {
        name: 'cosmos', repo: dependencyEnv.COSMOS_SDK_REPO, rev: 'v0.53.7', dir: 'proto',
        exclude_mods: ['cosmos/benchmark', 'cosmos/counter', 'cosmos/epochs', 'cosmos/protocolpool'],
      },
      wasmd: {
        name: 'wasm', repo: dependencyEnv.WASMD_REPO, rev: 'v0.61.11', dir: 'proto', exclude_mods: [],
      },
      cometbft: {
        name: 'cometbft', repo: dependencyEnv.COMETBFT_REPO, rev: 'v0.38.23', dir: 'proto', exclude_mods: [],
      },
      ibc_go: {
        name: 'ibc-go', repo: dependencyEnv.IBC_GO_REPO, rev: 'v10.6.0', dir: 'proto', exclude_mods: [],
      },
      ics23: {
        name: 'ics23', repo: dependencyEnv.ICS23_REPO, rev: 'go/v0.11.0', dir: 'proto', exclude_mods: [],
      },
    },
  });
});

test('preserves prerelease flags from a published release payload', () => {
  const result = request({
    action: 'published',
    release: { tag_name: 'v31.0.0-rc.1', draft: false, prerelease: true },
  });

  assert.equal(result.client_payload.release_tag, 'v31.0.0-rc.1');
  assert.equal(result.client_payload.is_draft, false);
  assert.equal(result.client_payload.is_prerelease, true);
});

test('manual dispatch uses the release resolved by the workflow API lookup', () => {
  const result = request({
    inputs: { release_tag: 'v32.1.0-rc.2' },
    release: { tag_name: 'v32.1.0-rc.2', draft: false, prerelease: true },
  });

  assert.equal(result.client_payload.release_tag, 'v32.1.0-rc.2');
  assert.equal(result.client_payload.is_draft, false);
  assert.equal(result.client_payload.is_prerelease, true);
});

test('accepts Juno semantic-version tags including release candidates', () => {
  for (const tag of ['v31.0.0', 'v31.0.0-rc.1', 'v32.4.5-beta.2+build.7']) {
    assert.equal(validateReleaseTag(tag), tag);
  }
});

test('rejects branches, refs, and arbitrary strings as release tags', () => {
  for (const tag of [
    '', 'main', 'release/v31', 'refs/tags/v31.0.0', 'v31', 'v31.0', 'v31.01.0',
    'v31.0.0-', 'v31.0.0-rc..1', 'v31.0.0 rc1', '../v31.0.0',
  ]) {
    assert.throws(() => validateReleaseTag(tag), /Juno release tag/i, JSON.stringify(tag));
  }
});

test('workflow listens only for release published and derives release state from payload', () => {
  assert.deepEqual(workflow.on.release.types, ['published']);
  assert.ok(workflow.on.workflow_dispatch.inputs.release_tag.required);
  assert.deepEqual(Object.keys(workflow.on.workflow_dispatch.inputs), ['release_tag']);
});

test('manual workflow validates tag existence before dispatch', () => {
  const script = workflow.jobs.dispatch.steps.find((step) => step.name === 'Dispatch release event').with.script;
  const validation = script.indexOf('validateReleaseTag(payload.inputs.release_tag)');
  const lookup = script.indexOf('getReleaseByTag');
  const dispatch = script.indexOf('createDispatchEvent');
  assert.ok(validation >= 0 && lookup > validation, 'tag format must be validated before API lookup');
  assert.ok(lookup >= 0, 'manual path must resolve an existing GitHub release');
  assert.ok(dispatch > lookup, 'release lookup must occur before repository dispatch');
  assert.match(script, /context\.repo/);
});

test('workflow checks out and executes the exact trusted workflow revision', () => {
  assert.deepEqual(workflow.permissions, { contents: 'read' });
  const checkout = workflow.jobs.dispatch.steps.find((step) => String(step.uses || '').startsWith('actions/checkout@'));
  assert.match(checkout.uses, /^actions\/checkout@[0-9a-f]{40}$/);
  assert.equal(checkout.with.ref, '${{ github.workflow_sha }}');
  assert.equal(checkout.with['persist-credentials'], false);
  assert.match(workflow.jobs.dispatch.steps.at(-1).uses, /^actions\/github-script@[0-9a-f]{40}$/);
});

test('workflow dependency revisions track the selected v31 Go dependency graph', () => {
  const goMod = fs.readFileSync(path.join(root, 'go.mod'), 'utf8');
  const moduleVersions = new Map(
    [...goMod.matchAll(/^\s*(\S+)\s+(v\S+)(?:\s+\/\/.*)?$/gm)].map((match) => [match[1], match[2]]),
  );
  const expected = {
    COSMOS_SDK_REV: moduleVersions.get('github.com/cosmos/cosmos-sdk'),
    WASMD_REV: moduleVersions.get('github.com/CosmWasm/wasmd'),
    COMETBFT_REV: moduleVersions.get('github.com/cometbft/cometbft'),
    IBC_GO_REV: moduleVersions.get('github.com/cosmos/ibc-go/v10'),
    ICS23_REV: `go/${moduleVersions.get('github.com/cosmos/ics23/go')}`,
  };

  for (const [name, revision] of Object.entries(expected)) {
    assert.ok(revision && !revision.includes('undefined'), `${name} dependency must exist in go.mod`);
    assert.equal(workflow.env[name], revision);
  }
});

test('dedicated CI runs offline Node tests and actionlint for dispatch files', () => {
  const ci = YAML.parse(fs.readFileSync(ciPath, 'utf8'));
  assert.deepEqual(ci.permissions, { contents: 'read' });
  const allSteps = Object.values(ci.jobs).flatMap((job) => job.steps);
  assert.ok(allSteps.some((step) => step.run && step.run.includes('npm ci')));
  assert.ok(allSteps.some((step) => step.run && step.run.includes('npm test')));
  assert.ok(allSteps.some((step) => step.run && step.run.includes('actionlint')));
  for (const step of allSteps.filter((candidate) => candidate.uses)) {
    assert.match(step.uses, /^[^@]+@[0-9a-f]{40}$/, `${step.uses} must use an immutable SHA`);
  }
});

test('offline payload builder cannot perform network lookup or dispatch', () => {
  const builder = fs.readFileSync(path.join(__dirname, 'release-dispatch.js'), 'utf8');
  assert.doesNotMatch(builder, /createDispatchEvent|getReleaseByTag|https?:/);
});
