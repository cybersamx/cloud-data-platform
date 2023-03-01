package main

func main() {
	err := rootCommand().Execute()
	checkErr(err)
}
