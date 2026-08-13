'use strict';

const TARGET_REPOSITORY = 'CosmosContracts/juno-std';
const EVENT_TYPE = 'juno-release';

function validateReleaseTag(tag) {
  // Juno release tags use SemVer with a mandatory leading "v". Numeric
  // prerelease identifiers may not contain leading zeroes (SemVer 2.0.0).
  const semver = /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-(?:(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$/;
  if (typeof tag !== 'string' || !semver.test(tag)) {
    throw new Error(`Invalid Juno release tag: '${tag || ''}'`);
  }

  return tag;
}

function requiredEnv(env, name) {
  const value = env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

function buildDispatchRequest(payload, env) {
  const release = payload && payload.release;
  if (!release) {
    throw new Error('A resolved GitHub release payload is required');
  }
  const releaseTag = validateReleaseTag(release.tag_name);
  const isDraft = Boolean(release.draft);
  const isPrerelease = Boolean(release.prerelease);
  const [owner, repo] = TARGET_REPOSITORY.split('/');

  const repos = {
    juno: {
      name: 'juno',
      repo: requiredEnv(env, 'JUNO_REPO'),
      rev: releaseTag,
      dir: requiredEnv(env, 'JUNO_DIR'),
      exclude_mods: [],
    },
    cosmos_sdk: {
      name: 'cosmos',
      repo: requiredEnv(env, 'COSMOS_SDK_REPO'),
      rev: requiredEnv(env, 'COSMOS_SDK_REV'),
      dir: requiredEnv(env, 'COSMOS_SDK_DIR'),
      exclude_mods: ['cosmos/benchmark', 'cosmos/counter', 'cosmos/epochs', 'cosmos/protocolpool'],
    },
    wasmd: {
      name: 'wasm',
      repo: requiredEnv(env, 'WASMD_REPO'),
      rev: requiredEnv(env, 'WASMD_REV'),
      dir: requiredEnv(env, 'WASMD_DIR'),
      exclude_mods: [],
    },
    cometbft: {
      name: 'cometbft',
      repo: requiredEnv(env, 'COMETBFT_REPO'),
      rev: requiredEnv(env, 'COMETBFT_REV'),
      dir: requiredEnv(env, 'COMETBFT_DIR'),
      exclude_mods: [],
    },
    ibc_go: {
      name: 'ibc-go',
      repo: requiredEnv(env, 'IBC_GO_REPO'),
      rev: requiredEnv(env, 'IBC_GO_REV'),
      dir: requiredEnv(env, 'IBC_GO_DIR'),
      exclude_mods: [],
    },
    ics23: {
      name: 'ics23',
      repo: requiredEnv(env, 'ICS23_REPO'),
      rev: requiredEnv(env, 'ICS23_REV'),
      dir: requiredEnv(env, 'ICS23_DIR'),
      exclude_mods: [],
    },
  };

  return {
    owner,
    repo,
    event_type: EVENT_TYPE,
    client_payload: {
      is_draft: isDraft,
      is_prerelease: isPrerelease,
      release_tag: releaseTag,
      repos,
    },
  };
}

module.exports = { buildDispatchRequest, validateReleaseTag };
