# Juno

![photo_2021-03-03_18-10-06](https://user-images.githubusercontent.com/79812965/110179648-ba938400-7e08-11eb-849c-a8da02c31a21.jpg)

An **interoperable smart contract hub** which automatically executes, controls or documents a procedure of relevant events and actions 
according to the terms of such contract or agreement to be valid & usable across multiple sovereign networks.

Juno as a **sovereign public blockchain** in the Cosmos ecosystem, aims to provide a sandbox environment for the deployment 
of such interoperable smart contracts. The network serves as a **decentralized, permissionless & censorship resistant** avenue 
for developers to efficiently and securely launch application specific smart contracts using proven frameworks 
and compile them in various languages **Rust & Go** with the potential future addition of C and C++
Battle tested modules such as CosmWasm, will allow for decentralized applications (dapps) to be compiled on robust and secure multi-chain smart contracts.
EVM support and additional specialized modules to be enabled after genesis subject to onchain governance.

At the heart of Cosmos is the Inter Blockchain Communication Protocol (IBC), which sets the table for an interoperable base layer 0 
to now be used to transfer data packets across thousands of independent networks supporting IBC. 
Naturally, the next evolutionary milestone is to enable cross-network smart contracts.

The Juno blockchain is built using the **Cosmos SDK framework**. 
A generalized framework that simplifies the process of building secure blockchain applications on top of Tendermint BFT. 
It is based on two major principles: Modularity & capabilities-based security.

Agreement on the network is reached via **Tendermint BFT consensus**.

Tendermint BFT is a solution that packages the networking and consensus layers of a blockchain into a generic engine, 
allowing developers to focus on application development as opposed to the complex underlying protocol. 
As a result, Tendermint saves hundreds of hours of development time.

Juno originates from a **community driven initiative**, prompted by developers, validators &delegators.
A common vision is to preserve the neutrality, performance & reliability of the Cosmos Hub and offload smart contract deployment to a dedicated sister Hub. 
Juno plans to make an early connection to the Cosmos Hub enabling IBC transfers, cross-chain smart contracts and making use of shared security.

An independent set of validators secures the Juno main-net via delegated proof of stake. 
$Juno, the native asset has many functions like securing the Juno Hub & serving as a work token to give access to on-chain governance voting rights 
and to provide utility in the deployment and execution of smart contracts.


**What differentiates JUNO from other Smart Contract networks?**

🟣 Interoperable smart contracts 

🟣 Modularity

🟣 Wasm + (EVM)

🟣 Compilation in multiple languages Rust & Go (C,C++)

🟣 High scalability

🟣 Ease of use

🟣 Fee balancing (Upper & lower bound)

🟣 Free & fair asset distribution 100% to staked atom only

🟣 Balanced governance (Zero top heavy control) 
                                                     
🟣 Value sharing model linked to smart contract usage
                                                  
🟣 Permissionless 
                                                     
🟣 Decentralized
                                             
🟣 Censorship resistant

![photo_2021-03-03_18-22-29](https://user-images.githubusercontent.com/79812965/110187125-8a9fad00-7e17-11eb-971f-1b56faf3e558.jpg)


🟣 **Distribution** 🟣

A 1:1 stakedrop is distributed to $ATOM holders, giving 100% of the $JUNO supply to staked $ATOM balances that had their assets bonded 
at the time of the Stargate snapshot on Feb. 18th 6:00 PM UTC. 
Addresses that qualify will be included in the JUNO genesis block at launch. 
Exchange validator balances that failed to vote on prop #37 will be excluded. Additionally unbonded ATOM at the time of the snapshot will be excluded.
A whale cap was voted in by the community, effectively hard-capping $ATOM accounts at 50 thousand $ATOM in order to ensure a less top heavy distribution.
10% of the supply difference to be allocated to a multisig committee address for the initial funding of a core-development team. The remaining 90% of the excess supply to be allocated to the community pool.
(Multi-sig committee to be selected by the community before genesis!)

🟣 **Asset & network metrics** 🟣

The community has proposed the following parameters for the network and native asset (subject to change before genesis based on community polling):


🟣 **Ticker**: JUNO

🟣 **Supply**: Snapshot of Cosmoshub-3 at 06:00 PM UTC on Feb 18th 2021

🟣 **Inflation**: Dynamic 15-40%

🟣 **Community pool tax**: 5% of block rewards

![JUNO (LOGO 3)](https://user-images.githubusercontent.com/79812965/110180550-4b1e9400-7e0a-11eb-8386-ef1f5a9252f9.png)

🟣 **Game theory & value sharing economy** 🟣

The JUNO community proposes an initial dynamic inflation of 15-40%. 15% being the lower bound and 40% the upper bound. 
As the bonded rate drops below 66% the inflation would slowly increase towards the upper bound. 
While a move above the 66% mark would slowly move inflation toward the lower bound.
A relatively high initial inflation may incentivize a higher bonded rate and therefore stability/security of the network.

In addition to the inflation economics a value sharing model is proposed in direct connection with the deployment of smart contracts. 
One that mimics a closed value loop. The primary target being to incentivize validators, delegators, community pool 
and additionally include a token burn mechanism to offset the much needed initially high inflation rewards.

When a smart contract is deployed on the Juno Network, the amount of $JUNO used for deployment is not initially locked but split 3 ways
(subject to change based on community polling!):

🟣 50% shared with validators & delegators 

🟣 20% allocated to the community pool 
  
🟣 30% permanent token burn 

This value split happens during the early agreement/deployment stage of every smart contract deployment on the Juno Hub.
Such an implementation would require a slight modification to the Cosmos SDK itself. This could be performed either 
before genesis or later on introduced by onchain governance.








**Juno** is a blockchain built using Cosmos SDK and Tendermint and created with [Starport](https://github.com/tendermint/starport).

## Get started

```
starport serve
```

`serve` command installs dependencies, builds, initializes and starts your blockchain in development.

## Configure

Your blockchain in development can be configured with `config.yml`. To learn more see the [reference](https://github.com/tendermint/starport#documentation).

## Launch

To launch your blockchain live on mutliple nodes use `starport network` commands. Learn more about [Starport Network](https://github.com/tendermint/spn).

## Learn more

- [Starport](https://github.com/tendermint/starport)
- [Cosmos SDK documentation](https://docs.cosmos.network)
- [Cosmos SDK Tutorials](https://tutorials.cosmos.network)
- [Discord](https://discord.gg/W8trcGV)
