<!--
order: 0
title: "Drip Overview"
parent:
  title: "drip"
-->

# `drip`

## Abstract

This document specifies the internal `x/drip` module of Juno Network.

The `x/drip` allows specific addresses (usually smart contracts) to send tokens to the fee_pool module in order to perform a live airdrop to Juno Stakers.

It consists only on one new message `MsgDistributeTokens`, when called from an authorized address all the funds sent with it are distributed at the next block. 

On an ideal scenario, projects are allocating tokens to a smart contract that then split the amount over a custom schedule, using for example [https://www.cron.cat/](CronCat).

## Contents

1. **[Authorization](01_authorization.md)**
2. **[Distribute Tokens](02_distribute_tokens.md)**
3. **[Example Contract](03_example.md)**