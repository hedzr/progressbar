module github.com/hedzr/progressbar

go 1.18

//replace gopkg.in/hedzr/errors.v3 => ../../go-cmdr/05.errors
//
//replace github.com/hedzr/log => ../../go-cmdr/10.log
//
//replace github.com/hedzr/logex => ../../go-cmdr/15.logex
//
//replace github.com/hedzr/cmdr => ../../go-cmdr/50.cmdr

//replace github.com/hedzr/tuilive => ../tuilive

require (
	golang.org/x/crypto v0.27.0
	golang.org/x/net v0.29.0
)

require (
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/term v0.24.0 // indirect
)
