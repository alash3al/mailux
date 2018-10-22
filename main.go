package main

func main() {
	go initHTTP()
	go initSMTP()

	select {}
}
