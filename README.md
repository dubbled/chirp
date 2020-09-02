# Chirp

Is a simple module for concurrent safe fan out delivery to io.Writers


## Usage

```go
nest := NewNest(nil) // nil supplies Nest with default settings
topic := "testing"
go func() {
    tick := time.Tick(time.Second)
    for {
        <-tick
        err := nest.MsgSubscribers(topic, []byte("tick"))
    }
}()

l, err := net.Listen("tcp", ":8080")
for {
    conn, err := l.Accept()
    client := &Client{Writer: conn}
    nest.InsertClient(topic, client)
    err := client.Write([]byte("connected to hub!"))
}
```

## Contributing
Pull requests are welcome.