## v2.1.0 Patch

Originally we had planned on a straight-up `v2.2.0` release, but testing that in the timeslot we had available wasn't possible. As a result, we are distributing this set of instructions for patching your node. Note:

- **Until you have patched your node, it is vulnerable**
- **Until 2/3 of validators have patched their nodes, the network is vulnerable**

So you can understand our desire to get the entire validator set _patched as soon as possible_.

For this update, you should update using the patched binary supplied on the `v2.1.0` release. It contains a statically compiled wasmvm that needs to be built on alpine.

This patched `junod` binary can be run on a linux box. It was built on alpine and has been tested on various Ubuntu LTSs and Debian bullseye. If you are concerned about your OS, then try patching a sentry first before patching your validator.

**Important:** you can compile this release yourself, _but only on alpine, and you will need to verify using the steps below, or your box will still be vulnerable_. As a result, we recommend you use the binary we've built.

### Step 1: stop your node

If you are not using cosmovisor, you should stop your node before switching the binary.

If you are using cosmovisor, you can download and verify the binary, then follow the steps underneath.

### Step 2: download and verify

```sh
# find out where junod is - will likely be /home/<your-user>/go/bin/junod
which junod

# put new binary there i.e. in path/to/juno
wget https://github.com/CosmosContracts/juno/releases/download/v2.1.0/junod -O /home/<your-user>/go/bin/junod

# if you run this, you should see build_tags: netgo muslc,
# if there is a permissions problem use chmod/chown to make sure it is executable
junod version --long

# confirm it is using the static lib - should return "not a dynamic executable"
ldd /home/<your-user>/go/bin/junod

# if you really want to be sure
# this should return:
# ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, 
# Go BuildID=4Ec3fj_EKdvh_u8K3RGJ/JIKOgYFXTJ9LzGROhs8n/uC9gpeaM9MaYurh9DJiN/YcvB8Jc2ivQM2zUSHMhg, stripped
file /home/<your-user>/go/bin/junod
```

### Step 3: restart (if not using cosmovisor)

If you are using a service file that points to this `junod`, you can restart the service.

### Step 3: install and restart (if using cosmovisor)

If you are using cosmovisor:

- follow the steps above to download and verify the binary
- stop cosmovisor
- copy the binary `cp /home/<your-user>/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/moneta-patch/bin`
- check the binary has the `muslc` flag in the output for `$DAEMON_HOME/cosmovisor/upgrades/moneta-patch/bin/junod version --long`
- restart cosmovisor
- 🤞 