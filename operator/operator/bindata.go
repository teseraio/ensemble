// Code generated for package operator by go-bindata DO NOT EDIT. (@generated)
// sources:
// resources/crd-cluster.json
// resources/crd-resource.json
// resources/srv-cluster-role-binding.json
// resources/srv-cluster-role.json
// resources/srv-deployment.json.template
// resources/srv-service-account.json
package operator

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

var _resourcesCrdClusterJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xa4\x55\x4d\x8f\xd4\x30\x0c\x3d\x27\xbf\xa2\xca\x79\x34\x08\x71\x41\x73\x43\x20\x21\x2e\x08\x81\xb4\x97\xd5\x1e\xd2\x8c\x29\x61\xda\x24\xd8\xce\x08\xb4\xea\x7f\x47\xe9\x47\xb6\x3b\x1f\xdd\x76\xb6\xa7\xc6\x7e\x2f\x7e\x7e\xb2\x9c\x47\x29\x94\x0e\xf6\x0e\x90\xac\x77\x6a\x57\xa4\x13\xfc\x65\x70\xe9\x4c\xdb\xc3\x7b\xda\x5a\xff\xe6\xf8\x56\x6d\xa4\x50\x07\xeb\xf6\x09\xf3\x31\x12\xfb\xe6\x3b\x90\x8f\x68\xe0\x13\xfc\xb4\xce\x72\xe2\x27\x50\x03\xac\xf7\x9a\xb5\xda\x15\x8f\x52\x08\xe5\x74\x03\x89\x64\xea\x48\x0c\x48\x5b\x70\x04\x4d\x59\x83\xa7\x74\xb7\x92\x45\x51\x14\x6d\x62\x52\x00\x33\xb2\x2a\xf4\x31\x24\xda\x09\x7a\x93\x92\xc7\x5e\x2e\xa9\x5d\x71\x2f\x85\x10\x89\x31\xa9\xd4\xab\x4d\x11\x02\x3c\x42\x92\xcc\x18\x61\x8c\xb1\x47\x5d\xc1\x49\xd0\xfc\x82\x66\x94\x9c\x02\x3e\x80\xfb\xf0\xed\xcb\xdd\xbb\x1f\x4f\x99\xe2\xca\xa7\xf8\x5f\xe8\x0a\xfb\xf2\x37\x18\x56\x9b\xeb\xc8\x80\x3e\x00\xb2\x05\x9a\xbd\xb1\xc3\x66\x3b\xe6\x50\xeb\xea\xdf\xa2\x23\x73\x4a\x6d\x0e\xd0\x0d\xc0\x32\xc2\x6d\xd2\x5e\x23\x31\x73\x87\x49\x58\xc7\x7a\xa6\x97\x18\xad\xab\xd4\xaa\x0b\xda\xc5\xe8\x76\x85\x0f\x08\x7f\xa2\xc5\x6e\x8a\xef\xfb\xc6\x1e\x16\x91\x17\xd6\x50\x08\xa1\xb6\x46\xaf\x73\x39\xfb\x64\x1d\x43\x05\xb8\xcc\xa8\x97\x0d\x5a\x20\xfa\xb9\x21\xe3\x50\x6e\x26\x8d\xcc\xfb\xf3\x42\x09\x45\xac\x39\x2e\x73\xe3\x7c\xba\xc5\xf0\x9d\x4e\xaf\xc8\x09\x5f\xf6\x4b\xe9\x33\x38\x40\xcd\xfd\xd6\x9d\x00\xc4\xb9\xb5\x39\xd5\xca\xa7\xbf\xd9\x16\xaf\x66\x2f\x67\xce\xa3\x17\x4c\x52\x14\x4b\x1c\x56\xfe\x75\x7b\x26\xf6\x5d\xb8\x55\x5e\x3e\x3d\x74\x5b\x9d\x8c\xef\x1b\xff\xaa\x1b\xa0\xa0\x0d\xec\xfb\x75\x9f\x86\x3e\xdb\xa8\x42\x1d\x51\xd7\xd3\x17\xa5\xf7\x5d\x91\x75\x55\xac\x35\x4e\x52\x43\x26\xbf\x5b\x43\x54\x8e\x06\xb6\xb2\xfd\x1f\x00\x00\xff\xff\x0f\xe9\x93\x20\x02\x07\x00\x00")

func resourcesCrdClusterJsonBytes() ([]byte, error) {
	return bindataRead(
		_resourcesCrdClusterJson,
		"resources/crd-cluster.json",
	)
}

func resourcesCrdClusterJson() (*asset, error) {
	bytes, err := resourcesCrdClusterJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/crd-cluster.json", size: 1794, mode: os.FileMode(420), modTime: time.Unix(1609623345, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesCrdResourceJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x55\x4d\x8f\xd3\x30\x10\x3d\x3b\xbf\xc2\xf2\xb9\x2a\x42\x5c\x50\x6f\x08\x24\xc4\x05\xad\x40\xda\x0b\xe2\x30\x49\x86\x60\x9a\xd8\x96\x67\x5c\x81\x56\xf9\xef\x68\xe2\x26\xed\xf6\x33\xed\xa9\x3e\xd5\x33\xef\xcd\x3c\xbf\x3a\x9e\x97\x42\x19\x08\xf6\x19\x23\x59\xef\xcc\x4a\xcb\x0e\xff\x32\x3a\xd9\xd3\x72\xfd\x9e\x96\xd6\xbf\xd9\xbc\x35\x8b\x42\x99\xb5\x75\xb5\x60\x3e\x26\x62\xdf\x7d\x43\xf2\x29\x56\xf8\x09\x7f\x59\x67\x59\xf8\x02\xea\x90\xa1\x06\x06\xb3\xd2\x2f\x85\x52\xc6\x41\x87\x42\x8a\x5b\x38\x2d\xd1\x11\x76\x65\x8b\x9e\xa4\xb8\x29\xb4\xd6\xba\x17\x2a\x05\xac\x46\x5a\x13\x7d\x0a\xc2\x3b\x40\x2f\x24\xb9\xc9\x7a\xc9\xac\xf4\x8f\x42\x29\x25\x8c\xbd\x56\x59\xae\x44\x08\xe3\x06\x45\x33\xc7\x84\x63\x8c\x7d\x84\x06\x0f\x82\xd5\x6f\xec\x46\xcd\x12\xf0\x01\xdd\x87\xa7\x2f\xcf\xef\xbe\xef\x32\xfa\xcc\x32\xfc\x2f\x0c\x8d\x7d\xf9\x07\x2b\x36\x8b\xf3\xc8\x10\x7d\xc0\xc8\x16\xe9\x62\xc5\x01\x3b\xd9\x71\x09\x75\x5b\xff\x7b\x74\x4c\x9c\x12\xaa\x35\x0e\x37\x60\x1e\xe1\x95\x34\xe2\x68\x5d\x63\x66\x11\xfb\xeb\x07\x18\x6a\x57\x6d\x22\xc6\xf8\x38\x82\xc6\x3b\xfe\x38\x8a\x02\x44\xe8\xe6\xff\xc7\xfa\xae\xdb\x34\x31\xa1\xae\x87\x87\x00\xda\xa7\xdb\xef\xd7\x51\xff\x5b\xfc\x90\xd5\xcf\x73\xee\x2a\xea\x32\xe2\x8a\xf3\x86\x18\x38\xcd\x3b\xf5\xb1\xd3\x6a\xbb\x0e\xbf\x4f\x35\x25\x7c\x99\xdf\xb4\xcf\xe8\x30\x02\xe7\x57\x7b\x0f\xa0\xa6\xa2\xd6\x31\x36\x18\xcd\x2e\xd7\x17\xbb\x5f\xf7\x19\x70\x3a\x73\x1c\x3d\x61\x92\xa1\x54\x4e\x33\xe0\xac\x3d\x7b\xf6\x9d\xa8\x5a\x9c\xde\xfd\x1c\x86\x02\x55\x3e\x1f\xfc\x2b\x74\x48\x01\x2a\xac\xf3\xb4\x90\xb1\x30\xd9\x68\x42\x9b\x22\xb4\xaf\x26\x52\x36\xde\x90\x75\x4d\x6a\x21\xee\xe7\xb6\xa9\x71\xf2\x8d\x33\x4f\x4c\xcd\xfd\xfb\xa2\xff\x1f\x00\x00\xff\xff\x1b\x0b\xd8\xb4\x45\x07\x00\x00")

func resourcesCrdResourceJsonBytes() ([]byte, error) {
	return bindataRead(
		_resourcesCrdResourceJson,
		"resources/crd-resource.json",
	)
}

func resourcesCrdResourceJson() (*asset, error) {
	bytes, err := resourcesCrdResourceJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/crd-resource.json", size: 1861, mode: os.FileMode(420), modTime: time.Unix(1609623532, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesSrvClusterRoleBindingJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x8e\x3f\x4f\x03\x31\x0c\x47\xf7\x7e\x8a\xc8\x33\x1c\x62\x43\xb7\x01\x03\x7b\x91\x58\x10\x83\x2f\xe7\x82\x69\xce\x8e\x1c\xa7\x03\xe8\xbe\x3b\x6a\x5a\x4a\x2b\xf1\x6f\x8b\xe2\xf7\xf4\x7e\xef\x8b\x10\x42\x00\xcc\xfc\x40\x56\x58\x05\xfa\x00\x36\x60\xec\xb0\xfa\x8b\x1a\xbf\xa1\xb3\x4a\xb7\xbe\x2a\x1d\xeb\xc5\xe6\x12\xce\x76\xc6\x9a\x65\xdc\xb2\xb7\xa9\x16\x27\x5b\x6a\xa2\x1b\x96\x91\xe5\xf9\x93\x98\xc8\x71\x44\x47\xe8\xc3\xae\xd2\x7e\x05\x27\xda\x7a\x24\x85\xa6\x21\xd1\xb9\x66\x32\x74\x35\x68\xcc\xbc\x97\x4d\x13\x2d\x69\x75\xea\x62\xe6\x3b\xd3\x9a\x7f\xdd\xb8\xcf\xff\x34\xf2\xf8\xfc\xcf\x2d\xa5\x0e\xaf\x14\xbd\x40\x1f\x1e\x0f\xf2\xd7\xac\x93\xd2\x3d\xd9\x86\x23\x5d\xc7\xa8\x55\xfc\x28\xf6\x47\xf0\x1b\xb0\x64\x8c\x8d\x1e\x69\x85\x35\x39\x1c\x90\xb9\xbd\x9e\x16\xf3\x47\x00\x00\x00\xff\xff\x29\x9b\xba\x4d\xbe\x01\x00\x00")

func resourcesSrvClusterRoleBindingJsonBytes() ([]byte, error) {
	return bindataRead(
		_resourcesSrvClusterRoleBindingJson,
		"resources/srv-cluster-role-binding.json",
	)
}

func resourcesSrvClusterRoleBindingJson() (*asset, error) {
	bytes, err := resourcesSrvClusterRoleBindingJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/srv-cluster-role-binding.json", size: 446, mode: os.FileMode(420), modTime: time.Unix(1608931048, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesSrvClusterRoleJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xdc\x90\x31\x4f\xc3\x30\x10\x85\xf7\xfe\x0a\xcb\x23\x0a\xad\xd8\x50\x57\x06\x76\x06\x16\xc4\x70\x71\x8e\x62\x35\xf1\x59\x77\xe7\x08\x81\xf2\xdf\x51\x12\x92\x2a\x28\xce\x50\x36\x46\xfb\x7b\xef\xe9\xde\xfb\xda\x19\x63\x8c\x85\xe8\x9f\x91\xc5\x53\xb0\x47\x63\xb9\x04\xb7\x87\xa4\xef\xc4\xfe\x13\xd4\x53\xd8\x9f\xef\x65\xef\xe9\xd0\xde\xd9\x62\x74\x9c\x7d\xa8\x7a\xed\x43\x9d\x44\x91\x9f\xa8\xc6\x09\x35\xa8\x50\x81\x82\x3d\x9a\x31\x7e\xf8\x0d\xd0\x60\x6f\xc0\x20\xd8\x94\x35\xde\x52\x44\x06\x25\xb6\x83\xa6\xfb\x31\x73\xaa\x51\xec\xd1\xbc\xcc\xce\x4b\xc6\x74\xea\x23\x53\x8a\x4b\xd1\x8c\xa7\x78\x92\xfe\x60\xbb\x10\xbc\x16\xcb\x28\x46\xa1\xc4\x0e\x33\x51\x6e\xac\x26\xb6\xc8\xb3\x83\x28\x68\x5a\x95\x5c\xd2\xb7\xe0\x46\x40\xa0\x6a\xdd\x3c\x80\xc9\xb8\xd9\xb0\x45\x2e\x33\xed\x6e\x7e\x39\xe7\x57\x57\x5c\x39\xfd\x9f\xc6\x8e\x54\xad\x76\xed\xff\x0f\xf8\x81\x6e\x0d\x0a\x72\xeb\x33\x0b\x63\xa8\x22\xf9\xa0\xeb\xb0\xc5\x0c\x71\x14\xde\xfc\xa9\x81\xf8\x6f\x96\x15\x74\x8c\x7a\x7d\x9d\x13\x6a\xb6\xd0\x6e\x7c\x77\xdf\x01\x00\x00\xff\xff\xc7\x14\x4b\x0e\x47\x04\x00\x00")

func resourcesSrvClusterRoleJsonBytes() ([]byte, error) {
	return bindataRead(
		_resourcesSrvClusterRoleJson,
		"resources/srv-cluster-role.json",
	)
}

func resourcesSrvClusterRoleJson() (*asset, error) {
	bytes, err := resourcesSrvClusterRoleJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/srv-cluster-role.json", size: 1095, mode: os.FileMode(420), modTime: time.Unix(1608931048, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesSrvDeploymentJsonTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x54\x4d\x6f\xd3\x40\x10\xbd\xf7\x57\x58\x3e\x07\x70\x0e\x5c\x72\x8b\x40\x48\x48\x50\xaa\x4a\xe5\x82\x38\x6c\x37\x4f\x61\xc5\xec\x07\x3b\xe3\xc2\x2a\xf2\x7f\x47\x6b\xb7\x49\x8c\xd7\x2e\x91\x9a\x53\x3c\xf3\xde\xcc\x7b\x33\x63\x1f\xae\xaa\xaa\xaa\x6a\x15\xcc\x57\x44\x36\xde\xd5\x9b\xfc\x14\xf8\xcd\xc3\xba\x5e\x0d\xc9\x9f\xc6\xed\x72\xf8\x3d\x02\xf9\x64\xe1\xe4\x29\x63\x21\x6a\xa7\x44\xd5\x9b\x6a\x28\xd4\x47\x49\xdd\x83\x78\x14\x7b\x6c\x12\x72\x19\x38\x86\xbd\x27\xbc\xf2\x01\x51\x89\x8f\xf5\x11\xd6\xad\x4e\x55\x9c\xb2\x58\x80\x3f\x42\x6b\x0e\xd0\xe3\xf6\x11\x81\x8c\x56\x59\xc0\xfa\xac\x1e\x83\xa0\x33\x7d\xa2\xcb\x2a\xd1\x3f\x3e\x95\x45\xff\xa7\xf0\x5e\x51\xd1\x06\x4b\x54\x82\x7d\x9a\xb6\x95\x14\x7a\x83\xb7\x9e\xc8\xb8\xfd\x5d\xd8\x29\x41\x79\x16\x02\x1b\x28\x67\xa7\xda\x4b\x1b\x78\x6e\x13\x17\x1a\x1b\x9b\xfb\x47\x5b\x55\x5c\xc2\x29\x83\xf8\x60\x34\xb6\x5a\xfb\xd6\xc9\xf5\xec\x4e\x57\x53\xaa\xf6\x4e\x94\x71\x88\xd9\xc0\xb7\xa2\x81\xb2\xad\x9e\x6d\xac\xda\xf7\xbd\x0e\x87\xd7\x1f\xf3\xff\xae\x2b\x34\x19\xc3\x6f\x5a\xa2\x1b\x4f\x46\xa7\x73\xe2\x29\xba\x5c\xe2\xe9\x60\xf3\x50\x17\x60\x0c\xdd\x46\x23\xe9\x9d\x77\x82\x3f\x32\xbb\x9c\x23\x21\x42\xed\xbe\x38\x4a\xb7\xde\xcb\x07\x43\xe0\xc4\x02\x5b\x6f\x2a\x89\x2d\xe6\x1b\x0d\xdc\xd6\x6d\xf9\xda\xbb\xcc\xbd\x80\x71\xc7\xc8\xef\xc9\xba\x69\x9a\x59\x74\xb7\xe0\x31\x82\x7d\x1b\x35\xe6\x4f\xef\x08\x25\x63\x8d\x3c\x8f\xeb\xb1\x3a\xb4\x79\xc0\xeb\xa6\xb1\x0b\x13\x3e\xc2\x2d\xac\x8f\xfd\x2a\xdf\x36\x9f\xcd\xf4\xa6\xcf\x7f\x0b\x6e\xaa\xc1\xd1\xaf\x16\x7c\xb1\xd0\x97\xd7\x39\xbf\x8f\x62\x66\x1a\xfd\x5e\x78\xd5\x22\x58\x54\x94\xd3\xed\x6f\xe9\xb7\x4a\x3c\xfb\x7d\x1b\xbe\xc0\x57\xdd\xdf\x00\x00\x00\xff\xff\xcf\x31\x12\x8a\x3e\x06\x00\x00")

func resourcesSrvDeploymentJsonTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesSrvDeploymentJsonTemplate,
		"resources/srv-deployment.json.template",
	)
}

func resourcesSrvDeploymentJsonTemplate() (*asset, error) {
	bytes, err := resourcesSrvDeploymentJsonTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/srv-deployment.json.template", size: 1598, mode: os.FileMode(420), modTime: time.Unix(1608931048, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesSrvServiceAccountJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x34\xcc\x3d\x0e\xc2\x30\x0c\x05\xe0\xbd\xa7\xb0\x3c\xc3\xc0\xda\x8d\x33\x20\xb1\x9b\xf4\x0d\x16\xc4\xae\x5c\xd3\x05\xf5\xee\x28\x89\xba\x7e\xef\xe7\x37\x11\x11\xb1\xac\xfa\x44\x6c\xea\xc6\x33\xf1\x7e\xe3\xcb\xf0\xb7\xda\xd2\xe4\x81\xd8\xb5\xe0\x5e\x8a\x7f\x2d\xcf\xb4\x22\x65\x91\x14\x9e\x69\xfc\x74\x35\xa9\x68\x1b\xd8\x86\xfa\xfa\xe0\xea\x2b\x42\xd2\x83\x7b\xe7\x98\x8e\x7f\x00\x00\x00\xff\xff\x30\xd5\xc7\x81\x75\x00\x00\x00")

func resourcesSrvServiceAccountJsonBytes() ([]byte, error) {
	return bindataRead(
		_resourcesSrvServiceAccountJson,
		"resources/srv-service-account.json",
	)
}

func resourcesSrvServiceAccountJson() (*asset, error) {
	bytes, err := resourcesSrvServiceAccountJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/srv-service-account.json", size: 117, mode: os.FileMode(420), modTime: time.Unix(1608931048, 0)}
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
	"resources/crd-cluster.json":              resourcesCrdClusterJson,
	"resources/crd-resource.json":             resourcesCrdResourceJson,
	"resources/srv-cluster-role-binding.json": resourcesSrvClusterRoleBindingJson,
	"resources/srv-cluster-role.json":         resourcesSrvClusterRoleJson,
	"resources/srv-deployment.json.template":  resourcesSrvDeploymentJsonTemplate,
	"resources/srv-service-account.json":      resourcesSrvServiceAccountJson,
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
		"crd-cluster.json":              &bintree{resourcesCrdClusterJson, map[string]*bintree{}},
		"crd-resource.json":             &bintree{resourcesCrdResourceJson, map[string]*bintree{}},
		"srv-cluster-role-binding.json": &bintree{resourcesSrvClusterRoleBindingJson, map[string]*bintree{}},
		"srv-cluster-role.json":         &bintree{resourcesSrvClusterRoleJson, map[string]*bintree{}},
		"srv-deployment.json.template":  &bintree{resourcesSrvDeploymentJsonTemplate, map[string]*bintree{}},
		"srv-service-account.json":      &bintree{resourcesSrvServiceAccountJson, map[string]*bintree{}},
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
