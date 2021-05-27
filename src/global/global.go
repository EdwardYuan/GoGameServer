package global

var ServerMap *ServerMapAddress

func GlobalInit() {
	ServerMap = NewServerMapAddress()
}
