# net/smc: reading pnet table as unprivileged user

The flags of the `SMC_PNET_GET` netlink command allow only privileged users to
read entries from SMC's pnet table. But, the content of the pnet table can be
useful for all users, e.g., for debugging SMC connection problems.

The patch in [net-next.patch](net-next.patch) removes the `GENL_ADMIN_PERM`
flag and, thus, allows unprivileged users to read the pnet table.
