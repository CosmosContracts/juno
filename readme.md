# Juno

![JUNO BANNER TEXT](https://user-images.githubusercontent.com/79812965/114202416-73446a00-9957-11eb-9cfc-cfec3d0b56ea.png)

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

⚫️ Interoperable smart contracts 

⚫️ Modularity

⚫️ Wasm + (EVM)

⚫️ Compilation in multiple languages Rust & Go (C,C++)

⚫️ High scalability

⚫️ Ease of use

⚫️ Fee balancing (Upper & lower bound)

⚫️ Free & fair asset distribution 100% to staked atom only

⚫️ Balanced governance (Zero top heavy control) 
                                                     
⚫️ Value sharing model linked to smart contract usage
                                                  
⚫️ Permissionless 
                                                     
⚫️ Decentralized
                                             
⚫️ Censorship resistant


**Distribution**

A 1:1 stakedrop is distributed to $ATOM holders, giving 100% of the $JUNO supply to staked $ATOM balances that had their assets bonded 
at the time of the Stargate snapshot on Feb. 18th 6:00 PM UTC. 
Addresses that qualify will be included in the JUNO genesis block at launch. 
Exchange validators that failed to vote on prop #37 (Stargate upgrade) will be excluded from the genesis allocation. Including delegators bonding ATOM to those exchange validators. Additionally any unbonded ATOM at the time of the snapshot will be excluded.
A whale cap was voted in by the community, effectively hard-capping $ATOM accounts at 50 thousand $ATOM in order to ensure a less top heavy distribution.
10% of the supply difference to be allocated to a multisig committee address for the funding of a core-development efforts. The remaining 90% of the excess supply to be allocated in the following ways (20 million $Juno community pool, Smart contract competition 2.373.341,66 million to be managed/distributed by the multi-sig committee. The remaining difference will not be included in the genesis file ie. burned)


**Asset & network metrics**

The community has proposed the following parameters for the network and native asset:


⚫️ **Ticker**: JUNO

⚫️ **Supply**: Snapshot of Cosmoshub-3 at 06:00 PM UTC on Feb 18th 2021

⚫️ **Inflation**: Fixed yearly inflation (Reward model below)

⚫️ **Community pool tax**: 5% of block rewards

![JUNO BANNER](https://user-images.githubusercontent.com/79812965/114202517-8ce5b180-9957-11eb-842f-584a2d729b2b.png)

**Tokenomics & reward shedule** (updated on 05.06.2021)


✅ Circulating Supply at genesis 64.903.243,548222 $JUNO (64.9 Million)


**Breakdown**

⚫️ Stakedrop: 30.663.193,590002 $JUNO

⚫️ Community Pool: 20.000.000,00 $JUNO

⚫️ Development Multi-sig Committee: 11.866.708,29852 $JUNO

⚫️ Smart Contract Challenges: 2.373.341,6597 $JUNO


**Reward Schedule**

Initial fixed inflation 40% (+25.961.297,4192)

⚫️ After year 1: 90.864.540,967422 JUNO

Inflation reduction to 20% (+18.172.908,19348)

⚫️ After year 2: 109.037.449,160902 JUNO

Inflation reduction to 10% (+10.903.744,91609)

⚫️ After year 3: 119.941.194,076992 JUNO

Once the inflation reaches 10% it gradually reduces on a fixed 1% basis each year.

⚫️ Year 4 = 9% (+10.794.707,46693) Supply = 130.735.901,543922 JUNO

⚫️ Year 5 = 8% (+10.458.872,12351) Supply = 141.194.773,667432 JUNO

⚫️ Year 6 = 7% (+9.883.634,15672) Supply = 151.078.407,824152 JUNO

⚫️ Year 7 = 6% (+9.064.704,46945) Supply = 160.143.112,293602 JUNO

⚫️ Year 8 = 5% (+8.007.155,61468) Supply = 168.150.267,908282 JUNO

⚫️ Year 9 = 4% (+6.726.010,71633) Supply = 174.876.278,624612 JUNO

⚫️ Year 10 = 3% (+5.246.288,35874) Supply = 180.122.566,983352 JUNO

⚫️ Year 11 = 2% (+3.602.451,33967) Supply = 183.725.018,323022 JUNO


⚫️ Year 12 = 1% (+1.837.250,18323) Supply = 185.562.268,506252 JUNO MAX SUPPLY (185.5 Million)

After year 12 the inflation reward schedule ends. Network incentives would primarily come from smart contract usage & regular tx fees generated on the network.












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

- [Juno](https://junochain.com)
- [Starport](https://github.com/tendermint/starport)
- [Cosmos SDK documentation](https://docs.cosmos.network)
- [Cosmos SDK Tutorials](https://tutorials.cosmos.network)
- [Discord](https://discord.gg/W8trcGV)
