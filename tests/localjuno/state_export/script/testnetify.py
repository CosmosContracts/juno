import os
import json
import argparse
from datetime import datetime
from dataclasses import dataclass

# Classes 
@dataclass
class Validator:
    moniker: str
    pubkey: str
    hex_address: str
    operator_address: str
    consensus_address: str

@dataclass
class Account:
    pubkey: str
    address: str

# /cosmos.auth.v1beta1.ModuleAccount in genesis state export "distribution" & "bonded_tokens_pool"
DISTRIBUTION_MODULE_ADDRESS = "juno1jv65s3grqf6v6jl3dp4t6c9t9rk99cd83d88wr"
BONDED_TOKENS_POOL_MODULE_ADDRESS = "juno1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3rf257t"

config = {
    "governance_voting_period": "180s",
}

def replace(d, old_value, new_value):
    """
    Replace all the occurrences of `old_value` with `new_value`
    in `d` dictionary
    """
    for k in d.keys():
        if isinstance(d[k], dict):
            replace(d[k], old_value, new_value)
        elif isinstance(d[k], list):
            for i in range(len(d[k])):
                if isinstance(d[k][i], dict) or isinstance(d[k][i], list):
                    replace(d[k][i], old_value, new_value)
                else:
                    if d[k][i] == old_value:
                        d[k][i] = new_value
        else:
            if d[k] == old_value:
                d[k] = new_value

def replace_validator(genesis, old_validator: Validator, new_validator: Validator):
    replace(genesis, old_validator.hex_address, new_validator.hex_address)
    replace(genesis, old_validator.consensus_address, new_validator.consensus_address)
    
    replace(genesis, old_validator.pubkey, new_validator.pubkey)
    for validator in genesis["validators"]:
        if validator['name'] == old_validator.moniker:
            validator['pub_key']['value'] = new_validator.pubkey
        
    for validator in genesis['app_state']['staking']['validators']:
        if validator['description']['moniker'] == old_validator.moniker:
            validator['consensus_pubkey']['key'] = new_validator.pubkey

    # This creates problems. TODO: change somewhere else in state as well? baseapp crash
    # replace(genesis, old_validator.operator_address, new_validator.operator_address)    

def replace_account(genesis, old_account: Account, new_account: Account):
    replace(genesis, old_account.address, new_account.address)
    replace(genesis, old_account.pubkey, new_account.pubkey)

def create_parser():
    parser = argparse.ArgumentParser(
    formatter_class=argparse.RawDescriptionHelpFormatter,
    description='Create a testnet from a state export')

    parser.add_argument(
        '-c',
        '--chain-id',
        type = str,
        default="localjuno",
        help='Chain ID for the testnet \nDefault: localjuno\n'
    )

    parser.add_argument(
        '-i',
        '--input',
        type = str,
        default="state_export.json",
        dest='input_genesis',
        help='Path to input genesis'
    )

    parser.add_argument(
        '-o',
        '--output',
        type = str,
        default="testnet_genesis.json",
        dest='output_genesis',
        help='Path to output genesis'
    )
    
    parser.add_argument(
        '--validator-hex-address',
        type = str,
        help='Validator hex address to replace'
    )

    parser.add_argument(
        '--validator-operator-address',
        type = str,
        help='Validator operator address to replace'
    )

    parser.add_argument(
        '--validator-consensus-address',
        type = str,
        help='Validator consensus address to replace'
    )

    parser.add_argument(
        '--validator-pubkey',
        type = str,
        help='Validator pubkey to replace'
    )

    parser.add_argument(
        '--account-address',
        type = str,
        help='Account address to replace'
    )

    parser.add_argument(
        '--account-pubkey', 
        type = str,        
        help='The accounts public key'
    )

    parser.add_argument(
        '-q',
        '--quiet',        
        action='store_false',
        help='Less verbose output'
    )

    parser.add_argument(
        '--prune-ibc', 
        action='store_true',
        help='Prune the IBC module'
    )

    parser.add_argument(
        '--pretty-output', 
        action='store_true',
        help='Properly indent output genesis (increases time and file size)'
    )

    return parser

def main():
    parser = create_parser()
    args = parser.parse_args()

    new_validator = Validator(
        moniker = "val",
        pubkey = args.validator_pubkey,
        hex_address = args.validator_hex_address,
        operator_address = args.validator_operator_address,
        consensus_address = args.validator_consensus_address
    )

    old_validator = Validator(
        # from ping.pub
        moniker = "notional",
        pubkey = "ux/IM9uD+a/4rhIurbRiudh9K+M6tH1cNfffpX48Lrw=", 
        hex_address = "6EC804DBB72380D0AA5AC6A82650A1FA75FBABC5",
        operator_address = "junovaloper1083svrca4t350mphfv9x45wq9asrs60cpqzg0y",
        consensus_address = "junovalcons1dmyqfkahywqdp2j6c65zv59plf6lh279unewtr" # `junod q  tendermint-validator-set | grep -B 5 -A 5 ux/IM9uD`  (where ux/ was found on ping.pub)
    )

    new_account = Account(
        pubkey = args.account_pubkey,
        address = args.account_address
    )

    old_account = Account(
        pubkey = "Ah8/EMTRW6D+Gk3xZghbcoRkKeRA43S8Qo9J+lzf2HnK", # junod q  account juno1083svrca4t350mphfv9x45wq9asrs60c7a585a        
        address = "juno1083svrca4t350mphfv9x45wq9asrs60c7a585a"  # validators account
    )   


    # print(args)
    print("📝 Opening {}... (it may take a while)".format(args.input_genesis))

    # input genesis from within the docker container
    with open(args.input_genesis, 'r') as f:
        genesis = json.load(f)
    
    # print the size of the genesis
    print("📝 Genesis size: {} bytes".format(len(json.dumps(genesis))))    
    # Replace chain-id
    if args.quiet:
        print("🔗 Replace chain-id {} with {}".format(genesis['chain_id'], args.chain_id))
    genesis['chain_id'] = args.chain_id
    

    # Update gov module
    if args.quiet:
        print("🗳️ Update gov module")
        print("\tModify governance_voting_period from {} to {}".format(
            genesis['app_state']['gov']['voting_params']['voting_period'],
            config["governance_voting_period"]))
    genesis['app_state']['gov']['voting_params']['voting_period'] = config["governance_voting_period"]

    # Prune IBC
    if args.prune_ibc:
        if args.quiet:
            print("🕸 Pruning IBC module")

        genesis['app_state']["ibc"]["channel_genesis"]["ack_sequences"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["acknowledgements"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["channels"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["commitments"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["receipts"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["recv_sequences"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["send_sequences"] = []

        genesis['app_state']["ibc"]["client_genesis"]["clients"] = []
        genesis['app_state']["ibc"]["client_genesis"]["clients_consensus"] = []
        genesis['app_state']["ibc"]["client_genesis"]["clients_metadata"] = []

    # Impersonate validator
    if args.quiet:
        print("🚀 Replace validator")

        # print("\t{:50} -> {}".format(old_validator.moniker, new_validator.moniker))
        print("\t{:20} {}".format("Pubkey", new_validator.pubkey))
        print("\t{:20} {}".format("Consensus address", new_validator.consensus_address))
        print("\t{:20} {}".format("Operator address", new_validator.operator_address))
        print("\t{:20} {}".format("Hex address", new_validator.hex_address))

    replace_validator(genesis, old_validator, new_validator)

    # Impersonate account
    if args.quiet:
        print("🧪 Replace account")
        print("\t{:20} {}".format("Pubkey", new_account.pubkey))
        print("\t{:20} {}".format("Address", new_account.address))
    
    replace_account(genesis, old_account, new_account)
        
    # Update staking module
    if args.quiet:
        print("🥩 Update staking module")

    # Replace validator pub key in genesis['app_state']['staking']['validators']
    for validator in genesis['app_state']['staking']['validators']:
        if validator['description']['moniker'] == old_validator.moniker:
            
            # Update delegator shares            
            validator['delegator_shares'] = str(int(float(validator['delegator_shares']) + 1000000000000000)) + ".000000000000000000"
            if args.quiet:
                print("\tUpdate delegator shares to {}".format(validator['delegator_shares']))

            # Update tokens
            validator['tokens'] = str(int(validator['tokens']) + 1000000000000000)
            if args.quiet:
                print("\tUpdate tokens to {}".format(validator['tokens']))
            break
    
    # Update self delegation on operator address
    for delegation in genesis['app_state']['staking']['delegations']:
        if delegation['delegator_address'] == new_account.address:
            # delegation['validator_address'] = new_validator.operator_address
            delegation['shares'] = str(int(float(delegation['shares'])) + 1000000000000000) + ".000000000000000000"
            if args.quiet:
                print("\tUpdate {} delegation shares to {} to {}".format(new_account.address, delegation['validator_address'], delegation['shares']))
            break

    # Update genesis['app_state']['distribution']['delegator_starting_infos'] on operator address
    for delegator_starting_info in genesis['app_state']['distribution']['delegator_starting_infos']:
        if delegator_starting_info['delegator_address'] == new_account.address:
            delegator_starting_info['starting_info']['stake'] = str(int(float(delegator_starting_info['starting_info']['stake']) + 1000000000000000))+".000000000000000000"
            if args.quiet:
                print("\tUpdate {} stake to {}".format(delegator_starting_info['delegator_address'], delegator_starting_info['starting_info']['stake']))
            break

    if args.quiet:
        print("🔋 Update validator power")

    # Update power in genesis["validators"]
    for validator in genesis["validators"]:
        if validator['name'] == old_validator.moniker:
            validator['power'] = str(int(validator['power']) + 1000000000)
            if args.quiet:
                print("\tUpdate {} validator power to {}".format(validator['address'], validator['power']))
            break 
    
    for validator_power in genesis['app_state']['staking']['last_validator_powers']:
        if validator_power['address'] == old_validator.operator_address:
            validator_power['power'] = str(int(validator_power['power']) + 1000000000)
            if args.quiet:
                print("\tUpdate {} last_validator_power to {}".format(old_validator.operator_address, validator_power['power']))
            break
    
    # Update total power
    genesis['app_state']['staking']['last_total_power'] = str(int(genesis['app_state']['staking']['last_total_power']) + 1000000000)
    if args.quiet:
        print("\tUpdate last_total_power to {}".format(genesis['app_state']['staking']['last_total_power']))

    # Update bank module
    if args.quiet:
        print("💵 Update bank module")

    for balance in genesis['app_state']['bank']['balances']:
        if balance['address'] == new_account.address:
            for coin in balance['coins']:
                if coin['denom'] == "ujuno":
                    coin["amount"] = str(int(coin["amount"]) + 1000000000000000) # used to be only 1, but we removed a module so added another 1bn here
                    if args.quiet:
                        print("\tUpdate {} ujuno balance to {}".format(new_account.address, coin["amount"]))
                    break
            break
        
    # Add 1 BN ujuno to bonded_tokens_pool module address
    for balance in genesis['app_state']['bank']['balances']:
        if balance['address'] == BONDED_TOKENS_POOL_MODULE_ADDRESS:
            # Find ujuno
            for coin in balance['coins']:
                if coin['denom'] == "ujuno":
                    coin["amount"] = str(int(coin["amount"]) + 1000000000000000)
                    if not args.quiet:
                        print("\tUpdate {} (bonded_tokens_pool_module) ujuno balance to {}".format(BONDED_TOKENS_POOL_MODULE_ADDRESS, coin["amount"]))
                    break
            break

    for supply in genesis['app_state']['bank']['supply']:
        if supply["denom"] == "ujuno":
            if args.quiet:
                print("\tUpdate total ujuno supply from {} to {}".format(supply["amount"], str(int(supply["amount"]) + 2000000000000000)))
            supply["amount"] = str(int(supply["amount"]) + 2000000000000000)
            break
                  
    os.makedirs(os.path.dirname(args.output_genesis)  , exist_ok=True)  
    
    print("📝 Writing {}... (it may take a while)".format(args.output_genesis))
    with open(args.output_genesis, 'w') as f:
        if args.pretty_output:
            f.write(json.dumps(genesis, indent=2))
        else:
            f.write(json.dumps(genesis))

if __name__ == '__main__':
    main()