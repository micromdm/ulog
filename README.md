`ulog` is a utility for shipping macOS unified logs to a remote server. It also includes the server which will stream all incoming logs to stdout.

Why? Often I go through the DEP setup process on a VM and I want to be able to debug the process. `ulog` ensures that all the VM logs are immediately available on the server whenever I'm testing micromdm or DEP bootstrap. 

## Using the client

`make pkg` will create a pkg which can be added to the VM. Modify the `SERVER_URL` in the Makefile first. The process will automatically load on install with the following command:

`ulog client -config /etc/micromdm/ulog/server.json`

## Using the server

```
ulog server -addr=:8080 > mdmlog.log
```
