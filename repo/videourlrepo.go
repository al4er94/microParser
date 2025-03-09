package repo

var RepoVkId map[int]struct{}
var RepoUrl map[int]VideoEntity

type VideoEntity struct {
	Url    string
	ImgUrl string
}

func InitRepo() {
	RepoVkId = make(map[int]struct{})
	RepoUrl = make(map[int]VideoEntity)
}
