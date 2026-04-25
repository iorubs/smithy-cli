module github.com/iorubs/smithy-cli

go 1.26.1

require (
	github.com/alecthomas/kong v1.15.0
	github.com/iorubs/mcpsmithy v0.0.0-00010101000000-000000000000
)

replace github.com/iorubs/mcpsmithy => ../mcpsmithy

require (
	go.yaml.in/yaml/v4 v4.0.0-rc.4 // indirect
	golang.org/x/net v0.53.0 // indirect
)
