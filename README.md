# [PoC] Firecracker Go SDK: socket connection leak

When we use the [Firecracker Go SDK][fc-go-sdk] to query Firecracker API on a
unix socket, the SDK does not closes the connection.

If the query comes from a short lived CLI tool, it wouldn't be an issue, but
with [Flintlock][flintlock], we have a long running service, that checks the
state of the MicroVM periodically.

[fc-go-sdk]: https://github.com/firecracker-microvm/firecracker-go-sdk
[flintlock]: https://github.com/weaveworks/flintlock/issues/266

## Usage

1. Open two terminal session (to be sure they are not tangled together)
2. One terminal start a new Firecracker instance
```
firecracker --api-sock /tmp/firecracker.socket
```
3. On the other one, run this go application: `go run .`

Output shows we made 3 queries, and after they we still have 3 connections
open.

```
❯ go run .
INFO[0000] Firecracker API response                      state="Not started"
INFO[0000] Firecracker API response                      state="Not started"
INFO[0000] Firecracker API response                      state="Not started"
INFO[0000] Sleep a bit
INFO[0001] Check connections                             pid=861112
INFO[0001] Open connection                               count=3 pid=861112 target=/tmp/firecracker.socket
```
