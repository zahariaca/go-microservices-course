module github.com/zahariaca/broker

go 1.20

replace github.com/zahariaca/toolbox => ../toolbox

require github.com/zahariaca/toolbox v0.0.0-00010101000000-000000000000

require (
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-chi/cors v1.2.1
)

require github.com/go-chi/chi v1.5.4
