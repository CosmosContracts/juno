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
Exchange validator balances that failed to vote on prop #37 will be excluded. Additionally unbonded ATOM at the time of the snapshot will be excluded.
A whale cap was voted in by the community, effectively hard-capping $ATOM accounts at 50 thousand $ATOM in order to ensure a less top heavy distribution.
10% of the supply difference to be allocated to a multisig committee address for the initial funding of a core-development team. The remaining 90% of the excess supply to be allocated to the community pool.
(Multi-sig committee to be selected by the community before genesis!)

**Asset & network metrics**

The community has proposed the following parameters for the network and native asset (subject to change before genesis based on community polling):


⚫️ **Ticker**: JUNO

⚫️ **Supply**: Snapshot of Cosmoshub-3 at 06:00 PM UTC on Feb 18th 2021

⚫️ **Inflation**: Fixed yearly inflation (Reward model below)

⚫️ **Community pool tax**: 5% of block rewards

![JUNO BANNER](https://user-images.githubusercontent.com/79812965/114202517-8ce5b180-9957-11eb-842f-584a2d729b2b.png)

**Proposed JUNO inflation reward shedule**

✅ Initial Supply at genesis 268,335,648 $Juno (Stargate Snapshot)

Initial fixed inflation 40% (+ 107.334.259,2)

⚫️ After year 1: 375669907.2 JUNO 

Inflation reduction to 20% (+75.133.981,44)

⚫️ After year 2: 450803888.64 JUNO

Inflation reduction to 10% (+45.080.388,864)

⚫️ After year 3: 495884277.504 JUNO

Once the inflation reaches 10% it reduces on a fixed 1% basis each year.

⚫️ Year 4 = 9% (+44.629.584,975) Supply = 540513862.479 JUNO

⚫️ Year 5 = 8% (+43.241.108,99832) Supply = 583754971.47732 JUNO

⚫️ Year 6 = 7% (+40.862.848,00341) Supply = 624617819.48073 JUNO

⚫️ Year 7 = 6% (+37.477.069,16884) Supply = 662094888.64957 JUNO

⚫️ Year 8 = 5% (+33.104.744,43248) Supply = 695199633.08205 JUNO

⚫️ Year 9 = 4% (+27.807.985,32328) Supply = 723007618.40533 JUNO

⚫️ Year 10 = 3% (+21.690.228,55216) Supply = 744697846.95749 JUNO

⚫️ Year 11 = 2% (+14.893.956,93915) Supply = 759591803.89664 JUNO

⚫️ Year 12 = 1% (+7.595.918,03897) Supply = 767187721.93561 JUNO MAX SUPPLY

After year 12 the inflation reward shedule ends. 
Network incentives would primarily come from smart contract usage & regular tx fees generated on the network.










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
