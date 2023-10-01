# Juno Staking Hooks

This module allows smart contracts to execute logic when staking / validator events occur without any custom bindings.

[Juno Staking Hooks Spec](./spec/README.md)

---

junod tx cw-hooks register-staking juno1qsrercqegvs4ye0yqg93knv73ye5dc3prqwd6jcdcuj8ggp6w0us66deup --from juno1 --keyring-backend=test --home $HOME/.juno1/ --chain-id local-1 --fees 500ujuno --yes
junod tx cw-hooks register-staking juno15u3dt79t6sxxa3x3kpkhzsy56edaa5a66wvt3kxmukqjz2sx0hes5sn38g --from juno1 --keyring-backend=test --home $HOME/.juno1/ --chain-id local-1 --fees 500ujuno --yes

junod q cw-hooks staking-contracts
