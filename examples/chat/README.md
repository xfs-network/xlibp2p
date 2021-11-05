# xlibp2p-chat

This program demonstrates a simple p2p chat application. It can work between two or more peers if

## Build

From the xlibp2p/examples/chat directory run the following:

```shell
go build
```

## Usage

First, you need to start a main node as the seed node of other nodes

```shell
./chat [-addr <your_listen_addr>]

...
p2p listen and serve on "<main_node_ip>:<main_node_port>"
p2p server node id: "<main_node_id>"
...
```
After successful operation, some output node information will be displayed on the screen. Remember them

> You can use the `-addr` option to specify the P2P service listening address.
If it is not set, the local random port will be used by default.


And then you need to create a new terminal program and start another node to connect to the master node

```shell
./chat [-addr <your_listen_addr>] -bootstrap xfsnode://<main_node_ip>:<main_node_port>/?id=<main_node_id>
```

Next, you can enter any character in the terminal of main node, such as "hello"

```shell
> hello
```

If successful, it will echo on other terminal screens

```shell
<(<node_id>): hello
```

Finally, you can try to follow the above steps to start multiple nodes to test connectivity

> You can use the `-h` option to get more command line help

