## BGP commander

The point of this is to integrate with [ExaBGP](https://github.com/Exa-Networks/exabgp) and [etcd](https://github.com/coreos/etcd). Basically this script will monitor keys in etcd, and announce/withdraw things via the ExaBGP api. Right now it's a proof of concept, but I plan on finishing it.

### Example usage

        Usage of ./bgpcommander:
          -c=http://127.0.0.1:4001: Comma separated list of etcd cluster members
          -n="killallhumans.local": Override node's name
          -p="/tmp": Path to store healthcheck scripts


When running in conjuction with ExaBGP, a process configuration like this is inserted:

        process bgpcommander {
          encoder json;
          neighbor-changes;
          run /path/to/bgpcommander;
        }

Any of the flags available can be passed to override various defaults. You probably want to point to a real etcd cluster for example. Once you have this configured and the process is running, there are a few keys in etcd that you'll want to modify.

**Creating a new route**

All routes exist in the `/bgp/routes` directory and can be created with a couple curl commands.

        # This key is what is used to announce or withdraw routes with ExaBGP
        $ curl localhost:4001/v2/keys/bgp/routes/CUSTOM_ROUTE/config -XPUT \
             --data-urlencode value="route 172.16.1.5/32 next-hop self med 100"

        # This key is what is downloaded and ran as a healthcheck for the route.
        # As long as the script returns 0, the healthcheck is considered passing. If anything
        # else is returned, a withdraw command is used to remove the route.
        $ cat < EOF > mytest.sh
        #!/bin/sh
        dig @172.16.1.5 google.com
        $ curl localhost:4001/v2/keys/bgp/routes/CUSTOM_ROUTE/healthcheck -XPUT \
            --data-urlencode value@mytest.sh

**Assigning a route to a node**

Unless you specify that a node should advertise a route, it will not. To do that you simply add to the `subscribedRoutes` key in etcd for that node.

        # This key should be a space separate list of route names that you would like to
        # have this node monitor and announce
        $ curl localhost:4001/v2/keys/bgp/node/killallhumans.local/subscribedRoutes -XPUT -d value="CUSTOM_ROUTE"

**Checking for neighbor state**

In addition to controlling routes, it's easy to grab state information about neighbors. This data is available like so:

        $ curl localhost:4001/v2/keys/bgp/node/killallhumans.local/neighbors?recursive=true
        # TODO: add sample output here, basically you have state and ip address information per neighbor
