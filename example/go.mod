module testyexample

go 1.24

require (
	github.com/google/uuid v1.6.0
	github.com/rom8726/testy v1.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.0 // indirect
	github.com/kinbiko/jsonassert v1.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/rom8726/pgfixtures v1.2.0 // indirect
)

require (
	github.com/julienschmidt/httprouter v1.3.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rom8726/testy v1.0.0 => ../
