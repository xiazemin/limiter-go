package main
import(
"fmt"
"net"
"golang.org/x/net/netutil"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fatalf("Listen: %v", err)
	}
	defer l.Close()
	l = LimitListener(l, max)

	http.Serve(l, http.HandlerFunc())

}