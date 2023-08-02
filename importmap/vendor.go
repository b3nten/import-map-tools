package importmap

import (
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var VendorPath = "_vendor"

func (m *ImportMap) Vendor() error {
	// create vendor directory
	err := os.MkdirAll(VendorPath, 0755)
	if err != nil {
		return err
	}
	vendorMap := ImportMap{}
	for module, specifier := range m.Imports {
		if isExternalPath(specifier) {
			localPath := vendorModule(specifier)
			vendorMap.Add(module, localPath)
		}
	}
	// write vendor importmap
	err = os.WriteFile(path.Join(VendorPath, "import_map.json"), []byte(vendorMap.String()), 0644)
	return nil
}

func vendorModule(modpath string) string {
	if !isExternalPath(modpath) {
		panic("xtern path?!?")
	}
	code, err := fetchModule(modpath)
	if err != nil {
		panic(err)
	}

	imports := getImports(code)

	for _, v := range imports {
		if isLocalPath(v) {
			panic("Relative import paths are not supported yet.")
		} else if isAbsolutePath(v) {
			pathurl, _ := url.Parse(modpath)
			v = "https://" + path.Join(pathurl.Host, v)
		}
		vendorModule(v)
	}

	replacements := map[string]string{}

	for _, v := range imports {
		vpath, _ := url.Parse(modpath)
		relpath, _ := filepath.Rel(path.Dir(urlToVendorPath(vpath.Path)), urlToVendorPath(v))
		replacements[v] = strings.ReplaceAll(relpath, "\\", "/")
	}

	for k, v := range replacements {
		code = strings.ReplaceAll(code, k, v)
	}

	os.MkdirAll(path.Dir(urlToVendorPath(modpath)), 0755)
	file, _ := os.Create(urlToVendorPath(modpath))
	defer file.Close()
	file.WriteString(code)

	return urlToVendorPath(modpath)
}

func fetchModule(p string) (string, error) {
	response, err := http.Get(p)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func isExternalPath(p string) bool {
	return strings.HasPrefix(p, "https://") || strings.HasPrefix(p, "http://")
}

func isLocalPath(p string) bool {
	return strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../")
}

func isAbsolutePath(p string) bool {
	return path.IsAbs(p)
}

func urlToVendorPath(p string) string {
	url, err := url.Parse(p)
	if err != nil {
		return ""
	}
	if len(path.Ext(p)) == 0 {
		// return path + index.js
		return path.Join(VendorPath, url.Host, url.Path, "index.js")
	} else {
		// return the path
		return path.Join(VendorPath, url.Host, url.Path)
	}
}

func getImports(s string) []string {
	var imports []string
	pluginGetImports := api.Plugin{
		Name: "getImports",
		Setup: func(build api.PluginBuild) {
			build.OnResolve(api.OnResolveOptions{Filter: `.*`},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					imports = append(imports, args.Path)
					return api.OnResolveResult{
						Path:     args.Path,
						External: true,
					}, nil
				},
			)
		},
	}
	result := api.Build(api.BuildOptions{
		Plugins: []api.Plugin{pluginGetImports},
		Stdin: &api.StdinOptions{
			Contents: s,
		},
		Bundle: true,
		Write:  false,
		Format: api.FormatESModule,
	})
	if len(result.Errors) > 0 {
		fmt.Println(result.Errors)
	}
	return imports
}
