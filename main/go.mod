module main

go 1.14

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/gobwas/ws v1.1.0
	leapsy.com/packages/configurations v0.0.0-00010101000000-000000000000
	leapsy.com/packages/logings v0.0.0-00010101000000-000000000000
	leapsy.com/packages/network v0.0.0-00010101000000-000000000000
	leapsy.com/packages/networkHub v0.0.0-00010101000000-000000000000
)

replace leapsy.com/packages/configurations => ../LeapsyPackages/configurations

replace leapsy.com/packages/logings => ../LeapsyPackages/logings

replace leapsy.com/packages/network => ../LeapsyPackages/network

replace leapsy.com/packages/networkHub => ../LeapsyPackages/networkHub

replace leapsy.com/packages/jwts => ../LeapsyPackages/jwts

replace leapsy.com/packages/paths => ../LeapsyPackages/paths
