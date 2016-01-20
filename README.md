
This application will listen on tcp 8000 for proxy parser and also 8081 for sample http server.

`go run proxy.go`

Please setting up a HAproxy on machine B porting to machine A, running proxy.go, tcp 8000.
When you query to machine B and will get your real IP address from machine A response by sample http server.

