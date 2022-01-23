package fsfc

import "fsfc/server"

func main() {
	//Logger

	go server.MyServer.Start()
}
