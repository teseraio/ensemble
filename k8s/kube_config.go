package k8s

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// KubeConfig is the kubeconfig to connect with the K8s apiserver
type KubeConfig struct {
	Host     string
	CertFile string
	KeyFile  string
	CAFile   string
}

func (k *KubeConfig) ToConfig() (*Config, error) {
	tlsConfig, err := k.GetTLSConfig()
	if err != nil {
		return nil, err
	}
	c := &Config{
		Host:      k.Host,
		TLSConfig: tlsConfig,
	}
	return c, nil
}

// GetTLSConfig returns the tls.Config for this kubeconfig
func (k *KubeConfig) GetTLSConfig() (*tls.Config, error) {
	cert, err := tls.X509KeyPair([]byte(k.CertFile), []byte(k.KeyFile))
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(k.CAFile))

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}
	return tlsConf, nil
}

// GetDefaultKubeConfigPath returns the default location for the kube path
func GetDefaultKubeConfigPath() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".kube", "config"), nil
}

// NewKubeConfig loads the kube config from a given path
func NewKubeConfig(path string, context string) (*KubeConfig, error) {
	var err error
	if path == "" {
		if path, err = GetDefaultKubeConfigPath(); err != nil {
			return nil, err
		}
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var k8sConfig k8sConfig
	if err := yaml.Unmarshal(content, &k8sConfig); err != nil {
		return nil, err
	}

	if context == "" {
		if k8sConfig.CurrentContext == "" {
			return nil, fmt.Errorf("there is no default context")
		}
		context = k8sConfig.CurrentContext
	}

	// find the specific context
	var clusterContext, userContext string

	for _, c := range k8sConfig.Contexts {
		if c.Name == k8sConfig.CurrentContext {
			clusterContext = c.Context.Cluster
			userContext = c.Context.User
		}
	}

	// decode the specific context data
	var cluster *k8sCluster
	for _, k := range k8sConfig.Clusters {
		if k.Name == clusterContext {
			cluster = k
		}
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster info not found for context %s", clusterContext)
	}

	var user *k8sUser
	for _, k := range k8sConfig.Users {
		if k.Name == userContext {
			user = k
		}
	}
	if user == nil {
		return nil, fmt.Errorf("auth info not found for context %s", userContext)
	}

	caFile, err := loadTLSCert(cluster.Cluster.CertificateAuthorityData, cluster.Cluster.CertificateAuthority)
	if err != nil {
		return nil, err
	}
	certFile, err := loadTLSCert(user.User.ClientCertificateData, user.User.ClientCertificate)
	if err != nil {
		return nil, err
	}
	keyFile, err := loadTLSCert(user.User.ClientKeyData, user.User.ClientKey)
	if err != nil {
		return nil, err
	}

	config := &KubeConfig{
		Host:     cluster.Cluster.Server,
		CAFile:   caFile,
		CertFile: certFile,
		KeyFile:  keyFile,
	}
	return config, nil
}

func loadTLSCert(data, file string) (string, error) {
	if data == "" && file == "" {
		return "", fmt.Errorf("either one of those has to be true")
	}
	if data != "" {
		// data in base64 format
		buf, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	// data from file
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

type k8sConfig struct {
	Kind           string        `yaml:"kind,omitempty"`
	Clusters       []*k8sCluster `yaml:"clusters"`
	Users          []*k8sUser    `yaml:"users"`
	Contexts       []*k8sContext `yaml:"contexts"`
	CurrentContext string        `yaml:"current-context"`
}

type k8sClusterDetail struct {
	Server                   string `yaml:"server"`
	CertificateAuthority     string `yaml:"certificate-authority,omitempty"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
}

type k8sCluster struct {
	Name    string
	Cluster *k8sClusterDetail
}

type k8sUserDetail struct {
	ClientCertificate     string `yaml:"client-certificate"`
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKey             string `yaml:"client-key"`
	ClientKeyData         string `yaml:"client-key-data"`
}

type k8sUser struct {
	Name string
	User *k8sUserDetail
}

type k8sContextDetail struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type k8sContext struct {
	Name    string
	Context *k8sContextDetail
}
