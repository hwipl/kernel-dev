# wireguard: problem sending packets with AF\_PACKET socket

Sending packets with an `AF_PACKET` socket (e.g, with libpcap) does not work
with a wireguard interface in contrast to a regular ethernet interface.

It seems like `skb->protocol` is not set to an accepted protocol (`ETH_P_IP` or
`ETH_P_IPV6`) and the following checks

* `skb->protocol == real_protocol`,
* `skb->protocol == htons(ETH_P_IP)` and
* `skb->protocol == htons(ETH_P_IPV6)`

in the functions

* `wg_check_packet_protocol()`,
* `wg_xmit()` and
* `wg_allowedips_lookup_dst()`

fail because `skb->protocol` is `0`.

The patch in [net-next.patch](net-next.patch) is an initial attempt to fix the
issue for testing and was posted when reporting the issue.
