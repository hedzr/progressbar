module github.com/hedzr/progressbar/v2

go 1.24.0

//replace gopkg.in/hedzr/errors.v3 => ../../go-cmdr/05.errors
//
//replace github.com/hedzr/log => ../../go-cmdr/10.log
//
//replace github.com/hedzr/logex => ../../go-cmdr/15.logex
//
//replace github.com/hedzr/cmdr => ../../go-cmdr/50.cmdr

//replace github.com/hedzr/tuilive => ../tuilive

require github.com/hedzr/is v0.8.65

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
)
