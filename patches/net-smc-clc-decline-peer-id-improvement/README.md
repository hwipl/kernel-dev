# net/smc: peer ID in CLC Decline improvement

According to RFC 7609, all CLC messages contain a peer ID that consists of a
unique instance ID and the MAC address of one of the host's RoCE devices. But
if a SMC-R connection cannot be established, e.g., because no matching pnet
table entry is found, the current implementation uses a zero value in the CLC
decline message although the host's peer ID is set to a proper value.

As preparation for the second patch, the patch in
[net-next-1.patch](net-next-1.patch) changes the peer ID initialization. It
sets the peer ID to a random instance ID and a zero MAC address. If a RoCE
device is in the host, the MAC address part of the peer ID is overwritten with
the respective address. The patch also contains a helper function for checking
if the peer ID is valid. That is the case if the MAC address part contains a
non-zero MAC address.

The patch in [net-next-2.patch](net-next-2.patch) fixes the peer ID in CLC
decline messages if RoCE devices are in the host but no suitable device is
found for a connection. For this case, it modifies the LGR check in
`smc_clc_send_decline()` to allow that a valid peer ID, if there is one, is
copied into the CLC decline message for both SMC-D and SMC-R.
