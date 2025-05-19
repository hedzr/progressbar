# multibar1

For solving issue #8 and earlier ones, `multibar1` app demonstrates how to coding with `MultiPB`.

## Try

Running this app with these flags:

```bash
$ go run ./examples/multibar1 --help
Usage of /var/folders/zv/5r7hq8bs6qs3cx2z743t_3_h0000gn/T/go-build624010689/b001/exe/multibar1:
  -algor int
        select a algor (0..2)
  -resume
        continue the uncompleted task
  -stop-at int
        the percent which task should puase it at
  -which int
        choose a stepper (0..4)

$ # For example:
$ go run ./examples/multibar1 -resume -which 2
$ WHICH=2 ALGOR=1 go run ./examples/multibar1 -resume
```
