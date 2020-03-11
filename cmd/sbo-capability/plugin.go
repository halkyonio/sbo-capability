package main

import (
	plugins "halkyon.io/operator-framework/plugins/capability"
	"halkyon.io/sbo-capability/pkg/plugin/sbo"
)

func main() {
	plugins.StartPluginServerFor(sbo.NewPluginResource())
}
