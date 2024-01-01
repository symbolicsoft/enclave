module github.com/symbolicsoft/enclave/v2

go 1.21

toolchain go1.21.5

require (
	github.com/charmbracelet/bubbles v0.17.1
	github.com/charmbracelet/bubbletea v0.25.0
	github.com/charmbracelet/huh v0.2.3
	github.com/charmbracelet/huh/spinner v0.0.0-20231222231237-4bd4657a36ac
	github.com/charmbracelet/lipgloss v0.9.1
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/crypto v0.17.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/catppuccin/go v0.2.0 // indirect
	github.com/charmbracelet/glamour v0.6.0 // indirect
	github.com/containerd/console v1.0.4-0.20230706203907-8f6c4e4faef5 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/microcosm-cc/bluemonday v1.0.26 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sahilm/fuzzy v0.1.1-0.20230530133925-c48e322e2a8f // indirect
	github.com/yuin/goldmark v1.6.0 // indirect
	github.com/yuin/goldmark-emoji v1.0.2 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
)

replace github.com/charmbracelet/bubbles => github.com/wesen/bubbles v0.10.4-0.20231101034402-90254c4c2839

replace github.com/charmbracelet/bubbletea => github.com/knz/bubbletea v0.0.0-20230422204939-97ee90cf5a2c
