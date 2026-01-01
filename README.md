# TinyRedis: Lightweight In-Memory Storage Engine

TinyRedis is a minimal, drop-in Redis replacement built in Go. It speaks the RESP protocol, so it works seamlessly with any Redis client, including redis-cli. 

Start storage engine
```bash
go run main.go
```
Connect with redis-cli
```bash
redis-cli -h 127.0.0.1 -p 8379
```