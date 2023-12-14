<!--
order: 2
-->

# Distributing Tokens

Once an address is authorized, at any time it can use the `MsgDistributToken` message to distribute all the attached funds at the next block.

From command line is as easy as running the following instruction

```
junod tx drip distribute-tokens 100000tf/yourcontract/yourtoken
```

Only native tokens and the ones made with tokenfactory are allowed.

If you have a CW-20 token, you can wrap it to native using [https://github.com/CosmosContracts/tokenfactory-contracts/tree/main/contracts/migrate](this contract).