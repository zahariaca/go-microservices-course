module github.com/zahariaca/broker

go 1.20

replace github.com/zahariaca/toolbox => ../toolbox

require github.com/zahariaca/toolbox v0.0.0-00010101000000-000000000000

require (
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-chi/cors v1.2.1
)

require github.com/go-chi/chi v1.5.4

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/rabbitmq/amqp091-go v1.8.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
