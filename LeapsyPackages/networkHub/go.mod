module networkHub

go 1.14

replace leapsy.com/packages/logings => ../logings

replace leapsy.com/packages/configurations => ../configurations

replace leapsy.com/packages/jwts => ../jwts

replace leapsy.com/packages/network => ../network

replace leapsy.com/packages/paths => ../paths

replace leapsy.com/packages/tools => ../tools

replace leapsy.com/packages/model => ../model

replace leapsy.com/databases => ../databases

replace leapsy.com/packages/serverDataStruct => ../serverDataStruct

replace leapsy.com/packages/serverResponseStruct => ../serverResponseStruct

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/gobwas/ws v1.1.0
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	leapsy.com/databases v0.0.0-00010101000000-000000000000
	leapsy.com/packages/configurations v0.0.0-00010101000000-000000000000
	leapsy.com/packages/jwts v0.0.0-00010101000000-000000000000
	leapsy.com/packages/logings v0.0.0-00010101000000-000000000000
	leapsy.com/packages/model v0.0.0-00010101000000-000000000000
	leapsy.com/packages/network v0.0.0-00010101000000-000000000000
	leapsy.com/packages/paths v0.0.0-00010101000000-000000000000
	leapsy.com/packages/serverDataStruct v0.0.0-00010101000000-000000000000
	leapsy.com/packages/serverResponseStruct v0.0.0-00010101000000-000000000000
	leapsy.com/packages/tools v0.0.0-00010101000000-000000000000
)
