<!--
order: 1
-->

# Authorization

For security reasons, only specific addresses can distribute tokens to $JUNO stakers. We accept any kind of address: multisig, smart contracts, regular and [https://daodao.zone](DAODAO) DAOs. 

Governance can decide wether to approve or deny a new address to be added to the authorized list.

## Query the allowed addresses

You can query the list of allowed addresses directly from x/drip params

```
% junod q drip params --output json
{"enable_drip":true,"allowed_addresses":[]}
```

## Governance proposal

To update the authorized address is possible to create a onchain new proposal. You can use the following example `proposal.json` file

```json
{
 "messages": [
  {
   "@type": "/juno.drip.v1.MsgUpdateParams",
   "authority": "juno10d07y265gmmuvt4z0w9aw880jnsr700jvss730",
   "params": {
    "enable_drip": false,
    "allowed_addresses": ["juno1j0a9ymgngasfn3l5me8qpd53l5zlm9wurfdk7r65s5mg6tkxal3qpgf5se"]
   }
  }
 ],
 "metadata": "{\"title\": \"Allow an amazing contract to distribute tokens using drip\", \"authors\": [\"dimi\"], \"summary\": \"If this proposal passes juno1j0a9ymgngasfn3l5me8qpd53l5zlm9wurfdk7r65s5mg6tkxal3qpgf5se will be added to the authorized addresses of the drip module\", \"details\": \"If this proposal passes juno1j0a9ymgngasfn3l5me8qpd53l5zlm9wurfdk7r65s5mg6tkxal3qpgf5se will be added to the authorized addresses of the drip module\", \"proposal_forum_url\": \"https://commonwealth.im/juno/discussion/9697-juno-protocol-level-defi-incentives\", \"vote_option_context\": \"yes\"}",
 "deposit": "1000ujuno",
 "title": "Allow an amazing contract to distribute tokens using drip",
 "summary": "If this proposal passes juno1j0a9ymgngasfn3l5me8qpd53l5zlm9wurfdk7r65s5mg6tkxal3qpgf5se will be added to the authorized addresses of the drip module"
}
```

It can be submitted with the standard `junod tx gov submit-proposal proposal.json --from yourkey` command.
