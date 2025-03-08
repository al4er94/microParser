package repo

var RepoVkId map[int]struct{}
var RepoUrl map[int]string

func InitRepo() {
	RepoVkId = make(map[int]struct{})
	RepoUrl = make(map[int]string)
}
