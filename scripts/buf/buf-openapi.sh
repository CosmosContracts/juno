#!/usr/bin/env sh
set -eo pipefail

buf dep update
buf generate --template buf.gen.openapi.yaml
buf generate --template buf.gen.openapi-external.yaml

yq eval -i \
  '.paths |= with_entries(select(.key | test("/cosmos/mint/") | not))' \
  ./gen/external/openapi.yaml

yq eval-all 'select(fileIndex == 0) *+ select(fileIndex == 1)' \
  ./gen/internal/openapi.yaml ./gen/external/openapi.yaml \
  >./gen/openapi.yaml

cd gen

yq eval -i 'del(.tags)' openapi.yaml
yq eval -i 'del(.paths[][].tags)' openapi.yaml

yq eval '.paths | keys | .[]' openapi.yaml | while IFS= read -r path; do
  module=$(printf '%s' "$path" | cut -d/ -f3) # "clock"

  case "$path" in
  # ------ Juno: FeePay ------
  "/juno/feepay/v1/contract/{contractAddress}")
    opName="contract"
    ;;

  "/juno/feepay/v1/contract/{contractAddress}/uses/{walletAddress}")
    opName="uses"
    ;;

  "/juno/feepay/v1/contract/{contractAddress}/eligible/{walletAddress}")
    opName="eligible"
    ;;

  # ------ Cosmos: Auth ------
  "/cosmos/auth/v1beta1/accounts/{address}")
    opName="account"
    ;;

  "/cosmos/auth/v1beta1/account_info/{address}")
    opName="account_info"
    ;;

  "/cosmos/auth/v1beta1/address_by_id/{id}")
    opName="address_by_id"
    ;;

  "/cosmos/auth/v1beta1/bech32/{addressBytes}")
    opName="bech32_bytes"
    ;;

  "/cosmos/auth/v1beta1/bech32/{addressBytes}")
    opName="bech32_string"
    ;;

  "/cosmos/auth/v1beta1/module_accounts/{name}:")
    opName="module_accounts"
    ;;

  # ------ Cosmos: Authz ------
  "/cosmos/authz/v1beta1/grants/grantee/{grantee}")
    opName="grants_by_grantee"
    ;;

  "/cosmos/authz/v1beta1/grants/granter/{granter}")
    opName="grants_by_granter"
    ;;

  # ------ Cosmos: Bank ------
  "/cosmos/bank/v1beta1/denoms_metadata/{denom}")
    opName="denom_metadata"
    ;;

  "/cosmos/bank/v1beta1/spendable_balances/{address}/by_denom")
    opName="spendable_balances_by_denom"
    ;;

  "/cosmos/bank/v1beta1/spendable_balances/{address}")
    opName="spendable_balances"
    ;;

  "/cosmos/bank/v1beta1/supply/by_denom")
    opName="supply_by_denom"
    ;;

  "/cosmos/bank/v1beta1/balances/{address}")
    opName="balances"
    ;;

  "/cosmos/bank/v1beta1/denom_owners/{denom}")
    opName="denom_owners"
    ;;

  # ----- Cosmos: Base ------
  "/cosmos/base/tendermint/v1beta1/validatorsets/{height}")
    opName="validatorsets_height"
    ;;

  "/cosmos/base/tendermint/v1beta1/validatorsets/latest")
    opName="validatorsets_latest"
    ;;

  "/cosmos/base/tendermint/v1beta1/blocks/{height}")
    opName="block_height"
    ;;

  # ------ Cosmos: Circuit ------
  "/cosmos/circuit/v1/accounts/{address}")
    opName="account"
    ;;

  # ------ Cosmos: Distribution ------
  "/cosmos/distribution/v1beta1/validators/{validatorAddress}")
    opName="validator_rewards"
    ;;

  "/cosmos/distribution/v1beta1/delegators/{delegatorAddress}/rewards/{validatorAddress}")
    opName="delegator_rewards_by_validator"
    ;;

  # ------ Cosmos: Evidence ------
  "/cosmos/evidence/v1beta1/evidence/{hash}")
    opName="hash"
    ;;

  # ------ Cosmos: Feegrant ------
  "/cosmos/feegrant/v1beta1/allowances/{grantee}")
    opName="allowances"
    ;;

  "/cosmos/feegrant/v1beta1/allowance/{granter}/{grantee}")
    opName="allowance"
    ;;

  "/cosmos/feegrant/v1beta1/issued/{granter}")
    opName="issued_by_granter"
    ;;

  # ------ Cosmos: Gov v1 ------
  "/cosmos/gov/v1/proposals/{proposalId}")
    opName="proposal"
    ;;

  "/cosmos/gov/v1/proposals/{proposalId}/deposits/{depositor}")
    opName="proposal_deposit_by_depositor"
    ;;

  "/cosmos/gov/v1/proposals/{proposalId}/votes/{voter}")
    opName="proposal_vote_by_voter"
    ;;

  # ------ Cosmos: Gov v1beta1 ------
  "/cosmos/gov/v1beta1/params/{paramsType}")
    opName="v1beta1_params"
    ;;

  "/cosmos/gov/v1beta1/proposals/{proposalId}/deposits/{depositor}")
    opName="v1beta1_proposal_deposit_by_depositor"
    ;;

  "/cosmos/gov/v1beta1/proposals/{proposalId}/deposits")
    opName="v1beta1_proposal_deposits"
    ;;

  "/cosmos/gov/v1beta1/proposals/{proposalId}")
    opName="v1beta1_proposal"
    ;;

  "/cosmos/gov/v1beta1/proposals")
    opName="v1beta1_proposals"
    ;;

  "/cosmos/gov/v1beta1/proposals/{proposalId}/tally")
    opName="v1beta1_proposal_tally"
    ;;

  "/cosmos/gov/v1beta1/proposals/{proposalId}/votes/{voter}")
    opName="v1beta1_proposal_vote_by_voter"
    ;;

  "/cosmos/gov/v1beta1/proposals/{proposalId}/votes")
    opName="v1beta1_proposal_votes"
    ;;

  "/cosmos/gov/v1/params/{paramsType}")
    opName="params"
    ;;

  # ------ Cosmos: Group------
  "/cosmos/group/v1/group_members/{groupId}")
    opName="members"
    ;;

  "/cosmos/group/v1/group_policies_by_group/{groupId}")
    opName="policies_by_group"
    ;;

  "/cosmos/group/v1/groups_by_admin/{admin}")
    opName="groups_by_admin"
    ;;

  "/cosmos/group/v1/groups_by_member/{address}")
    opName="groups_by_member"
    ;;

  "/cosmos/group/v1/proposals_by_group_policy/{address}")
    opName="proposals_by_group_policy"
    ;;

  "/cosmos/group/v1/votes_by_proposal/{proposalId}")
    opName="votes_by_proposal"
    ;;

  "/cosmos/group/v1/votes_by_voter/{voter}")
    opName="votes_by_voter"
    ;;

  "/cosmos/group/v1/group_info/{groupId}")
    opName="info"
    ;;

  "/cosmos/group/v1/group_policies_by_admin/{admin}")
    opName="policies_by_admin"
    ;;

  "/cosmos/group/v1/group_policy_info/{address}")
    opName="policy_info_by_address"
    ;;

  "/cosmos/group/v1/proposal/{proposalId}")
    opName="proposal"
    ;;

  "/cosmos/group/v1/vote_by_proposal_voter/{proposalId}/{voter}")
    opName="vote_by_proposal_voter"
    ;;

  # ------ Cosmos: NFT ------
  "/cosmos/nft/v1beta1/classes/{classId}")
    opName="class"
    ;;

  "/cosmos/nft/v1beta1/owner/{classId}/{id}")
    opName="owner"
    ;;

  "/cosmos/nft/v1beta1/supply/{classId}")
    opName="supply"
    ;;

  "/cosmos/nft/v1beta1/balance/{owner}/{classId}")
    opName="balance"
    ;;

  "/cosmos/nft/v1beta1/nfts/{classId}/{id}")
    opName="by_class_and_id"
    ;;

  # ------ Cosmos: Slashing ------
  "/cosmos/slashing/v1beta1/signing_infos/{consAddress}")
    opName="signing_infos_by_cons_address"
    ;;

  # ------ Cosmos: Staking ------

  "/cosmos/staking/v1beta1/delegations/{delegatorAddr}")
    opName="delegations_by_delegator"
    ;;

  "/cosmos/staking/v1beta1/validators/{validatorAddr}/delegations/{delegatorAddr}")
    opName="validator_delegation_by_delegator"
    ;;

  "/cosmos/staking/v1beta1/validators/{validatorAddr}/unbonding_delegations")
    opName="validator_unbonding_delegations"
    ;;

  "/cosmos/staking/v1beta1/validators/{validatorAddr}")
    opName="validator"
    ;;

  "/cosmos/staking/v1beta1/historical_info/{height}")
    opName="historical_info"
    ;;

  "/cosmos/staking/v1beta1/delegators/{delegatorAddr}/validators/{validatorAddr}")
    opName="validator_by_delegator"
    ;;

  "/cosmos/staking/v1beta1/validators")
    opName="all_validators"
    ;;

  # ------ Cosmos: Upgrade ------
  "/cosmos/upgrade/v1beta1/applied_plan/{name}")
    opName="applied_plan"
    ;;

  "/cosmos/upgrade/v1beta1/upgraded_consensus_state/{lastHeight}")
    opName="upgraded_consensus_state"
    ;;

  *)
    opName="${path##*/}"
    ;;
  esac

  for method in get; do
    # only touch if the method exists
    if yq eval ".paths[\"$path\"].$method" openapi.yaml | grep -qv '^null$'; then
      newOpId="${module}_${opName}"

      # set operationId
      yq eval -i \
        '.paths["'"$path"'"].'"$method"'.operationId = "'"$newOpId"'"' \
        openapi.yaml

      # set the tag array
      yq eval -i \
        '.paths["'"$path"'"].'"$method"'.tags = ["'"$module"'"]' \
        openapi.yaml
    fi
  done
done

# collect all tags from the paths, remove duplicates and add them to the tags array
yq eval -i '
  .tags = (
    [ .paths.*.*.tags[] ]
    | unique
    | map({"name": ., "description": ""})
  )
' openapi.yaml

# filter out any path whose key matches "/tx/" to remove incorrect post methods (they dont work this way on cosmos)
yq eval -i '
  .paths |= with_entries(select(.key | test("/tx/") | not))
' openapi.yaml

tail -n +4 openapi.yaml >tmp && mv tmp openapi.yaml

yq eval -i '
  .tags[].name |= (
    capture("(?<first>.)(?<rest>.*)")
    | .first |= upcase
    | .first + .rest
  ) |

  .paths[][].tags |= map(
    capture("(?<first>.)(?<rest>.*)")
    | .first |= upcase
    | .first + .rest
  )
' openapi.yaml

mv openapi.yaml ../app/openapi.yaml

cd ..
rm -rf gen

echo "✅ OpenAPI spec generated at ./app/openapi.yaml"
