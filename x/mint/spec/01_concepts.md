<!--
order: 0
-->

# Concepts

## The Minting Mechanism

The minting mechanism was designed to:
<<<<<<< HEAD
 - allow for a flexible inflation rate determined by market demand targeting a particular bonded-stake ratio
 - effect a balance between market liquidity and staked supply

In order to best determine the appropriate market rate for inflation rewards, a
moving change rate is used.  The moving change rate mechanism ensures that if
the % bonded is either over or under the goal %-bonded, the inflation rate will
adjust to further incentivize or disincentivize being bonded, respectively. Setting the goal
%-bonded at less than 100% encourages the network to maintain some non-staked tokens
which should help provide some liquidity.

It can be broken down in the following way: 
 - If the inflation rate is below the goal %-bonded the inflation rate will
   increase until a maximum value is reached
 - If the goal % bonded (67% in Cosmos-Hub) is maintained, then the inflation
   rate will stay constant 
 - If the inflation rate is above the goal %-bonded the inflation rate will
   decrease until a minimum value is reached
=======
 - allow for a inflation rate determined by Juno Tokenemics

It can be broken down in the following way: 
- Phase 1: Fixed inflation 40%
- Phase 2: Fixed inflation 20%
- Phase 3: Fixed inflation 10%
- Phase 4: Fixed inflation 9%
- Phase 5: Fixed inflation 8%
- Phase 6: Fixed inflation 7%
- Phase 7: Fixed inflation 6%
- Phase 8: Fixed inflation 5%
- Phase 9: Fixed inflation 4%
- Phase 10: Fixed inflation 3%
- Phase 11: Fixed inflation 2%
- Phase 12: Fixed inflation 1%
>>>>>>> disperze/mint-module
