package schema

#Semver: =~#"^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$"#

#GithubSource: {
	tag:       #Semver
	gitTag:    string | *tag
	githubURL: *"https://github.com" | string
	owner:     =~#"^[\w\.-]+$"#
	repo:      =~#"^[\w\.-]+$"#
	files: [...string]
	dirs: [...string]
	assets: [...string]
}

#KubernetesSource: {
	version: #Semver
}

#Schema: {
	github: [...#GithubSource]
	kubernetes: [...#KubernetesSource]
}
