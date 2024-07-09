module termbox-go

go 1.22

require (
	github.com/mattn/go-runewidth v0.0.15
	golang.org/x/sys v0.22.0
)

require github.com/rivo/uniseg v0.4.7 // indirect

retract v1.1.0 // panics on BSD
