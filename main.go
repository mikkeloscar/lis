package main

func main() {
	lis, _ := NewLis("/var/lib/lis", 5*1000)
	lis.run()
}
