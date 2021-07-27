package global

const ProjectName = "GoGameServer"

var ServerMap *ServerMapAddress

func GlobalInit() {
	ServerMap = NewServerMapAddress()
}
