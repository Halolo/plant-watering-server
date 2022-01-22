# plant-watering-server

Small http server in golang that uses gpiod to trigger plant watering system (relays).

it's in a demo / sample code state.

Used on an [Orange pi pc](http://www.orangepi.org/orangepipc/), but could be used on any device with kernel version >= 4.8.

Obviously, if used as-is, has to be restricted to a controlled private network, local or behind a (properly configured) VPN. With firewall rules properly configured on router / devices to be sure a port forwarding rule won't reach the device from outside.
