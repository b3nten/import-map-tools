package main

import (
	"github.com/B3nten/imt/importmap"
)

func main() {
	importMap := importmap.ImportMap{}
	importMap.Add("react", "https://cdn.skypack.dev/react")
	importMap.Add("react-dom", "https://cdn.skypack.dev/react-dom")
	importMap.Vendor()
}
