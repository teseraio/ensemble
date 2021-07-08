// Code generated by go-bindata. (@generated) DO NOT EDIT.

// Package k8s generated by go-bindata.// sources:
// resources/config-map.template
// resources/crd-cluster.json
// resources/crd-resource.json
// resources/crd.template
// resources/generic.template
// resources/headless-service.template
// resources/kind-cluster.template
// resources/kind-resource.template
// resources/mock-crd.template
// resources/mock.template
// resources/pod.template
// resources/srv-cluster-role-binding.json
// resources/srv-cluster-role.json
// resources/srv-deployment.json.template
// resources/srv-service-account.json
// resources/volume-claim.template
package k8s

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
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
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

// ModTime return file modify time
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

var _resourcesConfigMapTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\x8d\xb1\x0e\xc2\x30\x0c\x44\x67\xfb\x2b\x22\xcf\xa8\x12\x6b\x57\x58\x61\x60\x60\xb7\x88\x8b\x22\x94\xa4\x4a\x03\x8b\xe5\x7f\x47\x86\x2e\x55\xc7\x3b\xdd\xbd\xa7\x08\xc4\x73\xba\x4b\x5b\x52\x2d\x34\x06\xfa\x1c\xe9\x80\x40\xaf\x54\xa2\xc7\x53\x2d\x53\x7a\x5e\x78\xfe\xb5\x59\x3a\x47\xee\x4c\x63\x50\x04\x50\x4d\x53\x18\x6e\xb2\xd4\x77\x7b\xc8\x0a\x31\x43\x00\x6a\xdb\xd2\x51\xaa\xfb\xa9\x53\x41\x55\x4a\xfc\xdf\x0a\x67\x59\xb7\x57\xce\x62\x46\x08\xe6\xe6\x8d\x75\x38\x73\x67\x3f\x18\xda\x37\x00\x00\xff\xff\x99\x93\x36\x69\xc2\x00\x00\x00")

func resourcesConfigMapTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesConfigMapTemplate,
		"resources/config-map.template",
	)
}

func resourcesConfigMapTemplate() (*asset, error) {
	bytes, err := resourcesConfigMapTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/config-map.template", size: 194, mode: os.FileMode(436), modTime: time.Unix(1615882763, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesCrdClusterJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xd4\x57\x4f\x8b\xdb\x3e\x10\x3d\xcb\x9f\x42\xe8\x1c\xf2\xe3\x47\x2f\x25\xb7\xd2\x42\xe9\xa5\x2c\x2d\xec\x65\xd9\xc3\xc4\x9e\xa6\x6a\x6c\x49\x1d\x49\xa1\xcb\xe2\xef\x5e\xe4\x3f\x5a\xe7\xef\x4a\x36\x14\xd6\xa7\x58\x33\x6f\xf4\xf4\xf4\x32\xb2\x9e\x0b\x26\xc0\xc8\x7b\x24\x2b\xb5\x12\x1b\x1e\xde\xf0\x8f\x43\x15\xde\xed\x7a\xff\xde\xae\xa5\xfe\xef\xf0\xbf\x58\x15\x4c\xec\xa5\xaa\x42\xce\x47\x6f\x9d\x6e\xbe\xa1\xd5\x9e\x4a\xfc\x84\x3f\xa4\x92\x2e\xe0\x43\x52\x83\x0e\x2a\x70\x20\x36\xfc\xb9\x60\x4c\x28\x68\x30\x80\xca\xda\x5b\x87\x64\xd7\xa8\x2c\x36\xdb\x1a\xb5\x0d\xb5\x45\xc1\x39\xe7\x6d\x40\x5a\x83\xe5\x88\xda\x91\xf6\x26\xc0\x4e\xb2\x57\x21\x78\xe8\xe9\x5a\xb1\xe1\x0f\x05\x63\x2c\x20\x26\x33\xf5\x6c\xc3\x88\x45\x3a\x60\xa0\xec\xc8\xe3\x38\xe6\x34\xc1\x0e\x4f\x06\xcb\x9f\xd8\x8c\x94\xc3\x80\x36\xa8\x3e\xdc\x7d\xb9\x7f\xf7\xfd\x25\xc2\xaf\x3c\xc2\x3d\x99\x6e\x62\xbd\xfd\x85\xa5\x13\xab\xeb\x99\x86\xb4\x41\x72\x12\xed\xcd\x8a\x5d\x6e\x94\xe3\x56\x56\xde\xfc\x73\x78\x44\xcc\x16\xca\x3d\x76\x06\x48\x03\xcc\xa3\xb6\x84\x62\xc4\x0e\x4e\xc8\x43\x1d\xf1\xb5\x8e\xa4\xda\x89\xac\x02\x6d\x72\x76\x9b\xa1\x03\xe1\x6f\x2f\xa9\x73\xf1\x43\xbf\xb0\xc7\x24\x70\xe2\x1c\xfd\x3f\x2d\x4f\xe3\xa8\x12\x10\xc1\x53\xce\xa6\x4a\x87\xcd\x8c\xfd\x9c\xed\x22\xbe\xd0\x49\x7c\x91\x9b\xf8\x52\x47\xf1\x3c\xaf\x9c\xce\xf9\x86\x18\x13\x9a\x5a\x96\x30\x6f\x87\xf8\x94\xb5\x54\x0e\x77\x48\xff\x86\xb6\x01\x82\x19\x86\x8e\xf8\x45\xc6\x8e\x55\xa0\xaa\xba\xf3\x17\xea\xbb\x65\x56\x3f\xe3\x35\xd7\x02\x3c\xab\x1d\xce\x43\x64\xee\xd7\x51\x23\x5d\x60\xd0\x2c\x68\x5a\xab\xe6\xc9\x6b\x4f\x6d\xea\x15\x1a\x54\xd5\x1b\xe9\xea\xb9\x26\x4b\x94\xea\xd5\xac\x04\x31\x8f\x4f\xdf\xf1\x0b\xe8\xf6\xae\xbe\x52\x56\x58\x07\xce\xa7\xc9\x75\xde\x1f\xd8\xf0\x9c\x1e\x6a\x2c\x06\xf4\xb6\xff\xea\xfd\x8c\x0a\x09\x5c\xff\x59\x3f\x49\x60\xe7\x9d\x32\x86\xda\xe2\xe5\xd7\xcd\x25\x5e\x8d\x5e\x8e\x9c\x8f\x5e\x10\x49\x58\xbf\xa5\xe1\x4e\x71\x5d\x9e\x89\x7c\x17\xaa\x16\x97\xdf\x1e\xbb\x6b\x83\x2d\x75\xbf\xf0\xaf\xd0\xa0\x35\x50\x62\xd5\xdf\x27\xc2\x01\x1f\x65\x14\xa6\xf6\x04\xf5\xf4\xca\xd2\xeb\x2e\xac\x54\x3b\x5f\x03\x4d\x42\x43\x24\x5e\x8c\x86\xd1\x62\x14\xb0\x2d\xda\xbf\x01\x00\x00\xff\xff\x9a\x82\x55\x92\x63\x0d\x00\x00")

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

	info := bindataFileInfo{name: "resources/crd-cluster.json", size: 3427, mode: os.FileMode(436), modTime: time.Unix(1625668177, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesCrdResourceJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x55\x4d\x8f\xd4\x30\x0c\x3d\xa7\xbf\x22\xf2\x79\x34\x08\x71\x41\x73\x43\x20\x21\x2e\x68\x05\xd2\x5e\x56\x7b\xc8\xb4\xa6\x84\x6d\x93\x60\x27\x23\xd0\xaa\xff\x1d\xb9\x9d\x66\xba\xf3\xd1\xe9\xec\x89\x9c\x12\xfb\x3d\xdb\x79\x49\x9c\xe7\x42\x81\x09\xf6\x1e\x89\xad\x77\xb0\xd1\xb2\xc2\x3f\x11\x9d\xac\x79\xfd\xf4\x9e\xd7\xd6\xbf\xd9\xbd\x85\x55\xa1\xe0\xc9\xba\x4a\x30\x1f\x13\x47\xdf\x7e\x43\xf6\x89\x4a\xfc\x84\x3f\xac\xb3\x51\xf8\x02\x6a\x31\x9a\xca\x44\x03\x1b\xfd\x5c\x28\x05\xce\xb4\x28\x24\xda\xc3\x79\x8d\x8e\xb1\xdd\x36\xe8\x59\x82\x43\xa1\xb5\xd6\x9d\x50\x39\x60\x39\xd2\x6a\xf2\x29\x08\xef\x08\xbd\x12\xe7\x6e\xa8\x97\x61\xa3\x1f\x0a\xa5\x94\x30\x26\xa9\x86\x72\xc5\xc2\x48\x3b\x94\x9a\x23\x25\x1c\x6d\xd1\x93\xa9\xf1\xc8\x58\xfe\xc4\x76\xac\x59\x0c\x3e\xa0\xfb\x70\xf7\xe5\xfe\xdd\xf7\x83\x47\x5f\x18\x10\xff\x86\x3e\xb1\xdf\xfe\xc2\x32\xc2\xea\x32\x32\x90\x0f\x48\xd1\x22\xcf\x46\xec\xb1\x59\x8e\x39\xd4\x6d\xf9\x5f\x53\x47\xe6\x94\x4d\xe2\x88\xb4\x98\xf0\xa2\x34\x8e\x64\x5d\x0d\x8b\x88\xdd\xf5\x0d\xf4\xb1\xc7\x2b\xf5\xff\x54\x14\x0c\x99\x76\xb9\xa4\xfa\x55\x87\x97\x99\xa6\xaa\xfa\x77\x67\x9a\xbb\xdb\x8f\xf3\x24\xff\x2d\x7a\xc8\xe8\x96\x29\x77\x15\xb5\x40\x5b\x20\xfc\x9d\x2c\xf5\x2f\xf9\x21\xdf\xc3\xd5\xe4\x06\x3c\xce\xc6\xb8\x92\x02\x38\x9a\x98\x96\x49\x77\x7a\x5c\x6a\x3f\x8e\xdf\x94\xca\x0e\xbf\x1d\xfa\xd0\x67\x74\x48\x26\x0e\x9d\x76\x02\x50\x39\xa8\x75\x11\x6b\x24\x38\xf8\xba\xe2\x30\x9b\xdd\xe2\x45\xef\x79\xcf\xa9\xf5\x8c\x48\xc0\x69\x9b\xfb\xf6\x45\x79\x26\xf2\x9d\x89\x5a\x9c\x5f\x3d\xf6\x8d\x9c\x4b\x3f\x6c\xfc\xab\x69\x91\x83\x29\xb1\x1a\x3a\xbc\xb4\xf2\x2c\x23\x84\x26\x91\x69\x5e\xfc\x22\x83\xf0\xc0\xd6\xd5\xa9\x31\x34\xf5\xed\x5d\xe3\x6f\x35\xfe\x53\x22\xea\x90\xbf\x2b\xba\x7f\x01\x00\x00\xff\xff\x3c\x99\xd5\xb7\xf9\x06\x00\x00")

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

	info := bindataFileInfo{name: "resources/crd-resource.json", size: 1785, mode: os.FileMode(436), modTime: time.Unix(1623687138, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesCrdTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x93\x41\x4b\xc3\x30\x14\xc7\xcf\xed\xa7\x28\x39\x8f\x8a\x78\x91\xdd\x44\x61\x88\x20\xc3\xc2\x2e\xe2\x21\x6b\x9f\x35\x6e\x4d\x42\xf2\x32\x94\x90\xef\x2e\x69\xb2\xb6\x59\xeb\x7a\xea\xcb\xff\xf7\x5e\xde\x4b\xfe\xb1\x79\x51\x14\x05\xa1\x92\xed\x40\x69\x26\x38\x59\xf7\x11\xfc\x20\x70\x1f\xeb\xf2\x70\xaf\x4b\x26\x6e\x4e\xb7\x64\x15\xe0\x03\xe3\x8d\xc7\x1e\x8d\x46\xd1\xbd\x81\x16\x46\xd5\xf0\x04\x9f\x8c\x33\xf4\x25\x22\xd7\x01\xd2\x86\x22\x25\xeb\x22\x6c\xd3\xaf\x72\xda\x81\xcf\xb6\xb6\xdc\x1e\x8d\xa2\x47\xe7\x4a\x6b\xcb\x8d\x12\x46\x3a\x47\x7a\xd0\xc5\x0a\x5a\x42\x9d\x66\xb7\x1e\x8b\xe9\xe7\x94\xd5\x28\x9f\xc2\x10\x9a\xac\x8b\xf7\x61\xd5\x7f\x36\x89\x92\x46\x86\xc1\x12\x55\x83\x3a\x81\x1f\x13\x95\x81\x25\x1d\x85\xa2\x2d\x5c\x01\xea\x2f\xe8\xd2\xd9\x13\x5d\x48\xe0\x0f\xdb\xe7\xdd\x5d\x75\x1d\xec\x61\xfc\x95\x7d\xaf\x62\xff\x0d\x35\x2e\xf4\x3b\x90\x52\x09\x09\x0a\x19\xe8\xbe\x62\x16\xbe\xe1\x24\x6d\x59\x49\xa8\x2b\x54\xce\xad\x46\x11\x29\x9a\x84\xcf\xe6\x3b\x0e\xca\xf2\x0e\x59\x46\xc4\x3e\x1c\xda\x06\x38\x28\x8a\xc1\x4c\x13\x60\x2c\xca\x38\x42\x0b\x8a\x8c\x9a\xcb\x67\x7f\xee\xdf\x21\x97\x95\x90\x18\xc7\x22\xda\xec\x55\x74\xe6\xa4\xcf\xc9\xa8\x11\xcf\x97\x0b\x7f\x4c\x4c\xa5\x6b\x11\xda\x7e\xa5\x1d\x68\x49\x6b\x68\xa6\x9e\xf3\x3e\xd2\xb3\xeb\x23\xb2\x37\xf7\x85\xd3\x2f\x6e\x8e\x68\xc6\x5b\x73\xa4\x2a\x62\x55\x0c\xe7\xe0\xf9\xcd\x59\x5b\xbe\x30\xde\x9c\x1f\xca\xd8\xb4\xcb\xdd\x5f\x00\x00\x00\xff\xff\x58\x72\xf1\xe9\xcc\x03\x00\x00")

func resourcesCrdTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesCrdTemplate,
		"resources/crd.template",
	)
}

func resourcesCrdTemplate() (*asset, error) {
	bytes, err := resourcesCrdTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/crd.template", size: 972, mode: os.FileMode(436), modTime: time.Unix(1615882783, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesGenericTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8e\xbd\xae\x83\x30\x0c\x85\xe7\xe4\x29\x90\xe7\x2b\x1e\x80\xf9\x6e\xf7\x67\x68\xa5\xaa\xab\x0b\xae\x14\x95\x84\x88\x84\xc9\xf2\xbb\x57\x01\x53\x40\xed\x66\x1d\x9f\xef\xe8\x63\x6b\x00\xa3\xbb\xd0\x98\xdc\x10\xa0\xa9\x80\xb9\xfe\x1e\x3c\xba\x20\x02\x5f\xd6\xc0\xc3\x85\x4e\xf3\x1f\x17\xba\x25\x65\x76\xf7\xaa\x3e\x47\x6a\x45\xac\x81\x14\xa9\x85\xa6\x62\xf6\x18\xd7\x78\x6e\x51\x01\x5e\xf5\x8c\x79\x4a\x0b\x30\x9f\x3b\x44\x5f\x7b\x08\x3c\x65\xec\x30\x63\xa9\x59\xa3\x23\x27\x4a\xc3\x34\xb6\xa4\xc6\xa5\x68\x60\x3c\x86\xaa\xfb\x56\x2d\xe6\xdb\xbe\x0e\xfe\xe2\x8d\xfa\xb4\xec\xf4\xf3\xbd\x69\xad\xbf\x4f\xdc\x9f\xda\x69\xe4\x31\x5e\xf7\xe1\x11\x81\x80\x9e\x54\xeb\x1f\x3d\x89\x80\x35\x62\xe5\x19\x00\x00\xff\xff\x90\x91\x81\x59\x7f\x01\x00\x00")

func resourcesGenericTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesGenericTemplate,
		"resources/generic.template",
	)
}

func resourcesGenericTemplate() (*asset, error) {
	bytes, err := resourcesGenericTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/generic.template", size: 383, mode: os.FileMode(436), modTime: time.Unix(1615737734, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesHeadlessServiceTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\xcd\xbd\x0a\xc3\x30\x0c\x04\xe0\x59\x7a\x8a\xa0\xb9\x14\xba\x66\xef\xd0\xa5\x14\x0a\xdd\x1d\xe7\x06\x13\xff\x04\xdb\xcd\x62\xfc\xee\x45\x21\x1d\x0a\x1d\x25\x7d\xba\x6b\x4c\x62\x56\xf7\x42\x2e\x2e\x45\x19\x07\xd9\x2e\x72\x62\x92\xc5\xc5\x59\xc7\x27\xf2\xe6\x2c\xf6\x5d\x40\x35\xb3\xa9\x46\xc6\xa1\x31\x91\x44\x13\xa0\xa6\xb5\xf3\x35\x16\x84\xc9\xa3\x77\x95\x24\xde\x4c\xf0\xe5\x80\x24\x38\xce\xaa\xd3\x22\x4c\xd4\x99\xba\x66\x96\x15\xf6\x9b\x57\xe0\x61\x6b\xca\x7f\xdf\x7e\x4b\x34\x61\x2f\xb2\xfe\x5d\x2a\xf2\xed\xa1\xe6\x9e\x22\x84\xa9\x73\xff\x04\x00\x00\xff\xff\xd5\x07\x51\xcd\xd8\x00\x00\x00")

func resourcesHeadlessServiceTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesHeadlessServiceTemplate,
		"resources/headless-service.template",
	)
}

func resourcesHeadlessServiceTemplate() (*asset, error) {
	bytes, err := resourcesHeadlessServiceTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/headless-service.template", size: 216, mode: os.FileMode(436), modTime: time.Unix(1615882789, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesKindClusterTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\x8e\x31\xae\x83\x30\x10\x44\x7b\x4e\xb1\xda\xfa\x8b\xaf\xb4\x94\x49\x9f\x22\x45\x9a\x28\xc5\x02\x53\x58\x60\x83\x58\x27\x8d\xe5\xbb\x47\xd8\x20\xe2\x94\x33\x7e\xf3\xbc\xa1\x22\x22\x62\x99\xcd\x1d\x8b\x9a\xc9\x71\x43\x0c\xa7\xb0\xed\x88\x49\xb5\x36\xd3\xff\xfb\xc4\x7f\x19\x1b\x8c\xeb\x57\xe0\x32\xbe\xd4\x63\xd9\x6b\x0b\x2f\xbd\x78\xe1\x86\xb2\x2f\xb5\x4e\x2c\x56\x38\x84\xfa\x2a\x16\x31\x72\x7a\x8b\xdb\x48\x67\x74\xe5\xa0\x95\x6e\x40\xfa\xe0\x28\x7f\x4d\xe7\xcc\xec\xb2\x2f\x61\x96\xc2\x2b\x37\xf4\x28\xf6\xa5\x2d\x71\x0b\xe6\xd1\x74\xb2\xb2\x21\xd4\xb7\x2d\xc5\x58\x90\x47\x7a\xe6\xcb\xab\xf8\x09\x00\x00\xff\xff\x45\xd9\xe6\x39\x2f\x01\x00\x00")

func resourcesKindClusterTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesKindClusterTemplate,
		"resources/kind-cluster.template",
	)
}

func resourcesKindClusterTemplate() (*asset, error) {
	bytes, err := resourcesKindClusterTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/kind-cluster.template", size: 303, mode: os.FileMode(436), modTime: time.Unix(1623687138, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesKindResourceTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\x8e\xab\xae\xc3\x30\x10\x44\x79\xbe\x62\xb5\xf8\x2a\x57\xa5\xa1\xe5\x05\x05\xe5\xdb\x64\x80\xd5\xd8\x8e\xbc\x49\x89\xb5\xff\x5e\xc5\x8f\x4a\x85\x1e\x9f\x33\xb3\x79\x20\x22\x62\xd9\xdc\x03\x49\x5d\x0c\x3c\x11\x23\x28\xfc\x73\x45\x54\x1d\x5d\xfc\x7f\x5f\xf8\xaf\x62\x2f\x17\x96\x13\xb8\x43\xe3\x91\x66\xf4\xdc\x63\x97\x45\x76\xe1\x89\x6a\x61\x49\x83\x78\x9c\x74\xce\xe3\x4d\x3c\xcc\xb8\xfc\x59\x93\x74\xc3\xfc\x2b\xcc\xeb\xa1\x3b\x52\x73\xae\xf5\x65\xd6\x56\x0a\x92\xfa\x72\x65\xfa\x21\xdf\xee\xc1\x3e\x01\x00\x00\xff\xff\x1c\x77\xd4\xf7\xd2\x00\x00\x00")

func resourcesKindResourceTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesKindResourceTemplate,
		"resources/kind-resource.template",
	)
}

func resourcesKindResourceTemplate() (*asset, error) {
	bytes, err := resourcesKindResourceTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/kind-resource.template", size: 210, mode: os.FileMode(436), modTime: time.Unix(1623687138, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesMockCrdTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x54\x3d\x6f\xdb\x30\x10\x9d\xc9\x5f\x41\xdc\x6c\xa8\x28\xba\x14\xde\x8a\x76\xe9\x52\x14\x2d\xe0\xa5\xe8\x40\xcb\x17\x85\xb1\x44\x12\xbc\xa3\x91\x40\xd0\x7f\x0f\x4e\xb2\x9c\x38\x91\x64\xc7\x9a\xc4\xf7\x71\x7c\x38\xe1\xa9\xd5\x0a\x6c\x74\x1b\x4c\xe4\x82\x87\xb5\x91\x13\x3e\x32\x7a\x39\x53\xb1\xff\x4a\x85\x0b\x9f\x0e\x9f\x61\xa5\x15\xec\x9d\xdf\x89\xe6\x7b\x26\x0e\xcd\x1f\xa4\x90\x53\x89\x3f\xf0\xce\x79\xc7\xe2\x17\x51\x83\x6c\x77\x96\x2d\xac\x4d\xab\x95\x02\x6f\x1b\x14\x53\xdb\x16\xf2\xda\x75\x54\x34\xa1\xdc\x17\x2e\x80\x36\xc6\x98\x4e\x4c\x14\xb1\x1c\x0d\x55\x0a\x39\x8a\x63\x94\xad\x04\x3d\x0c\x11\x09\xd6\xe6\x9f\x56\x4a\x89\xf4\xd5\xf4\x21\xa1\x20\x84\xe9\x80\x12\x93\x53\xc6\x11\xe3\x90\x6c\x85\x6f\xc0\xf2\x1e\x9b\x31\xa6\x00\x21\xa2\xff\xf6\xfb\xe7\xe6\xcb\xdf\x17\xc6\xcc\x3c\xc0\x4f\xb1\xbf\x38\x6c\x1f\xb0\x64\x58\xcd\x2b\x63\x0a\x11\x13\x3b\xa4\xc5\x89\xbd\xf6\xb4\x87\x25\xd5\xc7\xee\xbf\x25\xc7\xc9\xb3\xbc\x84\xd9\x50\xce\x33\x56\x98\xe0\x2a\x67\x77\x39\x7b\x3f\x7c\x7b\x5b\x16\xe2\xe4\x7c\x75\x65\x94\x8b\xaa\x65\xc5\x3c\x3b\xcd\xbc\x47\x27\x96\x01\x94\xb7\xe9\xd8\xb5\xf9\x4f\x07\xc4\x96\x73\xcf\x4f\x4c\xd5\xd3\xa7\xff\x7d\xb5\xa8\x0c\xc3\xaa\x7e\xd9\x06\x29\xda\x12\x77\x43\xe7\xa4\x5c\x34\x16\x04\x62\x9d\x93\xad\xcf\xab\x3c\x94\x0e\xc8\xf9\x2a\xd7\x36\x9d\x91\x47\x6e\xfc\x69\xb4\x2d\x3b\xae\xd1\x8c\xac\x56\x6a\x08\xd2\xe9\xee\x39\x00\x00\xff\xff\x2b\xb3\x4a\xb8\x87\x04\x00\x00")

func resourcesMockCrdTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesMockCrdTemplate,
		"resources/mock-crd.template",
	)
}

func resourcesMockCrdTemplate() (*asset, error) {
	bytes, err := resourcesMockCrdTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/mock-crd.template", size: 1159, mode: os.FileMode(436), modTime: time.Unix(1623687138, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesMockTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x3c\xcc\xbf\x0a\x03\x21\x0c\xc7\xf1\xfd\x9e\x22\x64\x2e\x57\xba\xde\x83\x74\x0f\x9a\x21\x88\x46\xd4\x76\x91\xbc\x7b\xf1\x4f\x6f\x0b\xf9\xf0\xfd\xf5\x03\x00\x00\x29\xcb\x9b\x4b\x15\x4d\x78\x01\x46\x75\xe1\x14\x7d\x7e\x5f\xf8\x58\x1e\x24\xf9\x21\xbd\x9f\x85\xab\x7e\x8a\x63\xb3\x3f\x46\x6e\xe4\xa9\x11\x5e\xb0\xe6\xe6\x37\x51\xe4\x9d\x8c\xd3\x0c\xa7\xd9\x8e\x6a\x66\x77\x07\x76\xd8\x2f\x00\x00\xff\xff\xf1\x48\xe0\x7d\x89\x00\x00\x00")

func resourcesMockTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesMockTemplate,
		"resources/mock.template",
	)
}

func resourcesMockTemplate() (*asset, error) {
	bytes, err := resourcesMockTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/mock.template", size: 137, mode: os.FileMode(436), modTime: time.Unix(1623687138, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesPodTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x54\x4d\x8b\xdb\x30\x10\x3d\xcb\xbf\x42\x88\x1c\x53\x43\xaf\x81\x1e\x96\x76\xa1\x3d\x74\x1b\x7a\x58\x0a\x65\x0f\x13\x7b\x9a\x35\xab\x8f\x20\xc9\x2e\x41\xe8\xbf\x97\x91\x25\x5b\x6d\x53\x53\xf6\x14\x8f\xde\xd3\x93\xde\xd3\x4c\x42\xc3\x04\x5c\x86\x47\xb4\x6e\x30\x5a\x1c\xb8\x98\xde\x8a\x7d\xc3\xc4\xcb\xa0\x7b\x2a\x8f\xa6\x4f\xb5\x42\x0f\x3d\x78\x10\x07\x1e\x1a\xc6\x84\x06\x85\x84\x87\xd0\x7e\xfa\x10\x23\x71\x98\x90\x70\x42\xe9\x32\x85\x09\xd4\x0e\xd5\x49\x16\xde\x7d\x2e\x63\x14\x0d\x63\xb1\x61\x91\x84\xdd\x05\xbb\x22\xfa\x6c\x9c\xaf\x84\x3f\xe6\xb2\xc8\xbb\xf1\xd4\x1b\x05\x83\xfe\x5b\x30\xe1\x16\x9d\x07\xeb\x8f\x46\x0e\xdd\x95\x38\x0f\x38\xa1\x9d\xb1\xce\x68\x0f\x83\x46\x4b\xd7\xfb\x4e\xd7\x4b\x77\x64\x21\x0c\x3f\x78\x7b\xaf\xa7\x18\x53\x2d\x50\x4f\x85\x41\x28\xdf\x39\x7e\x78\xc7\x1d\x5e\xc0\x82\x37\x96\x8b\x3d\x17\x3c\x93\x09\xb7\xa0\xcf\xc8\x77\x2f\x78\xdd\xf3\xdd\x04\x72\x44\xe2\x93\x62\xcd\xea\x40\x4a\x92\x5a\x96\xe6\x9f\x3a\x47\x92\xc8\x4e\x12\x92\xb4\x32\x94\xbe\xe7\xdc\x58\x8a\x6e\x96\x45\xdd\x67\xc1\xa7\x7d\xf3\xe7\x92\x98\x8c\x1c\x15\x7e\x36\xa3\xf6\xae\xb6\x44\x86\x1f\x2b\xac\xba\xe6\x7f\x9a\x45\x89\x2a\xb9\xac\x65\x36\xbd\xd1\x8e\xf6\x61\x7d\xcb\x84\x2b\xda\x77\x04\xff\x5c\x93\xa8\xae\x49\x16\xa1\xff\xa2\x25\xbd\x67\xa1\x7c\xcd\x4b\xe5\xc0\x1b\x79\xdc\x0e\x87\x9c\xbf\x37\x4a\xc1\x1a\x52\x37\x97\x94\x0f\xb5\xd4\x82\x8a\x1b\x89\xce\x02\x77\xf6\x5c\x22\x13\x60\xcf\xee\x15\xdd\x92\x34\xb6\xda\x43\x84\xc0\x5b\xbe\x3c\x78\x08\x1c\x75\xcf\x37\xde\x7a\x50\x70\xc6\xe3\x28\xe5\xda\xfc\x77\xf2\x27\x5c\x5d\x0e\x72\x79\x0b\x1a\x9f\x37\xcb\x30\x14\x34\x6d\x2f\xe3\x4c\xdf\x31\x1e\x42\x68\xf3\xdf\x42\xbe\x07\x1d\x95\xce\xce\x9d\xb5\x4e\xd2\xa6\xef\x7f\xb6\x8d\x5b\x09\xbf\xfb\x2f\x93\xc9\x15\x5c\xbe\xe5\x6d\x33\x92\xf9\xc5\xf8\x53\xc3\x62\x13\x7f\x05\x00\x00\xff\xff\xf7\xc1\x2a\x0f\xc4\x04\x00\x00")

func resourcesPodTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesPodTemplate,
		"resources/pod.template",
	)
}

func resourcesPodTemplate() (*asset, error) {
	bytes, err := resourcesPodTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/pod.template", size: 1220, mode: os.FileMode(436), modTime: time.Unix(1625668177, 0)}
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

	info := bindataFileInfo{name: "resources/srv-cluster-role-binding.json", size: 446, mode: os.FileMode(436), modTime: time.Unix(1615882822, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesSrvClusterRoleJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xdc\x50\x3d\x4f\xc3\x30\x10\xdd\xfb\x2b\x2c\x8f\x28\x34\x62\x43\x59\x19\xd8\x19\x58\x10\xc3\xc5\x39\x8a\xd5\xc4\x67\xdd\x9d\x23\x04\xca\x7f\x47\x69\x48\xaa\xa0\xb8\x43\xd9\x18\xef\xde\x87\xde\x7b\x5f\x3b\x63\x8c\xb1\x10\xfd\x33\xb2\x78\x0a\xb6\x32\x96\x6b\x70\x7b\x48\xfa\x4e\xec\x3f\x41\x3d\x85\xfd\xf1\x5e\xf6\x9e\xca\xfe\xce\x16\x93\xe2\xe8\x43\x33\x72\x1f\xda\x24\x8a\xfc\x44\x2d\xce\x50\x87\x0a\x0d\x28\xd8\xca\x4c\xf6\xa7\x6f\x80\x0e\x47\x01\x06\xc1\xae\x6e\xf1\x96\x22\x32\x28\xb1\x3d\x71\x86\x1f\x31\xa7\x16\xc5\x56\xe6\x65\x51\x9e\x3d\xe6\xa8\x8f\x4c\x29\xae\x49\x0b\x3c\xdb\x93\x8c\x81\xed\x8a\xf0\x5a\xac\xad\x18\x85\x12\x3b\xcc\x58\xb9\xa9\x9a\xd8\x22\x8f\x95\xa2\xa0\x69\x93\x72\x76\xbf\x04\xce\x06\x17\x83\xf6\xc8\x75\x26\xe4\xcd\x2f\xe5\x72\x0d\xc5\x95\x0b\xfe\x69\xb3\x48\xcd\x66\xdf\xf1\x5f\xe2\x07\xba\x2d\x50\x90\x7b\x9f\x19\x0a\x43\x13\xc9\x07\xdd\x06\x7b\xcc\x20\x8e\xc2\x9b\x3f\x74\x10\xff\xcd\xb2\x82\x8e\x51\xaf\xaf\x73\x40\xcd\x16\xda\x4d\xf7\xf0\x1d\x00\x00\xff\xff\x8a\xfc\x1c\xde\x0e\x04\x00\x00")

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

	info := bindataFileInfo{name: "resources/srv-cluster-role.json", size: 1038, mode: os.FileMode(436), modTime: time.Unix(1623687138, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesSrvDeploymentJsonTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\x4d\x6f\xdb\x30\x0c\xbd\xf7\x57\x18\x3e\x37\x8d\x73\xd8\x25\xb7\x60\xc5\x80\x01\x6b\x17\x14\xe8\x2e\x43\x0f\x8c\xcc\xa5\xc2\xa8\x8f\x49\x54\x36\x23\xc8\x7f\x1f\x64\xe7\xcb\xb3\xec\x38\xc1\x72\x4a\xc4\x47\xf2\x3d\x3e\x53\xce\xf6\x2e\xcb\xb2\x2c\x07\x2b\xbf\xa1\xf3\xd2\xe8\x7c\x1e\x7f\x59\x3f\xdd\xcc\xf2\xfb\x26\xf8\x53\xea\x32\x1e\x3f\xa2\x25\x53\x29\xd4\x7c\x88\x28\x64\x28\x81\x21\x9f\x67\x4d\xa1\xfa\x94\x60\x85\xe4\x5b\x67\xfb\x26\x36\x96\x41\xed\x51\xad\x08\x27\xc6\xa2\x03\x36\x2e\x3f\xc2\x76\xf7\xa7\x2a\x1a\x14\x0e\xc0\xf7\xd0\xdc\x5b\x14\xed\xf6\x0e\x2d\x49\x01\x91\xc0\xec\xac\x9e\x47\x42\x11\xd3\x3b\xbc\x14\xb0\x78\xff\x92\x26\x3d\x92\x78\xcd\x28\x29\xc3\xb3\x03\xc6\x75\xd5\x6d\xcb\x95\xad\x05\xbe\x18\x22\xa9\xd7\xaf\xb6\x04\xc6\xf4\x2c\x18\x95\xa5\x18\xed\x72\x4f\x39\x70\xc9\x89\x2b\x85\xb5\xc5\xfd\xc3\x2d\x4b\x9a\x70\x8a\xa0\xdb\x48\x81\x0b\x21\x4c\xd0\xfc\xdc\xeb\xe9\x7d\x37\x55\x18\xcd\x20\x35\xba\x28\xe0\x7b\x52\x40\x5a\x56\x9d\x2d\x15\xac\xeb\x5e\xdb\xed\xc3\xe7\xf8\x7d\xb7\x4b\x34\x69\xc3\x97\x81\x68\x69\x48\x8a\xea\x3c\xf1\x74\x3a\x5c\xe2\xf0\xc0\xc6\xa1\x0e\xc0\x3c\x8a\xe0\x24\x57\x1f\x8d\x66\xfc\xc3\xbd\xe6\x1c\x13\x1c\x42\xf9\x55\x53\xf5\x62\x0c\x7f\x92\x84\xbe\xf2\x8c\x2a\x9f\x67\xec\x02\xf6\x37\x6a\x72\x83\x5e\xf8\x67\xa3\x63\xee\x15\x19\xaf\x1e\xe3\x9e\xcc\x8a\xa2\xe8\x45\xef\x06\x34\x3a\xf4\x26\x38\x81\xfd\x8f\xde\x11\x4a\x52\x49\xbe\x8c\xab\xb1\xc2\x86\x38\xe0\x59\x51\xa8\x81\x09\x1f\xe1\x0a\x95\x71\xb5\x95\x1f\x8a\x27\xd9\x7d\xa6\xcf\x3f\x03\x6a\xb2\x46\xd1\xaf\x80\xfe\x6a\xa2\xff\x9f\xe7\x4d\x7e\x80\x5b\xf7\x2f\xd1\x11\x15\x77\x15\x53\xab\xd8\x42\x4d\x26\x2b\x43\x5c\xae\x2e\xe1\xa6\xc1\xbb\x29\x19\x01\x34\x3d\xac\xfb\xd4\x33\x30\x3e\x94\xab\x7e\x8d\x6f\x03\x2a\x36\x86\x82\xc2\xa7\x78\x91\x5c\x56\x33\xc2\xa5\xc3\xc2\xd6\xac\x26\x4d\xf5\x51\x7e\x45\x06\x4b\xe0\xf7\x98\x9c\x90\x39\xa6\xc6\x61\xa9\xf3\x79\xf6\x03\xc8\xe3\x8d\xa6\xbf\x25\x23\x5d\x7c\x62\xac\xfb\x71\xde\x74\xb9\x5e\x39\xb9\x1c\x95\xe5\xea\x51\xd6\xef\xde\xb4\x98\x71\x94\x1d\x7a\x06\xc7\xa7\x3b\x7a\x41\xbf\xa1\xf2\xbd\xef\xe1\xe6\x9f\xc2\xdd\xee\x6f\x00\x00\x00\xff\xff\x9d\xf4\x96\xc0\xe6\x08\x00\x00")

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

	info := bindataFileInfo{name: "resources/srv-deployment.json.template", size: 2278, mode: os.FileMode(436), modTime: time.Unix(1623687138, 0)}
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

	info := bindataFileInfo{name: "resources/srv-service-account.json", size: 117, mode: os.FileMode(436), modTime: time.Unix(1615882842, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesVolumeClaimTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\x8f\x3d\x6b\x04\x21\x10\x86\xfb\xfd\x15\x32\xf5\x12\x48\xbb\x6d\xea\x6c\x42\x02\x9b\x22\x5c\x31\xe8\x70\xc8\xad\xba\xe7\xb8\xd7\x88\xff\xfd\xd0\xfd\x94\xb3\x92\xf7\x99\xe7\x65\x26\x36\x42\x08\x01\x38\xe9\x81\x3c\x6b\x67\xa1\x13\xf0\x78\x87\x76\xc9\x6f\xda\xaa\x9c\x7c\x67\xc8\x81\x6c\x18\xdc\x38\x1b\xfa\x18\x51\x9b\x6d\xc8\x50\x40\x85\x01\xa1\x13\x4b\x5d\x49\x2d\x1a\xca\x6a\x8c\x6f\x3d\x1a\x4a\x09\x0a\x4b\xab\xc4\x13\xc9\x5a\xe0\xe0\x3c\x5e\x73\x35\x73\x7f\xc8\xbf\x4b\xbc\x76\xb4\xc7\x3c\x4a\x49\xcc\x9f\x4e\x11\x43\x27\xfe\x77\x50\xe0\x0f\xa1\xfa\xf3\x3a\xd0\x97\x95\x04\x3b\xbb\x9c\x7c\x4f\xec\x66\x2f\x8b\x1d\x6b\xdb\xd3\x7d\x26\x0e\xaf\xe4\xbc\x68\xbd\xdf\x76\xdf\xf6\x52\x53\xff\x52\x93\x9e\x01\x00\x00\xff\xff\x81\xd7\x44\xe2\x6c\x01\x00\x00")

func resourcesVolumeClaimTemplateBytes() ([]byte, error) {
	return bindataRead(
		_resourcesVolumeClaimTemplate,
		"resources/volume-claim.template",
	)
}

func resourcesVolumeClaimTemplate() (*asset, error) {
	bytes, err := resourcesVolumeClaimTemplateBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/volume-claim.template", size: 364, mode: os.FileMode(436), modTime: time.Unix(1625668177, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
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
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
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
	"resources/config-map.template":           resourcesConfigMapTemplate,
	"resources/crd-cluster.json":              resourcesCrdClusterJson,
	"resources/crd-resource.json":             resourcesCrdResourceJson,
	"resources/crd.template":                  resourcesCrdTemplate,
	"resources/generic.template":              resourcesGenericTemplate,
	"resources/headless-service.template":     resourcesHeadlessServiceTemplate,
	"resources/kind-cluster.template":         resourcesKindClusterTemplate,
	"resources/kind-resource.template":        resourcesKindResourceTemplate,
	"resources/mock-crd.template":             resourcesMockCrdTemplate,
	"resources/mock.template":                 resourcesMockTemplate,
	"resources/pod.template":                  resourcesPodTemplate,
	"resources/srv-cluster-role-binding.json": resourcesSrvClusterRoleBindingJson,
	"resources/srv-cluster-role.json":         resourcesSrvClusterRoleJson,
	"resources/srv-deployment.json.template":  resourcesSrvDeploymentJsonTemplate,
	"resources/srv-service-account.json":      resourcesSrvServiceAccountJson,
	"resources/volume-claim.template":         resourcesVolumeClaimTemplate,
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
// AssetDir("foo.txt") and AssetDir("nonexistent") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
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
		"config-map.template":           &bintree{resourcesConfigMapTemplate, map[string]*bintree{}},
		"crd-cluster.json":              &bintree{resourcesCrdClusterJson, map[string]*bintree{}},
		"crd-resource.json":             &bintree{resourcesCrdResourceJson, map[string]*bintree{}},
		"crd.template":                  &bintree{resourcesCrdTemplate, map[string]*bintree{}},
		"generic.template":              &bintree{resourcesGenericTemplate, map[string]*bintree{}},
		"headless-service.template":     &bintree{resourcesHeadlessServiceTemplate, map[string]*bintree{}},
		"kind-cluster.template":         &bintree{resourcesKindClusterTemplate, map[string]*bintree{}},
		"kind-resource.template":        &bintree{resourcesKindResourceTemplate, map[string]*bintree{}},
		"mock-crd.template":             &bintree{resourcesMockCrdTemplate, map[string]*bintree{}},
		"mock.template":                 &bintree{resourcesMockTemplate, map[string]*bintree{}},
		"pod.template":                  &bintree{resourcesPodTemplate, map[string]*bintree{}},
		"srv-cluster-role-binding.json": &bintree{resourcesSrvClusterRoleBindingJson, map[string]*bintree{}},
		"srv-cluster-role.json":         &bintree{resourcesSrvClusterRoleJson, map[string]*bintree{}},
		"srv-deployment.json.template":  &bintree{resourcesSrvDeploymentJsonTemplate, map[string]*bintree{}},
		"srv-service-account.json":      &bintree{resourcesSrvServiceAccountJson, map[string]*bintree{}},
		"volume-claim.template":         &bintree{resourcesVolumeClaimTemplate, map[string]*bintree{}},
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
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
