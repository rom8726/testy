module testyexample

go 1.24

require github.com/rom8726/testy v1.0.0

require (
	github.com/kinbiko/jsonassert v1.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/rom8726/pgfixtures v1.1.1 // indirect
)

require (
	github.com/julienschmidt/httprouter v1.3.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rom8726/testy v1.0.0 => ../
