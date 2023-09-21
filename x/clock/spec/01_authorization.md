# Authorization

For security purposes, only the governance module can add new contracts to the EndBlocker executes.

## Query contracts

You can query the list of contracts that are 'ticked' every block with the following command:

```bash
    junod q clock contracts --output json
    # {"contract_addresses":[]}
```

## Governance proposal

To update the authorized address is possible to create an onchain new proposal. You can use the following example `proposal.json` file

```json
{
    "messages": [
     {
      "@type": "/juno.clock.v1.MsgUpdateParams",
      "authority": "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730",
      "params": {
       "contract_addresses": [
        "juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8"
       ]
      }
     }
    ],
    "metadata": "{\"title\": \"Allow a new contract to use the x/clock module for our features\", \"authors\": [\"Reece\"], \"summary\": \"If this proposal passes juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 will be added to the authorized addresses of the clock module\", \"details\": \"If this proposal passes juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 will be added to the authorized addresses of the clock module\", \"proposal_forum_url\": \"https://commonwealth.im/juno/discussion/9697-juno-protocol-level-defi-incentives\", \"vote_option_context\": \"yes\"}",
    "deposit": "1000000ujuno",
    "title": "Allow a new contract to use the x/clock module for our features",
    "summary": "If this proposal passes juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 will be allowed to use the x/clock module to perform XYZ actions",
}
```

It can be submitted with the standard `junod tx gov submit-proposal proposal.json --from=YOURKEY` command.
