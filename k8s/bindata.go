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

var _resourcesMockTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\x8e\x41\xca\xc3\x20\x10\x46\xd7\x7a\x8a\x61\xd6\x21\x3f\xff\x36\x07\xe9\x7e\x1a\xa7\x20\xa9\x1a\x34\xc9\x66\x98\xbb\x17\x1b\x6b\x69\x5c\x89\xef\xf9\xf8\xc4\x02\x00\x20\xad\xfe\xc6\xb9\xf8\x14\x71\x02\x0c\x69\x5e\x46\x9f\xfe\x8e\x7f\x1c\x4e\xbe\xf8\xe8\x2a\x11\x19\x33\x97\xb4\xe7\x99\x55\x3f\x30\xf0\x46\x8e\x36\xc2\x09\xce\x5c\x3d\x22\xfe\x01\x5d\x6e\x71\x55\x6b\x0c\x5e\x1e\x2f\xdd\xae\xe2\x60\x8d\x11\xe1\xe8\x54\x7b\x16\x23\x05\x6e\x3f\xea\x55\x15\xdf\x4c\xdb\x96\xb2\xf2\xfc\xb3\x03\xef\xcd\x3e\xe8\xb9\x7f\x75\xab\xaf\x00\x00\x00\xff\xff\x02\x19\xd8\x9c\xfa\x00\x00\x00")

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

	info := bindataFileInfo{name: "resources/mock.template", size: 250, mode: os.FileMode(436), modTime: time.Unix(1627312510, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesPodTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x54\x4d\x8b\xdb\x30\x10\x3d\xcb\xbf\x42\x0c\x39\xa6\x81\x5e\x03\x3d\x2c\xdd\x85\xf6\xd0\x6d\xe8\x61\x29\x94\x3d\x4c\xe2\x69\xd6\xac\x3e\x8c\x24\xbb\x04\xa1\xff\x5e\x46\x96\x6c\xb7\xa4\xa1\xec\x29\x1e\xbd\xa7\x19\xbd\x37\x33\x89\x8d\x00\xec\xbb\x27\x72\xbe\xb3\x06\xf6\x12\xc6\xf7\xb0\x6d\x04\xbc\x76\xa6\xe5\xf0\x60\xdb\x1c\x6b\x0a\xd8\x62\x40\xd8\xcb\xd8\x08\x01\x06\x35\x31\x1e\xe3\xee\xf3\x7d\x4a\xcc\x11\xa0\xf0\x48\xca\x17\x8a\x00\x32\x9e\xf4\x51\x55\xde\x43\x09\x0b\x5b\x40\x4b\xbd\xb2\x17\x4d\x26\x14\xc6\xfd\x7c\x90\x12\x34\x42\xa4\x46\x24\x2e\xee\x7b\x3a\xd5\xc2\x2f\xd6\x87\x55\xf1\x4f\x25\xac\x4f\xf0\xc3\xb1\xb5\x1a\x3b\x73\xbd\x28\x38\xf2\x01\x5d\x38\x58\xd5\x9d\x2e\xcc\x79\xa4\x91\xdc\x84\x9d\xac\x09\xd8\x19\x72\x2c\xe1\x07\x3f\x31\xeb\x10\x31\x76\x3f\xe5\xee\xc1\x8c\x29\xe5\x18\xc8\x8c\x95\xc1\xa8\xdc\x78\xb9\xff\x20\x3d\xf5\xe8\x30\x58\x27\x61\x2b\x41\x16\x32\xe3\x0e\xcd\x99\xe4\xe6\x95\x2e\x5b\xb9\x19\x51\x0d\xc4\x7c\xce\xb8\x66\x9d\x50\x29\x4e\x35\x1f\x4d\x3f\x6b\xaf\x39\x45\xb5\x8f\x91\x9c\xab\x40\xf9\x7b\xf2\x4d\x64\xeb\xa6\xb4\x64\xda\x92\xf0\x79\xdb\xfc\x7d\x04\xa3\x55\x83\xa6\x2f\x76\x30\xc1\xaf\x25\xb1\xe0\xa7\x15\xb6\x7a\xe6\x7f\x8a\x25\x45\x3a\xab\x5c\xa7\xb9\xa9\x8d\x6f\xec\x1e\x97\x5e\x66\x5c\xf3\xbd\x03\x86\x97\x35\x89\xe3\x35\xc9\x11\xb6\x5f\x8d\xe2\x7e\x56\xca\xb7\x72\x54\x0b\x5e\xf1\xe3\xba\x39\xac\xfc\xa3\xd5\x1a\x17\x93\x4e\x53\xc8\xfe\xf0\x48\xcd\x28\x5c\x71\x74\x4a\x70\xe7\xce\xd5\x32\x40\x77\xf6\x6f\x98\x96\x9c\xe3\xd6\x78\x40\x8c\x72\x27\xe7\x86\xc7\x28\xc9\xb4\xf2\x46\xaf\x3b\x8d\x67\x3a\x0c\x4a\x2d\xc3\x7f\xa7\x7e\xe1\xc5\x17\x23\xe7\x5e\xf0\xfa\xbc\x9b\x97\xa1\xa2\xf9\x7a\x5d\x79\xfe\x4e\x69\x1f\xe3\xae\xfc\x75\x94\x77\x70\xa9\x5c\xbb\x4c\xd6\xb2\x49\x37\x75\xff\x73\x6c\xfc\x42\xf8\x53\x7f\xdd\x4c\xa9\xb1\xff\x5e\xae\x4d\x48\xe1\x57\xe1\xcf\x8d\x48\x4d\xfa\x1d\x00\x00\xff\xff\x14\x5d\x6b\x73\xe8\x04\x00\x00")

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

	info := bindataFileInfo{name: "resources/pod.template", size: 1256, mode: os.FileMode(436), modTime: time.Unix(1627397902, 0)}
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

var _resourcesSrvClusterRoleJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x53\xbb\x4e\xc4\x30\x10\xec\xef\x2b\x2c\x97\x28\x24\xa2\x43\x69\x29\xe8\x29\x68\x10\xc5\xc6\x59\x0e\xeb\x12\xaf\xb5\x6b\x47\x08\x94\x7f\x47\xb9\x3c\x4e\x41\xf1\x15\xa4\xb9\xd2\x3b\x0f\xcd\x8c\xe4\x9f\x83\x52\x4a\x69\xf0\xf6\x15\x59\x2c\x39\x5d\x2a\xcd\x15\x98\x1c\x62\xf8\x24\xb6\xdf\x10\x2c\xb9\xfc\xf4\x28\xb9\xa5\xa2\x7b\xd0\xd9\xa8\x38\x59\x57\x0f\xdc\xa7\x26\x4a\x40\x7e\xa1\x06\x67\xa8\xc5\x00\x35\x04\xd0\xa5\x1a\xed\xcf\x57\x07\x2d\x0e\x02\x74\x82\x6d\xd5\xe0\x3d\x79\x64\x08\xc4\xfa\xcc\xe9\x27\x31\xc7\x06\x45\x97\xea\x6d\x51\x5e\x3c\xe6\xa8\xcf\x4c\xd1\xaf\x49\x0b\x3c\xdb\x93\x0c\x81\xf5\x8a\xf0\x9e\xad\xad\x18\x85\x22\x1b\x4c\x58\x99\xb1\x9a\xe8\x2c\x8d\x15\x12\x20\xc4\x4d\xca\xc5\xfd\x1a\x38\x1b\x5c\x0d\xda\x21\x57\x89\x90\x77\x7f\x94\xcb\xab\xcf\xfe\xb9\xe0\xae\xcd\x3c\xd5\x9b\x7d\x87\x7b\x81\x5f\x68\xb6\x40\x41\xee\x6c\x62\x28\x74\xb5\x27\xeb\xc2\x36\xd8\x61\x02\x31\xe4\x3e\xec\xb1\x05\x7f\x4b\xcb\x8e\x71\xa7\xbf\xb4\x6b\xe6\xa9\xf8\xed\x54\xdb\xd5\x46\xd0\x30\xee\xa8\x73\xc4\x90\x2c\x74\x18\xdf\xfd\x6f\x00\x00\x00\xff\xff\xc9\x69\x69\xba\xe9\x04\x00\x00")

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

	info := bindataFileInfo{name: "resources/srv-cluster-role.json", size: 1257, mode: os.FileMode(436), modTime: time.Unix(1627312510, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesSrvDeploymentJsonTemplate = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x56\x4d\x8f\xda\x30\x10\xbd\xef\xaf\x88\x72\x26\x10\x0e\xbd\x70\x43\x5d\x55\xaa\xd4\xdd\xa2\x95\xb6\x97\x6a\x0f\x83\x99\xb2\x56\xc7\x1f\xb5\x27\xb4\x11\xe2\xbf\x57\x76\xf8\x6c\x9c\x04\x50\x77\x2f\x60\xbf\x99\x79\xef\x8d\xc7\x66\xfb\x90\x65\x59\x96\x83\x95\xdf\xd0\x79\x69\x74\x3e\x0b\xdf\xac\x9f\x6c\xa6\xf9\xa8\xd9\xfc\x29\xf5\x2a\x2c\x3f\xa2\x25\x53\x2b\xd4\x7c\xd8\x51\xc8\xb0\x02\x86\x7c\x96\x35\x89\xe2\x2a\xc1\x12\xc9\x5f\xac\xed\x8b\xd8\x90\x06\xb5\x47\xb5\x24\x2c\x8c\x45\x07\x6c\x5c\x7e\x84\xed\x46\xa7\x2c\x1a\x14\xf6\xc0\xf7\xd0\xdc\x5b\x14\x97\xe5\x1d\x5a\x92\x02\x02\x81\xe9\x59\x3e\x8f\x84\x22\x84\xb7\x78\x29\x60\xf1\xfe\x25\x4d\xfa\x4a\xe2\x91\x51\x52\x86\x67\x07\x8c\xeb\xba\x5d\x96\x6b\x1b\x05\xbe\x18\x22\xa9\xd7\xaf\x76\x05\x8c\x69\x2f\x18\x95\xa5\xb0\xdb\xe6\x9e\xea\xc0\x50\x27\x6e\x14\x76\x29\xee\x1f\x6e\x59\xb2\x09\xa7\x1d\x74\x1b\x29\x70\x2e\x84\xa9\x34\x3f\x77\xf6\x74\xd4\x0e\x15\x46\x33\x48\x8d\x2e\x08\xf8\x9e\x14\x90\x96\x15\xa3\xa5\x82\x75\xac\xb5\xdd\x8e\x3f\x87\xcf\xbb\x5d\xa2\xc8\x25\x7c\x51\x11\x2d\x0c\x49\x51\x9f\x07\x9e\x56\xfb\x53\x1c\x0e\x6c\x30\xb5\x07\xe6\x51\x54\x4e\x72\xfd\xd1\x68\xc6\x3f\xdc\xd9\x9c\x63\x80\x43\x58\x7d\xd5\x54\xbf\x18\xc3\x9f\x24\xa1\xaf\x3d\xa3\xca\x67\x19\xbb\x0a\xbb\x0b\x35\xb1\x95\x9e\xfb\x67\xa3\x43\xec\x0d\x11\xaf\x1e\xc3\x9c\x4c\xcb\xb2\xec\x44\xef\x7a\x34\x3a\xf4\xa6\x72\x02\xbb\x8f\xde\x11\x4a\x52\x49\x1e\xc6\x45\xac\xb0\x55\x30\x78\x5a\x96\xaa\xc7\xe1\x23\x5c\xa1\x32\x2e\xb6\xf2\x43\xf9\x24\xdb\x67\xfa\xfc\xaf\x47\x4d\xd6\x28\xfa\x55\xa1\xbf\x99\xe8\xff\xe7\x79\x57\x3f\xc0\xad\xbb\x87\xe8\x88\x0a\xb3\x8a\xa9\x51\xbc\x40\x15\xc5\xd2\x10\xaf\x96\x43\xb8\x49\xe5\xdd\x84\x8c\x00\x9a\x1c\xc6\x7d\xe2\x19\x18\xc7\xc3\xb1\x45\xb1\x0c\x6f\xce\x00\xaa\x1c\xc7\xff\x6e\xc3\xde\x7a\x2c\xd9\x18\xaa\x14\x3e\x85\x5b\x69\xd8\x9a\x2b\x5a\x7e\x98\xfe\x28\xb1\x68\xb2\x5f\xd5\xfc\xc0\x60\x01\xfc\x1e\x82\x13\x9e\x5d\x93\xe3\x70\x43\xe4\xb3\xec\x07\x90\xc7\x3b\x4f\xd0\x5b\x72\xa7\x8d\x4f\xd8\xba\xb7\xf3\xae\x9b\xfa\x46\xe7\x72\x54\x96\xeb\x47\x19\x1f\xf2\xb4\x98\xeb\x28\x3b\xf4\x0c\x8e\x4f\x17\xfe\x9c\x7e\x43\xed\x3b\x1f\xf5\xe6\x67\xc7\xc3\xee\x6f\x00\x00\x00\xff\xff\x1f\x1f\xf9\x0c\x33\x09\x00\x00")

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

	info := bindataFileInfo{name: "resources/srv-deployment.json.template", size: 2355, mode: os.FileMode(436), modTime: time.Unix(1627373334, 0)}
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
