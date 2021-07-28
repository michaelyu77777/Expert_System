module databases

go 1.16

replace leapsy.com/packages/configurations => ../configurations

replace leapsy.com/packages/logings => ../logings

replace leapsy.com/packages/network => ../network

replace leapsy.com/packages/model => ../model

replace leapsy.com/packages/paths => ../paths

replace leapsy.com/packages/networkHub => ../networkHub

require (
	github.com/sirupsen/logrus v1.8.1
	go.mongodb.org/mongo-driver v1.5.4
	leapsy.com/packages/configurations v0.0.0-00010101000000-000000000000
	leapsy.com/packages/logings v0.0.0-00010101000000-000000000000
	// leapsy.com/packages/logings v0.0.0-00010101000000-000000000000
	leapsy.com/packages/model v0.0.0-00010101000000-000000000000
	leapsy.com/packages/network v0.0.0-00010101000000-000000000000
	leapsy.com/packages/paths v0.0.0-00010101000000-000000000000

)
