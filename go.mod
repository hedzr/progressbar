module github.com/hedzr/progressbar/v2

go 1.23.0

//replace gopkg.in/hedzr/errors.v3 => ../../go-cmdr/05.errors
//
//replace github.com/hedzr/log => ../../go-cmdr/10.log
//
//replace github.com/hedzr/logex => ../../go-cmdr/15.logex
//
//replace github.com/hedzr/cmdr => ../../go-cmdr/50.cmdr

//replace github.com/hedzr/tuilive => ../tuilive

require github.com/hedzr/is v0.8.27

require (
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
)
