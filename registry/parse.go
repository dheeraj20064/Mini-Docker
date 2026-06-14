package registry

import "strings"

func ParseImage(image string) (string, string){
	parts:=strings.Split(image,":")

	repo:=parts[0]
	tag:="latest"

	if len(parts)>1{
		tag=parts[1]
	}
	if !strings.Contains(repo,"/"){
	repo="library/"+repo
	}
	return repo, tag
}
