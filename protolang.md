# sketching out stuff

Dependencies/package managements/imports

// load types
Load(Path("/myrepo/mydef.proto"))
Load(URL("https", "types.accretional.net", "myrepo/mydef))
Load(Index("https", "index.accretional.net", "type=mydef"))

// becomes mybin
Load(Path("/myrepo/mybin.binarypb"))
Load(URL("https", "data.accretional.net", "myrepo/mybin"))
Load(Index("https", "index.accretional.net", "file=mybin"))

// load files: types, binaries, paths, source code
Load(Repo{Path{"/myrepo"}})
Load(URL{"https", "repos.accretional.net", "myrepo"})
Load(Index{"https", "index.accretional.net", "repo=myrepo"})



Resolve(URL{"https", "index.accretional.net", "type~FileServer"})
