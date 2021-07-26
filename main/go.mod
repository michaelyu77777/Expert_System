module main

go 1.14

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/gin-gonic/gin v1.7.2
	github.com/gobwas/ws v1.1.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/juliangruber/go-intersect v1.0.0 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
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
