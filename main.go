package main

import "git.alastairstuff.tk/nottimisreal/religion-creative-unit-api/server"

func main() {
	// Use the imported package here
	var server server.Main
	go func() {
		for {
			if <-server.Killed {
				server.Killed <- false
				server.Start()
			}
		}
	}()
	server.Start()

}
