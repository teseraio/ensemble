// Code generated for package clickhouse by go-bindata DO NOT EDIT. (@generated)
// sources:
// resources/cluster.template
package clickhouse

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _resourcesClusterTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xc4\x56\xcb\x92\xeb\x26\x10\xdd\xcf\x57\x10\xef\xed\x96\xec\x79\xd8\xb7\x18\x6e\x65\x91\x5d\xf2\x0d\x14\x82\xb6\x4d\x19\x81\x02\xc8\x63\xcf\xd7\xa7\x10\xb2\x2c\xf9\x91\xca\x26\x75\x6b\x16\x53\x7d\xfa\x74\x37\x0d\xdd\x47\xa6\x3f\x4f\xb5\x21\x47\xf4\x41\x3b\xfb\x39\x2b\x17\xc5\xec\x27\x7b\xa1\xd2\xd5\x8d\xb0\x67\xf6\x42\x08\x21\xd4\xb8\xdd\x0e\x7d\x36\x32\x80\x47\x34\x4c\x61\xd5\xee\x28\x64\xe3\xea\x94\xce\x06\x67\x90\x45\xdf\x22\x85\x8b\x35\x0a\x76\x3b\xe2\xb1\x76\x47\xfc\x9c\xe5\xff\x33\x18\xb9\xd1\x7b\xe7\x9f\x73\x28\x5c\x4e\x93\xcd\xbf\x5b\xf4\x67\x6e\xdc\x6e\x94\x42\x89\x28\x2a\x11\x90\x85\x73\x88\x58\x53\x18\x80\x2b\x27\x8a\xca\x20\x1b\xa2\x29\x64\xa0\xaf\x31\xca\x9a\x91\xdf\xe6\x73\xf2\xa7\x0e\x11\x2d\xf9\xd2\x46\x49\xe1\x15\x11\x4a\x79\x0c\x81\x44\x47\x84\x31\xee\x8b\x08\x29\xb1\x89\xda\xee\x88\x74\xd6\xa2\x8c\xda\xd9\x40\xb6\xde\xd5\xc4\xc5\x3d\xfa\x04\x47\xa1\x2d\xfa\x40\x84\x55\x64\xef\x42\x24\x16\xe3\x97\xf3\x87\x05\x99\xcf\x2f\xd7\xdd\xd5\xe1\xc9\xcb\x7e\xfc\xa0\x30\xb6\xef\x19\xc5\xa2\xfb\xfb\x17\x5a\xf4\x67\x56\x0e\xfe\x64\xf5\x4d\xed\x63\x6c\x78\xe3\x7c\x64\xeb\x72\xb9\xa2\x70\xb5\xb3\x3f\xca\xde\xdc\x14\x45\x41\x61\x30\xb3\x57\xdb\x88\x3e\xa0\x3f\xa2\xe7\x5d\x64\x57\xf9\x5d\x54\x28\xd6\x65\x35\x2f\x17\xbf\x53\x78\xcc\x79\x12\x7f\x29\xb5\x79\x10\x97\xeb\xe6\xc0\x5a\x9c\xf8\xe8\x82\xd9\x6b\xb1\x79\xa7\x70\x8b\x66\xee\x01\xb1\xe1\xc2\xe8\x23\xf2\xa8\x6b\x74\x6d\x64\x2b\x0a\x0f\xd0\x49\x6a\xd9\x7a\x8f\x36\xf2\x34\x06\x1a\x03\x2b\x53\xfb\x4f\x7c\x39\xb0\xb5\x69\x65\xd2\x38\xa0\xe2\x52\xc8\x3d\xf2\xa0\xbf\x91\xad\xdf\xd6\x9b\xcd\xea\xf5\x6d\xb3\xa4\xf0\x8c\x73\x29\xed\x0f\x63\xf4\x6d\xf5\xbe\xfe\x28\x36\xe5\xb2\x2b\x3d\xf5\xf5\x17\xd1\x88\xb8\x67\x70\x14\x1e\x8c\xae\x40\x1a\x2d\x0f\x7b\xd7\x06\x04\x0a\x9d\xab\x7f\xc5\xba\xe1\x4f\x99\xb1\x6e\x80\xc2\x40\xe9\x9b\x09\xe8\xf9\x56\x1b\x0c\xcf\x03\xaf\x1c\xa0\x70\x1b\xf0\x72\xcd\x13\xd2\x9d\x6d\xf5\x8e\x75\xc6\xe2\x54\x9b\x4c\x1f\xf0\x4c\x55\xb8\x15\xad\x89\xbc\xf1\x2e\xe5\x61\xbd\x4d\xe1\xd6\x31\xa5\x0f\x8b\x7d\xc7\x9f\xae\x3c\x4d\xcf\xfc\xed\x2c\xb2\x3f\x5a\xef\x1a\x84\xbf\x5c\x90\xee\x8b\xc2\x80\xf7\x8f\x60\x9c\x3c\x70\x3c\xa1\x6c\xb3\x1e\x6c\x85\x09\x48\xe1\x0e\xef\x3b\x4c\xda\x14\x91\xe7\x49\x0d\x13\x11\xec\x04\x94\x4b\xd3\x86\x38\xd6\xce\xce\x1b\xf6\xc2\xab\x29\xd6\xa7\x6b\x8c\x96\xe2\xde\x93\xd7\xf5\x7e\xbf\xae\xeb\x74\xc7\x1e\x2d\xee\x75\x69\x27\x0c\x78\x5a\xee\xbf\x1c\xa4\x7a\x2b\xde\x3f\xaa\x57\x39\x5f\xfe\xe2\x83\x14\xab\x8f\xd5\x56\xa8\x62\xbe\xfa\x9f\x0e\x42\xe1\xe6\xbd\xd2\x57\xed\xc1\xf3\xa6\xf0\xe9\x38\x64\xf8\xdb\xb9\x24\x38\x93\x2f\xa8\x75\x0a\x89\xb6\x0a\x4f\x9f\xb3\x72\x76\x53\xaf\xeb\x61\x88\x7a\xd4\x53\xee\x65\x59\xae\xcb\xdb\x5e\x28\xa4\xd4\x97\x03\x8d\x4a\x5f\x54\x46\x7a\x37\x19\xd4\xbe\x83\x9b\x8e\x28\xdc\x4d\x6e\x3f\xb5\x45\x39\xb9\x8f\x24\x4e\x39\x65\xbf\x99\x3a\x44\xaf\xab\x36\xa2\xe2\x4a\x8d\x7f\x17\x64\x29\x19\x6b\x8f\x08\x87\x24\xa3\x2d\x82\x52\x66\xa2\x59\x70\x97\x26\xe3\x5b\xe7\x6b\x11\x79\x90\x7b\xac\xc5\x73\x71\x9a\xd0\x92\x40\x3d\x88\x7b\x19\x5e\x91\xfd\x13\x00\x00\xff\xff\xeb\xcf\xc7\x54\x05\x09\x00\x00")

func resourcesClusterTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesClusterTemplate,
		"resources/cluster.template",
	)
}

func resourcesClusterTemplate() (*asset, error) {
	bytes, err := resourcesClusterTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/cluster.template", size: 2309, mode: os.FileMode(420), modTime: time.Unix(1625069331, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"resources/cluster.template": resourcesClusterTemplate,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"resources": &bintree{nil, map[string]*bintree{
		"cluster.template": &bintree{resourcesClusterTemplate, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
