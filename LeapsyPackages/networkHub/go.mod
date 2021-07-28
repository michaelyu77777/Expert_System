module networkHub

go 1.14

replace leapsy.com/packages/logings => ../logings

replace leapsy.com/packages/configurations => ../configurations

replace leapsy.com/packages/jwts => ../jwts

replace leapsy.com/packages/network => ../network

replace leapsy.com/packages/paths => ../paths

replace leapsy.com/packages/tools => ../tools

replace leapsy.com/databases => ../databases

replace leapsy.com/packages/model => ../model

replace leapsy.com/databases => ../databases

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/gobwas/ws v1.1.0
	github.com/juliangruber/go-intersect v1.0.0
	// github.com/juliangruber/go-intersect v1.0.0
	github.com/keepeye/logrus-filename v0.0.0-20190711075016-ce01a4391dd1 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	leapsy.com/databases v0.0.0-00010101000000-000000000000
	leapsy.com/packages/configurations v0.0.0-00010101000000-000000000000
	leapsy.com/packages/jwts v0.0.0-00010101000000-000000000000
	leapsy.com/packages/logings v0.0.0-00010101000000-000000000000
	leapsy.com/packages/model v0.0.0-00010101000000-000000000000
	leapsy.com/packages/network v0.0.0-00010101000000-000000000000
	leapsy.com/packages/paths v0.0.0-00010101000000-000000000000
	leapsy.com/packages/tools v0.0.0-00010101000000-000000000000
)
