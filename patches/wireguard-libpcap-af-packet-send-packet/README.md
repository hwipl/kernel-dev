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
issue for testing and was posted when reporting the issue. The patch is for
`net-next` version `5.8.0-rc2`.

The files [rawsock-pcap.go](rawsock-pcap.go) and [rawsock.go](rawsock.go)
contain test code to reproduce the issue. The pcap version reproduces the
situation that triggered this issue. The other version opens `AF_PACKET`
sockets without libpcap and sets the protocol of the sockets to `0`.

For testing, it is useful to enable debug messages for wireguard:

```console
# echo module wireguard +p > /sys/kernel/debug/dynamic_debug/control
```

After the issue was reported, Jason A. Donenfeld investigated the problem and
discovered that it is neither a libpcap-only nor a libpcap-specific issue and
applies to `AF_PACKET` sockets in general. As a side-note, he also found out
that, besides wireguard, it applies to other tunnel devices as well. The cause
of the issue is that wireguard does not implement `dev->header_ops`.

According to his analysis the test program triggers calls into `af_packet`'s
`packet_sendmsg->packet_snd` function and, thus, into
`packet_parse_headers(skb, sock)`:


```c
static void packet_parse_headers(struct sk_buff *skb, struct socket *sock)
{
    if ((!skb->protocol || skb->protocol == htons(ETH_P_ALL)) &&
        sock->type == SOCK_RAW) {
        skb_reset_mac_header(skb);
        skb->protocol = dev_parse_header_protocol(skb);
    }

    skb_probe_transport_header(skb);
}
```

Then, the function `dev_parse_header_protocol(skb)` resets `skb->protocol`
because there are no `dev->header_ops` and, thus, there is also no
`parse_protocol()` function:

```c
static inline __be16 dev_parse_header_protocol(const struct sk_buff *skb)
{
    const struct net_device *dev = skb->dev;

    if (!dev->header_ops || !dev->header_ops->parse_protocol)
        return 0;
    return dev->header_ops->parse_protocol(skb);
}
```

As a result, the `skb->protocol` is set to `0` when the packets reach
`wg_xmit()`.

He posted a patch series on the netdev mailing list that fixes this issue for
ipip, wireguard, and tun devices by adding `header_ops` for these devices. They
are called `ip_tunnel_header_ops` and use wireguard's
`wg_examine_packet_protocol()` function renamed to `ip_tunnel_parse_protocol()`
as `parse_protocol`:

```c
/* Returns either the correct skb->protocol value, or 0 if invalid. */
__be16 ip_tunnel_parse_protocol(const struct sk_buff *skb)
{
 if (skb_network_header(skb) >= skb->head &&
     (skb_network_header(skb) + sizeof(struct iphdr)) <= skb_tail_pointer(skb) &&
     ip_hdr(skb)->version == 4)
  return htons(ETH_P_IP);
 if (skb_network_header(skb) >= skb->head &&
     (skb_network_header(skb) + sizeof(struct ipv6hdr)) <= skb_tail_pointer(skb) &&
     ipv6_hdr(skb)->version == 6)
  return htons(ETH_P_IPV6);
 return 0;
}
EXPORT_SYMBOL(ip_tunnel_parse_protocol);

const struct header_ops ip_tunnel_header_ops = { .parse_protocol = ip_tunnel_parse_protocol };
EXPORT_SYMBOL(ip_tunnel_header_ops);
```

Willem de Bruijn also commented on the skb protocol in `AF_PACKET` sockets:

> Not setting skb protocol when sending over packet sockets causes many
> headaches. Besides packet\_parse\_headers, virtio\_net\_hdr\_to\_skb also
> tries to infer it.
>
> Packet sockets give various options to configure it explicitly: by
> choosing that protocol in socket(), bind() or, preferably, by passing
> it as argument to sendmsg. The socket/bind argument also configures
> the filter to receive packets, so for send-only sockets it is
> especially useful to choose ETH\_P\_NONE (0) there. This is not an
> "incorrect" option.
>
> Libpcap does have a pcap\_set\_protocol function, but it is fairly
> recent, so few processes will likely be using it. And again it is
> still not ideal if a socket is opened only for transmit.
