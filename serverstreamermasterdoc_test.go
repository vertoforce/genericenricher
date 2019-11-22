package serverstreamer

func Example() {
	server, err := GetServer("http://localhost:9200")
	if err != nil {
		return
	}

	p := make([]byte, 10)
	server.Read(p)
}
