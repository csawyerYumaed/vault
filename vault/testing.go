package vault

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	mathrand "math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	log "github.com/mgutz/logxi/v1"
	"github.com/mitchellh/copystructure"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/http2"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/audit"
	"github.com/hashicorp/vault/helper/jsonutil"
	"github.com/hashicorp/vault/helper/logformat"
	"github.com/hashicorp/vault/helper/reload"
	"github.com/hashicorp/vault/helper/salt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/hashicorp/vault/physical"
)

// This file contains a number of methods that are useful for unit
// tests within other packages.

const (
	testSharedPublicKey = `
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC9i+hFxZHGo6KblVme4zrAcJstR6I0PTJozW286X4WyvPnkMYDQ5mnhEYC7UWCvjoTWbPEXPX7NjhRtwQTGD67bV+lrxgfyzK1JZbUXK4PwgKJvQD+XyyWYMzDgGSQY61KUSqCxymSm/9NZkPU3ElaQ9xQuTzPpztM4ROfb8f2Yv6/ZESZsTo0MTAkp8Pcy+WkioI/uJ1H7zqs0EA4OMY4aDJRu0UtP4rTVeYNEAuRXdX+eH4aW3KMvhzpFTjMbaJHJXlEeUm2SaX5TNQyTOvghCeQILfYIL/Ca2ij8iwCmulwdV6eQGfd4VDu40PvSnmfoaE38o6HaPnX0kUcnKiT
`
	testSharedPrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAvYvoRcWRxqOim5VZnuM6wHCbLUeiND0yaM1tvOl+Fsrz55DG
A0OZp4RGAu1Fgr46E1mzxFz1+zY4UbcEExg+u21fpa8YH8sytSWW1FyuD8ICib0A
/l8slmDMw4BkkGOtSlEqgscpkpv/TWZD1NxJWkPcULk8z6c7TOETn2/H9mL+v2RE
mbE6NDEwJKfD3MvlpIqCP7idR+86rNBAODjGOGgyUbtFLT+K01XmDRALkV3V/nh+
GltyjL4c6RU4zG2iRyV5RHlJtkml+UzUMkzr4IQnkCC32CC/wmtoo/IsAprpcHVe
nkBn3eFQ7uND70p5n6GhN/KOh2j519JFHJyokwIDAQABAoIBAHX7VOvBC3kCN9/x
+aPdup84OE7Z7MvpX6w+WlUhXVugnmsAAVDczhKoUc/WktLLx2huCGhsmKvyVuH+
MioUiE+vx75gm3qGx5xbtmOfALVMRLopjCnJYf6EaFA0ZeQ+NwowNW7Lu0PHmAU8
Z3JiX8IwxTz14DU82buDyewO7v+cEr97AnERe3PUcSTDoUXNaoNxjNpEJkKREY6h
4hAY676RT/GsRcQ8tqe/rnCqPHNd7JGqL+207FK4tJw7daoBjQyijWuB7K5chSal
oPInylM6b13ASXuOAOT/2uSUBWmFVCZPDCmnZxy2SdnJGbsJAMl7Ma3MUlaGvVI+
Tfh1aQkCgYEA4JlNOabTb3z42wz6mz+Nz3JRwbawD+PJXOk5JsSnV7DtPtfgkK9y
6FTQdhnozGWShAvJvc+C4QAihs9AlHXoaBY5bEU7R/8UK/pSqwzam+MmxmhVDV7G
IMQPV0FteoXTaJSikhZ88mETTegI2mik+zleBpVxvfdhE5TR+lq8Br0CgYEA2AwJ
CUD5CYUSj09PluR0HHqamWOrJkKPFPwa+5eiTTCzfBBxImYZh7nXnWuoviXC0sg2
AuvCW+uZ48ygv/D8gcz3j1JfbErKZJuV+TotK9rRtNIF5Ub7qysP7UjyI7zCssVM
kuDd9LfRXaB/qGAHNkcDA8NxmHW3gpln4CFdSY8CgYANs4xwfercHEWaJ1qKagAe
rZyrMpffAEhicJ/Z65lB0jtG4CiE6w8ZeUMWUVJQVcnwYD+4YpZbX4S7sJ0B8Ydy
AhkSr86D/92dKTIt2STk6aCN7gNyQ1vW198PtaAWH1/cO2UHgHOy3ZUt5X/Uwxl9
cex4flln+1Viumts2GgsCQKBgCJH7psgSyPekK5auFdKEr5+Gc/jB8I/Z3K9+g4X
5nH3G1PBTCJYLw7hRzw8W/8oALzvddqKzEFHphiGXK94Lqjt/A4q1OdbCrhiE68D
My21P/dAKB1UYRSs9Y8CNyHCjuZM9jSMJ8vv6vG/SOJPsnVDWVAckAbQDvlTHC9t
O98zAoGAcbW6uFDkrv0XMCpB9Su3KaNXOR0wzag+WIFQRXCcoTvxVi9iYfUReQPi
oOyBJU/HMVvBfv4g+OVFLVgSwwm6owwsouZ0+D/LasbuHqYyqYqdyPJQYzWA2Y+F
+B6f4RoPdSXj24JHPg/ioRxjaj094UXJxua2yfkcecGNEuBQHSs=
-----END RSA PRIVATE KEY-----
`
)

// TestCore returns a pure in-memory, uninitialized core for testing.
func TestCore(t testing.TB) *Core {
	return TestCoreWithSeal(t, nil)
}

// TestCoreNewSeal returns an in-memory, ininitialized core with the new seal
// configuration.
func TestCoreNewSeal(t testing.TB) *Core {
	return TestCoreWithSeal(t, &TestSeal{})
}

// TestCoreWithSeal returns a pure in-memory, uninitialized core with the
// specified seal for testing.
func TestCoreWithSeal(t testing.TB, testSeal Seal) *Core {
	logger := logformat.NewVaultLogger(log.LevelTrace)
	physicalBackend := physical.NewInmem(logger)

	conf := testCoreConfig(t, physicalBackend, logger)

	if testSeal != nil {
		conf.Seal = testSeal
	}

	c, err := NewCore(conf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return c
}

func testCoreConfig(t testing.TB, physicalBackend physical.Backend, logger log.Logger) *CoreConfig {
	noopAudits := map[string]audit.Factory{
		"noop": func(config *audit.BackendConfig) (audit.Backend, error) {
			view := &logical.InmemStorage{}
			view.Put(&logical.StorageEntry{
				Key:   "salt",
				Value: []byte("foo"),
			})
			config.SaltConfig = &salt.Config{
				HMAC:     sha256.New,
				HMACType: "hmac-sha256",
			}
			config.SaltView = view
			return &noopAudit{
				Config: config,
			}, nil
		},
	}
	noopBackends := make(map[string]logical.Factory)
	noopBackends["noop"] = func(config *logical.BackendConfig) (logical.Backend, error) {
		b := new(framework.Backend)
		b.Setup(config)
		return b, nil
	}
	noopBackends["http"] = func(config *logical.BackendConfig) (logical.Backend, error) {
		return new(rawHTTP), nil
	}
	logicalBackends := make(map[string]logical.Factory)
	for backendName, backendFactory := range noopBackends {
		logicalBackends[backendName] = backendFactory
	}
	logicalBackends["generic"] = LeasedPassthroughBackendFactory
	for backendName, backendFactory := range testLogicalBackends {
		logicalBackends[backendName] = backendFactory
	}

	conf := &CoreConfig{
		Physical:           physicalBackend,
		AuditBackends:      noopAudits,
		LogicalBackends:    logicalBackends,
		CredentialBackends: noopBackends,
		DisableMlock:       true,
		Logger:             logger,
	}

	return conf
}

// TestCoreInit initializes the core with a single key, and returns
// the key that must be used to unseal the core and a root token.
func TestCoreInit(t testing.TB, core *Core) ([][]byte, string) {
	return TestCoreInitClusterWrapperSetup(t, core, nil, nil)
}

func TestCoreInitClusterWrapperSetup(t testing.TB, core *Core, clusterAddrs []*net.TCPAddr, handler http.Handler) ([][]byte, string) {
	core.SetClusterListenerAddrs(clusterAddrs)
	core.SetClusterHandler(handler)
	result, err := core.Initialize(&InitParams{
		BarrierConfig: &SealConfig{
			SecretShares:    3,
			SecretThreshold: 3,
		},
		RecoveryConfig: &SealConfig{
			SecretShares:    3,
			SecretThreshold: 3,
		},
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return result.SecretShares, result.RootToken
}

func TestCoreUnseal(core *Core, key []byte) (bool, error) {
	return core.Unseal(key)
}

// TestCoreUnsealed returns a pure in-memory core that is already
// initialized and unsealed.
func TestCoreUnsealed(t testing.TB) (*Core, [][]byte, string) {
	core := TestCore(t)
	keys, token := TestCoreInit(t, core)
	for _, key := range keys {
		if _, err := TestCoreUnseal(core, TestKeyCopy(key)); err != nil {
			t.Fatalf("unseal err: %s", err)
		}
	}

	sealed, err := core.Sealed()
	if err != nil {
		t.Fatalf("err checking seal status: %s", err)
	}
	if sealed {
		t.Fatal("should not be sealed")
	}

	return core, keys, token
}

func TestCoreUnsealedBackend(t testing.TB, backend physical.Backend) (*Core, [][]byte, string) {
	logger := logformat.NewVaultLogger(log.LevelTrace)
	conf := testCoreConfig(t, backend, logger)
	conf.Seal = &TestSeal{}

	core, err := NewCore(conf)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	keys, token := TestCoreInit(t, core)
	for _, key := range keys {
		if _, err := TestCoreUnseal(core, TestKeyCopy(key)); err != nil {
			t.Fatalf("unseal err: %s", err)
		}
	}

	sealed, err := core.Sealed()
	if err != nil {
		t.Fatalf("err checking seal status: %s", err)
	}
	if sealed {
		t.Fatal("should not be sealed")
	}

	return core, keys, token
}

func testTokenStore(t testing.TB, c *Core) *TokenStore {
	me := &MountEntry{
		Table:       credentialTableType,
		Path:        "token/",
		Type:        "token",
		Description: "token based credentials",
	}

	meUUID, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	me.UUID = meUUID

	view := NewBarrierView(c.barrier, credentialBarrierPrefix+me.UUID+"/")
	sysView := c.mountEntrySysView(me)

	tokenstore, _ := c.newCredentialBackend("token", sysView, view, nil)
	if err := tokenstore.Initialize(); err != nil {
		panic(err)
	}
	ts := tokenstore.(*TokenStore)

	router := NewRouter()
	err = router.Mount(ts, "auth/token/", &MountEntry{Table: credentialTableType, UUID: "authtokenuuid", Path: "auth/token", Accessor: "authtokenaccessor"}, ts.view)
	if err != nil {
		t.Fatal(err)
	}

	subview := c.systemBarrierView.SubView(expirationSubPath)
	logger := logformat.NewVaultLogger(log.LevelTrace)

	exp := NewExpirationManager(router, subview, ts, logger)
	ts.SetExpirationManager(exp)

	return ts
}

// TestCoreWithTokenStore returns an in-memory core that has a token store
// mounted, so that logical token functions can be used
func TestCoreWithTokenStore(t testing.TB) (*Core, *TokenStore, [][]byte, string) {
	c, keys, root := TestCoreUnsealed(t)
	ts := testTokenStore(t, c)

	return c, ts, keys, root
}

// TestCoreWithBackendTokenStore returns a core that has a token store
// mounted and used the provided physical backend, so that logical token
// functions can be used
func TestCoreWithBackendTokenStore(t testing.TB, backend physical.Backend) (*Core, *TokenStore, [][]byte, string) {
	c, keys, root := TestCoreUnsealedBackend(t, backend)
	ts := testTokenStore(t, c)

	return c, ts, keys, root
}

// TestKeyCopy is a silly little function to just copy the key so that
// it can be used with Unseal easily.
func TestKeyCopy(key []byte) []byte {
	result := make([]byte, len(key))
	copy(result, key)
	return result
}

func TestDynamicSystemView(c *Core) *dynamicSystemView {
	me := &MountEntry{
		Config: MountConfig{
			DefaultLeaseTTL: 24 * time.Hour,
			MaxLeaseTTL:     2 * 24 * time.Hour,
		},
	}

	return &dynamicSystemView{c, me}
}

func TestAddTestPlugin(t testing.TB, c *Core, name, testFunc string) {
	file, err := os.Open(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	hash := sha256.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		t.Fatal(err)
	}

	sum := hash.Sum(nil)
	c.pluginCatalog.directory, err = filepath.EvalSymlinks(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	c.pluginCatalog.directory = filepath.Dir(c.pluginCatalog.directory)

	command := fmt.Sprintf("%s --test.run=%s", filepath.Base(os.Args[0]), testFunc)
	err = c.pluginCatalog.Set(name, command, sum)
	if err != nil {
		t.Fatal(err)
	}
}

var testLogicalBackends = map[string]logical.Factory{}

// Starts the test server which responds to SSH authentication.
// Used to test the SSH secret backend.
func StartSSHHostTestServer() (string, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(testSharedPublicKey))
	if err != nil {
		return "", fmt.Errorf("Error parsing public key")
	}
	serverConfig := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Compare(pubKey.Marshal(), key.Marshal()) == 0 {
				return &ssh.Permissions{}, nil
			} else {
				return nil, fmt.Errorf("Key does not match")
			}
		},
	}
	signer, err := ssh.ParsePrivateKey([]byte(testSharedPrivateKey))
	if err != nil {
		panic("Error parsing private key")
	}
	serverConfig.AddHostKey(signer)

	soc, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("Error listening to connection")
	}

	go func() {
		for {
			conn, err := soc.Accept()
			if err != nil {
				panic(fmt.Sprintf("Error accepting incoming connection: %s", err))
			}
			defer conn.Close()
			sshConn, chanReqs, _, err := ssh.NewServerConn(conn, serverConfig)
			if err != nil {
				panic(fmt.Sprintf("Handshaking error: %v", err))
			}

			go func() {
				for chanReq := range chanReqs {
					go func(chanReq ssh.NewChannel) {
						if chanReq.ChannelType() != "session" {
							chanReq.Reject(ssh.UnknownChannelType, "unknown channel type")
							return
						}

						ch, requests, err := chanReq.Accept()
						if err != nil {
							panic(fmt.Sprintf("Error accepting channel: %s", err))
						}

						go func(ch ssh.Channel, in <-chan *ssh.Request) {
							for req := range in {
								executeServerCommand(ch, req)
							}
						}(ch, requests)
					}(chanReq)
				}
				sshConn.Close()
			}()
		}
	}()
	return soc.Addr().String(), nil
}

// This executes the commands requested to be run on the server.
// Used to test the SSH secret backend.
func executeServerCommand(ch ssh.Channel, req *ssh.Request) {
	command := string(req.Payload[4:])
	cmd := exec.Command("/bin/bash", []string{"-c", command}...)
	req.Reply(true, nil)

	cmd.Stdout = ch
	cmd.Stderr = ch
	cmd.Stdin = ch

	err := cmd.Start()
	if err != nil {
		panic(fmt.Sprintf("Error starting the command: '%s'", err))
	}

	go func() {
		_, err := cmd.Process.Wait()
		if err != nil {
			panic(fmt.Sprintf("Error while waiting for command to finish:'%s'", err))
		}
		ch.Close()
	}()
}

// This adds a logical backend for the test core. This needs to be
// invoked before the test core is created.
func AddTestLogicalBackend(name string, factory logical.Factory) error {
	if name == "" {
		return fmt.Errorf("Missing backend name")
	}
	if factory == nil {
		return fmt.Errorf("Missing backend factory function")
	}
	testLogicalBackends[name] = factory
	return nil
}

type noopAudit struct {
	Config    *audit.BackendConfig
	salt      *salt.Salt
	saltMutex sync.RWMutex
}

func (n *noopAudit) GetHash(data string) (string, error) {
	salt, err := n.Salt()
	if err != nil {
		return "", err
	}
	return salt.GetIdentifiedHMAC(data), nil
}

func (n *noopAudit) LogRequest(a *logical.Auth, r *logical.Request, e error) error {
	return nil
}

func (n *noopAudit) LogResponse(a *logical.Auth, r *logical.Request, re *logical.Response, err error) error {
	return nil
}

func (n *noopAudit) Reload() error {
	return nil
}

func (n *noopAudit) Invalidate() {
	n.saltMutex.Lock()
	defer n.saltMutex.Unlock()
	n.salt = nil
}

func (n *noopAudit) Salt() (*salt.Salt, error) {
	n.saltMutex.RLock()
	if n.salt != nil {
		defer n.saltMutex.RUnlock()
		return n.salt, nil
	}
	n.saltMutex.RUnlock()
	n.saltMutex.Lock()
	defer n.saltMutex.Unlock()
	if n.salt != nil {
		return n.salt, nil
	}
	salt, err := salt.NewSalt(n.Config.SaltView, n.Config.SaltConfig)
	if err != nil {
		return nil, err
	}
	n.salt = salt
	return salt, nil
}

type rawHTTP struct{}

func (n *rawHTTP) HandleRequest(req *logical.Request) (*logical.Response, error) {
	return &logical.Response{
		Data: map[string]interface{}{
			logical.HTTPStatusCode:  200,
			logical.HTTPContentType: "plain/text",
			logical.HTTPRawBody:     []byte("hello world"),
		},
	}, nil
}

func (n *rawHTTP) HandleExistenceCheck(req *logical.Request) (bool, bool, error) {
	return false, false, nil
}

func (n *rawHTTP) SpecialPaths() *logical.Paths {
	return &logical.Paths{Unauthenticated: []string{"*"}}
}

func (n *rawHTTP) System() logical.SystemView {
	return logical.StaticSystemView{
		DefaultLeaseTTLVal: time.Hour * 24,
		MaxLeaseTTLVal:     time.Hour * 24 * 32,
	}
}

func (n *rawHTTP) Logger() log.Logger {
	return logformat.NewVaultLogger(log.LevelTrace)
}

func (n *rawHTTP) Cleanup() {
	// noop
}

func (n *rawHTTP) Initialize() error {
	// noop
	return nil
}

func (n *rawHTTP) InvalidateKey(string) {
	// noop
}

func (n *rawHTTP) Setup(config *logical.BackendConfig) error {
	// noop
	return nil
}

func (n *rawHTTP) Type() logical.BackendType {
	return logical.TypeUnknown
}

func (n *rawHTTP) RegisterLicense(license interface{}) error {
	return nil
}

func GenerateRandBytes(length int) ([]byte, error) {
	if length < 0 {
		return nil, fmt.Errorf("length must be >= 0")
	}

	buf := make([]byte, length)
	if length == 0 {
		return buf, nil
	}

	n, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	if n != length {
		return nil, fmt.Errorf("unable to read %d bytes; only read %d", length, n)
	}

	return buf, nil
}

func TestWaitActive(t testing.TB, core *Core) {
	start := time.Now()
	var standby bool
	var err error
	for time.Now().Sub(start) < time.Second {
		standby, err = core.Standby()
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !standby {
			break
		}
	}
	if standby {
		t.Fatalf("should not be in standby mode")
	}
}

type TestCluster struct {
	BarrierKeys [][]byte
	CACert      *x509.Certificate
	CACertBytes []byte
	CACertPEM   []byte
	CAKey       *ecdsa.PrivateKey
	CAKeyPEM    []byte
	Cores       []*TestClusterCore
	ID          string
	RootToken   string
	RootCAs     *x509.CertPool
	TempDir     string
}

func (t *TestCluster) Start() {
	for _, core := range t.Cores {
		if core.Server != nil {
			for _, ln := range core.Listeners {
				go core.Server.Serve(ln)
			}
		}
	}
}

func (t *TestCluster) Cleanup() {
	for _, core := range t.Cores {
		if core.Listeners != nil {
			for _, ln := range core.Listeners {
				ln.Close()
			}
		}
	}

	if t.TempDir != "" {
		os.RemoveAll(t.TempDir)
	}

	// Give time to actually shut down/clean up before the next test
	time.Sleep(time.Second)
}

type TestListener struct {
	net.Listener
	Address *net.TCPAddr
}

type TestClusterCore struct {
	*Core
	Client          *api.Client
	Handler         http.Handler
	Listeners       []*TestListener
	ReloadFuncs     *map[string][]reload.ReloadFunc
	ReloadFuncsLock *sync.RWMutex
	Server          *http.Server
	ServerCert      *x509.Certificate
	ServerCertBytes []byte
	ServerCertPEM   []byte
	ServerKey       *ecdsa.PrivateKey
	ServerKeyPEM    []byte
	TLSConfig       *tls.Config
}

type TestClusterOptions struct {
	KeepStandbysSealed bool
	HandlerFunc        func(*Core) http.Handler
}

func NewTestCluster(t testing.TB, base *CoreConfig, opts *TestClusterOptions) *TestCluster {
	var testCluster TestCluster

	tempDir, err := ioutil.TempDir("", "vault-test-cluster-")
	if err != nil {
		t.Fatal(err)
	}
	testCluster.TempDir = tempDir

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	testCluster.CAKey = caKey
	caCertTemplate := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		DNSNames: []string{"localhost"},
		IPAddresses: []net.IP{
			net.IPv6loopback,
			net.ParseIP("127.0.0.1"),
		},
		KeyUsage:              x509.KeyUsage(x509.KeyUsageCertSign | x509.KeyUsageCRLSign),
		SerialNumber:          big.NewInt(mathrand.Int63()),
		NotBefore:             time.Now().Add(-30 * time.Second),
		NotAfter:              time.Now().Add(262980 * time.Hour),
		BasicConstraintsValid: true,
		IsCA: true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, caCertTemplate, caCertTemplate, caKey.Public(), caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caBytes)
	if err != nil {
		t.Fatal(err)
	}
	testCluster.CACert = caCert
	testCluster.CACertBytes = caBytes
	testCluster.RootCAs = x509.NewCertPool()
	testCluster.RootCAs.AddCert(caCert)
	caCertPEMBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}
	testCluster.CACertPEM = pem.EncodeToMemory(caCertPEMBlock)
	err = ioutil.WriteFile(filepath.Join(testCluster.TempDir, "ca_cert.pem"), testCluster.CACertPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}
	marshaledCAKey, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		t.Fatal(err)
	}
	caKeyPEMBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: marshaledCAKey,
	}
	testCluster.CAKeyPEM = pem.EncodeToMemory(caKeyPEMBlock)
	err = ioutil.WriteFile(filepath.Join(testCluster.TempDir, "ca_key.pem"), testCluster.CAKeyPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}

	s1Key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s1CertTemplate := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		DNSNames: []string{"localhost"},
		IPAddresses: []net.IP{
			net.IPv6loopback,
			net.ParseIP("127.0.0.1"),
		},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageKeyAgreement,
		SerialNumber: big.NewInt(mathrand.Int63()),
		NotBefore:    time.Now().Add(-30 * time.Second),
		NotAfter:     time.Now().Add(262980 * time.Hour),
	}
	s1CertBytes, err := x509.CreateCertificate(rand.Reader, s1CertTemplate, caCert, s1Key.Public(), caKey)
	if err != nil {
		t.Fatal(err)
	}
	s1Cert, err := x509.ParseCertificate(s1CertBytes)
	if err != nil {
		t.Fatal(err)
	}
	s1CertPEMBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: s1CertBytes,
	}
	s1CertPEM := pem.EncodeToMemory(s1CertPEMBlock)
	s1MarshaledKey, err := x509.MarshalECPrivateKey(s1Key)
	if err != nil {
		t.Fatal(err)
	}
	s1KeyPEMBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: s1MarshaledKey,
	}
	s1KeyPEM := pem.EncodeToMemory(s1KeyPEMBlock)

	s2Key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s2CertTemplate := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		DNSNames: []string{"localhost"},
		IPAddresses: []net.IP{
			net.IPv6loopback,
			net.ParseIP("127.0.0.1"),
		},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageKeyAgreement,
		SerialNumber: big.NewInt(mathrand.Int63()),
		NotBefore:    time.Now().Add(-30 * time.Second),
		NotAfter:     time.Now().Add(262980 * time.Hour),
	}
	s2CertBytes, err := x509.CreateCertificate(rand.Reader, s2CertTemplate, caCert, s2Key.Public(), caKey)
	if err != nil {
		t.Fatal(err)
	}
	s2Cert, err := x509.ParseCertificate(s2CertBytes)
	if err != nil {
		t.Fatal(err)
	}
	s2CertPEMBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: s2CertBytes,
	}
	s2CertPEM := pem.EncodeToMemory(s2CertPEMBlock)
	s2MarshaledKey, err := x509.MarshalECPrivateKey(s2Key)
	if err != nil {
		t.Fatal(err)
	}
	s2KeyPEMBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: s2MarshaledKey,
	}
	s2KeyPEM := pem.EncodeToMemory(s2KeyPEMBlock)

	s3Key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s3CertTemplate := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		DNSNames: []string{"localhost"},
		IPAddresses: []net.IP{
			net.IPv6loopback,
			net.ParseIP("127.0.0.1"),
		},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageKeyAgreement,
		SerialNumber: big.NewInt(mathrand.Int63()),
		NotBefore:    time.Now().Add(-30 * time.Second),
		NotAfter:     time.Now().Add(262980 * time.Hour),
	}
	s3CertBytes, err := x509.CreateCertificate(rand.Reader, s3CertTemplate, caCert, s3Key.Public(), caKey)
	if err != nil {
		t.Fatal(err)
	}
	s3Cert, err := x509.ParseCertificate(s3CertBytes)
	if err != nil {
		t.Fatal(err)
	}
	s3CertPEMBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: s3CertBytes,
	}
	s3CertPEM := pem.EncodeToMemory(s3CertPEMBlock)
	s3MarshaledKey, err := x509.MarshalECPrivateKey(s3Key)
	if err != nil {
		t.Fatal(err)
	}
	s3KeyPEMBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: s3MarshaledKey,
	}
	s3KeyPEM := pem.EncodeToMemory(s3KeyPEMBlock)

	logger := logformat.NewVaultLogger(log.LevelTrace)

	//
	// Listener setup
	//
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	s1CertFile := filepath.Join(testCluster.TempDir, fmt.Sprintf("node1_port_%d_cert.pem", ln.Addr().(*net.TCPAddr).Port))
	s1KeyFile := filepath.Join(testCluster.TempDir, fmt.Sprintf("node1_port_%d_key.pem", ln.Addr().(*net.TCPAddr).Port))
	err = ioutil.WriteFile(s1CertFile, s1CertPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(s1KeyFile, s1KeyPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}
	s1TLSCert, err := tls.X509KeyPair(s1CertPEM, s1KeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	s1CertGetter := reload.NewCertificateGetter(s1CertFile, s1KeyFile)
	s1TLSConfig := &tls.Config{
		Certificates:   []tls.Certificate{s1TLSCert},
		RootCAs:        testCluster.RootCAs,
		ClientCAs:      testCluster.RootCAs,
		ClientAuth:     tls.VerifyClientCertIfGiven,
		NextProtos:     []string{"h2", "http/1.1"},
		GetCertificate: s1CertGetter.GetCertificate,
	}
	s1TLSConfig.BuildNameToCertificate()
	c1lns := []*TestListener{&TestListener{
		Listener: tls.NewListener(ln, s1TLSConfig),
		Address:  ln.Addr().(*net.TCPAddr),
	},
	}
	var handler1 http.Handler = http.NewServeMux()
	server1 := &http.Server{
		Handler: handler1,
	}
	if err := http2.ConfigureServer(server1, nil); err != nil {
		t.Fatal(err)
	}

	ln, err = net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	s2CertFile := filepath.Join(testCluster.TempDir, fmt.Sprintf("node2_port_%d_cert.pem", ln.Addr().(*net.TCPAddr).Port))
	s2KeyFile := filepath.Join(testCluster.TempDir, fmt.Sprintf("node2_port_%d_key.pem", ln.Addr().(*net.TCPAddr).Port))
	err = ioutil.WriteFile(s2CertFile, s2CertPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(s2KeyFile, s2KeyPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}
	s2TLSCert, err := tls.X509KeyPair(s2CertPEM, s2KeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	s2CertGetter := reload.NewCertificateGetter(s2CertFile, s2KeyFile)
	s2TLSConfig := &tls.Config{
		Certificates:   []tls.Certificate{s2TLSCert},
		RootCAs:        testCluster.RootCAs,
		ClientCAs:      testCluster.RootCAs,
		ClientAuth:     tls.VerifyClientCertIfGiven,
		NextProtos:     []string{"h2", "http/1.1"},
		GetCertificate: s2CertGetter.GetCertificate,
	}
	s2TLSConfig.BuildNameToCertificate()
	c2lns := []*TestListener{&TestListener{
		Listener: tls.NewListener(ln, s2TLSConfig),
		Address:  ln.Addr().(*net.TCPAddr),
	},
	}
	var handler2 http.Handler = http.NewServeMux()
	server2 := &http.Server{
		Handler: handler2,
	}
	if err := http2.ConfigureServer(server2, nil); err != nil {
		t.Fatal(err)
	}

	ln, err = net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	s3CertFile := filepath.Join(testCluster.TempDir, fmt.Sprintf("node3_port_%d_cert.pem", ln.Addr().(*net.TCPAddr).Port))
	s3KeyFile := filepath.Join(testCluster.TempDir, fmt.Sprintf("node3_port_%d_key.pem", ln.Addr().(*net.TCPAddr).Port))
	err = ioutil.WriteFile(s3CertFile, s3CertPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(s3KeyFile, s3KeyPEM, 0755)
	if err != nil {
		t.Fatal(err)
	}
	s3TLSCert, err := tls.X509KeyPair(s3CertPEM, s3KeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	s3CertGetter := reload.NewCertificateGetter(s3CertFile, s3KeyFile)
	s3TLSConfig := &tls.Config{
		Certificates:   []tls.Certificate{s3TLSCert},
		RootCAs:        testCluster.RootCAs,
		ClientCAs:      testCluster.RootCAs,
		ClientAuth:     tls.VerifyClientCertIfGiven,
		NextProtos:     []string{"h2", "http/1.1"},
		GetCertificate: s3CertGetter.GetCertificate,
	}
	s3TLSConfig.BuildNameToCertificate()
	c3lns := []*TestListener{&TestListener{
		Listener: tls.NewListener(ln, s3TLSConfig),
		Address:  ln.Addr().(*net.TCPAddr),
	},
	}
	var handler3 http.Handler = http.NewServeMux()
	server3 := &http.Server{
		Handler: handler3,
	}
	if err := http2.ConfigureServer(server3, nil); err != nil {
		t.Fatal(err)
	}

	// Create three cores with the same physical and different redirect/cluster addrs
	// N.B.: On OSX, instead of random ports, it assigns new ports to new
	// listeners sequentially. Aside from being a bad idea in a security sense,
	// it also broke tests that assumed it was OK to just use the port above
	// the redirect addr. This has now been changed to 100 ports above, but if
	// we ever do more than three nodes in a cluster it may need to be bumped.
	coreConfig := &CoreConfig{
		LogicalBackends:    make(map[string]logical.Factory),
		CredentialBackends: make(map[string]logical.Factory),
		AuditBackends:      make(map[string]audit.Factory),
		RedirectAddr:       fmt.Sprintf("https://127.0.0.1:%d", c1lns[0].Address.Port),
		ClusterAddr:        fmt.Sprintf("https://127.0.0.1:%d", c1lns[0].Address.Port+100),
		DisableMlock:       true,
		EnableUI:           true,
	}

	if base != nil {
		coreConfig.DisableCache = base.DisableCache
		coreConfig.EnableUI = base.EnableUI
		coreConfig.DefaultLeaseTTL = base.DefaultLeaseTTL
		coreConfig.MaxLeaseTTL = base.MaxLeaseTTL
		coreConfig.CacheSize = base.CacheSize
		coreConfig.PluginDirectory = base.PluginDirectory
		coreConfig.Seal = base.Seal
		coreConfig.DevToken = base.DevToken

		if !coreConfig.DisableMlock {
			base.DisableMlock = false
		}

		if base.Physical != nil {
			coreConfig.Physical = base.Physical
		}

		if base.HAPhysical != nil {
			coreConfig.HAPhysical = base.HAPhysical
		}

		// Used to set something non-working to test fallback
		switch base.ClusterAddr {
		case "empty":
			coreConfig.ClusterAddr = ""
		case "":
		default:
			coreConfig.ClusterAddr = base.ClusterAddr
		}

		if base.LogicalBackends != nil {
			for k, v := range base.LogicalBackends {
				coreConfig.LogicalBackends[k] = v
			}
		}
		if base.CredentialBackends != nil {
			for k, v := range base.CredentialBackends {
				coreConfig.CredentialBackends[k] = v
			}
		}
		if base.AuditBackends != nil {
			for k, v := range base.AuditBackends {
				coreConfig.AuditBackends[k] = v
			}
		}
		if base.Logger != nil {
			coreConfig.Logger = base.Logger
		}
	}

	if coreConfig.Physical == nil {
		coreConfig.Physical = physical.NewInmem(logger)
	}
	if coreConfig.HAPhysical == nil {
		coreConfig.HAPhysical = physical.NewInmemHA(logger)
	}

	c1, err := NewCore(coreConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if opts != nil && opts.HandlerFunc != nil {
		handler1 = opts.HandlerFunc(c1)
		server1.Handler = handler1
	}

	coreConfig.RedirectAddr = fmt.Sprintf("https://127.0.0.1:%d", c2lns[0].Address.Port)
	if coreConfig.ClusterAddr != "" {
		coreConfig.ClusterAddr = fmt.Sprintf("https://127.0.0.1:%d", c2lns[0].Address.Port+100)
	}
	c2, err := NewCore(coreConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if opts != nil && opts.HandlerFunc != nil {
		handler2 = opts.HandlerFunc(c2)
		server2.Handler = handler2
	}

	coreConfig.RedirectAddr = fmt.Sprintf("https://127.0.0.1:%d", c3lns[0].Address.Port)
	if coreConfig.ClusterAddr != "" {
		coreConfig.ClusterAddr = fmt.Sprintf("https://127.0.0.1:%d", c3lns[0].Address.Port+100)
	}
	c3, err := NewCore(coreConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if opts != nil && opts.HandlerFunc != nil {
		handler3 = opts.HandlerFunc(c3)
		server3.Handler = handler3
	}

	//
	// Clustering setup
	//
	clusterAddrGen := func(lns []*TestListener) []*net.TCPAddr {
		ret := make([]*net.TCPAddr, len(lns))
		for i, ln := range lns {
			ret[i] = &net.TCPAddr{
				IP:   ln.Address.IP,
				Port: ln.Address.Port + 100,
			}
		}
		return ret
	}

	c2.SetClusterListenerAddrs(clusterAddrGen(c2lns))
	c2.SetClusterHandler(handler2)
	c3.SetClusterListenerAddrs(clusterAddrGen(c3lns))
	c3.SetClusterHandler(handler3)

	keys, root := TestCoreInitClusterWrapperSetup(t, c1, clusterAddrGen(c1lns), handler1)
	barrierKeys, _ := copystructure.Copy(keys)
	testCluster.BarrierKeys = barrierKeys.([][]byte)
	testCluster.RootToken = root

	err = ioutil.WriteFile(filepath.Join(testCluster.TempDir, "root_token"), []byte(root), 0755)
	if err != nil {
		t.Fatal(err)
	}
	marshaledKeys, err := jsonutil.EncodeJSON(barrierKeys)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(testCluster.TempDir, "barrier_keys.json"), marshaledKeys, 0755)
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range keys {
		if _, err := c1.Unseal(TestKeyCopy(key)); err != nil {
			t.Fatalf("unseal err: %s", err)
		}
	}

	// Verify unsealed
	sealed, err := c1.Sealed()
	if err != nil {
		t.Fatalf("err checking seal status: %s", err)
	}
	if sealed {
		t.Fatal("should not be sealed")
	}

	TestWaitActive(t, c1)

	if opts != nil && !opts.KeepStandbysSealed {
		for _, key := range keys {
			if _, err := c2.Unseal(TestKeyCopy(key)); err != nil {
				t.Fatalf("unseal err: %s", err)
			}
		}
		for _, key := range keys {
			if _, err := c3.Unseal(TestKeyCopy(key)); err != nil {
				t.Fatalf("unseal err: %s", err)
			}
		}

		// Let them come fully up to standby
		time.Sleep(2 * time.Second)

		// Ensure cluster connection info is populated
		isLeader, _, err := c2.Leader()
		if err != nil {
			t.Fatal(err)
		}
		if isLeader {
			t.Fatal("c2 should not be leader")
		}
		isLeader, _, err = c3.Leader()
		if err != nil {
			t.Fatal(err)
		}
		if isLeader {
			t.Fatal("c3 should not be leader")
		}
	}

	cluster, err := c1.Cluster()
	if err != nil {
		t.Fatal(err)
	}
	testCluster.ID = cluster.ID

	getAPIClient := func(port int, tlsConfig *tls.Config) *api.Client {
		transport := cleanhttp.DefaultPooledTransport()
		transport.TLSClientConfig = tlsConfig
		client := &http.Client{
			Transport: transport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				// This can of course be overridden per-test by using its own client
				return fmt.Errorf("redirects not allowed in these tests")
			},
		}
		config := api.DefaultConfig()
		config.Address = fmt.Sprintf("https://127.0.0.1:%d", port)
		config.HttpClient = client
		apiClient, err := api.NewClient(config)
		if err != nil {
			t.Fatal(err)
		}
		apiClient.SetToken(root)
		return apiClient
	}

	var ret []*TestClusterCore
	t1 := &TestClusterCore{
		Core:            c1,
		ServerKey:       s1Key,
		ServerKeyPEM:    s1KeyPEM,
		ServerCert:      s1Cert,
		ServerCertBytes: s1CertBytes,
		ServerCertPEM:   s1CertPEM,
		Listeners:       c1lns,
		Handler:         handler1,
		Server:          server1,
		TLSConfig:       s1TLSConfig,
		Client:          getAPIClient(c1lns[0].Address.Port, s1TLSConfig),
	}
	t1.ReloadFuncs = &c1.reloadFuncs
	t1.ReloadFuncsLock = &c1.reloadFuncsLock
	t1.ReloadFuncsLock.Lock()
	(*t1.ReloadFuncs)["listener|tcp"] = []reload.ReloadFunc{s1CertGetter.Reload}
	t1.ReloadFuncsLock.Unlock()
	ret = append(ret, t1)

	t2 := &TestClusterCore{
		Core:            c2,
		ServerKey:       s2Key,
		ServerKeyPEM:    s2KeyPEM,
		ServerCert:      s2Cert,
		ServerCertBytes: s2CertBytes,
		ServerCertPEM:   s2CertPEM,
		Listeners:       c2lns,
		Handler:         handler2,
		Server:          server2,
		TLSConfig:       s2TLSConfig,
		Client:          getAPIClient(c2lns[0].Address.Port, s2TLSConfig),
	}
	t2.ReloadFuncs = &c2.reloadFuncs
	t2.ReloadFuncsLock = &c2.reloadFuncsLock
	t2.ReloadFuncsLock.Lock()
	(*t2.ReloadFuncs)["listener|tcp"] = []reload.ReloadFunc{s2CertGetter.Reload}
	t2.ReloadFuncsLock.Unlock()
	ret = append(ret, t2)

	t3 := &TestClusterCore{
		Core:            c3,
		ServerKey:       s3Key,
		ServerKeyPEM:    s3KeyPEM,
		ServerCert:      s3Cert,
		ServerCertBytes: s3CertBytes,
		ServerCertPEM:   s3CertPEM,
		Listeners:       c3lns,
		Handler:         handler3,
		Server:          server3,
		TLSConfig:       s3TLSConfig,
		Client:          getAPIClient(c3lns[0].Address.Port, s3TLSConfig),
	}
	t3.ReloadFuncs = &c3.reloadFuncs
	t3.ReloadFuncsLock = &c3.reloadFuncsLock
	t3.ReloadFuncsLock.Lock()
	(*t3.ReloadFuncs)["listener|tcp"] = []reload.ReloadFunc{s3CertGetter.Reload}
	t3.ReloadFuncsLock.Unlock()
	ret = append(ret, t3)

	testCluster.Cores = ret
	return &testCluster
}

const (
	TestClusterCACert = `-----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUfIKsF2VPT7sdFcKOHJH2Ii6K4MwwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLbXl2YXVsdC5jb20wIBcNMTYwNTAyMTYwNTQyWhgPMjA2
NjA0MjAxNjA2MTJaMBYxFDASBgNVBAMTC215dmF1bHQuY29tMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuOimEXawD2qBoLCFP3Skq5zi1XzzcMAJlfdS
xz9hfymuJb+cN8rB91HOdU9wQCwVKnkUtGWxUnMp0tT0uAZj5NzhNfyinf0JGAbP
67HDzVZhGBHlHTjPX0638yaiUx90cTnucX0N20SgCYct29dMSgcPl+W78D3Jw3xE
JsHQPYS9ASe2eONxG09F/qNw7w/RO5/6WYoV2EmdarMMxq52pPe2chtNMQdSyOUb
cCcIZyk4QVFZ1ZLl6jTnUPb+JoCx1uMxXvMek4NF/5IL0Wr9dw2gKXKVKoHDr6SY
WrCONRw61A5Zwx1V+kn73YX3USRlkufQv/ih6/xThYDAXDC9cwIDAQABo4GBMH8w
DgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFOuKvPiU
G06iHkRXAOeMiUdBfHFyMB8GA1UdIwQYMBaAFOuKvPiUG06iHkRXAOeMiUdBfHFy
MBwGA1UdEQQVMBOCC215dmF1bHQuY29thwR/AAABMA0GCSqGSIb3DQEBCwUAA4IB
AQBcN/UdAMzc7UjRdnIpZvO+5keBGhL/vjltnGM1dMWYHa60Y5oh7UIXF+P1RdNW
n7g80lOyvkSR15/r1rDkqOK8/4oruXU31EcwGhDOC4hU6yMUy4ltV/nBoodHBXNh
MfKiXeOstH1vdI6G0P6W93Bcww6RyV1KH6sT2dbETCw+iq2VN9CrruGIWzd67UT/
spe/kYttr3UYVV3O9kqgffVVgVXg/JoRZ3J7Hy2UEXfh9UtWNanDlRuXaZgE9s/d
CpA30CHpNXvKeyNeW2ktv+2nAbSpvNW+e6MecBCTBIoDSkgU8ShbrzmDKVwNN66Q
5gn6KxUPBKHEtNzs5DgGM7nq
-----END CERTIFICATE-----`

	TestClusterCAKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAuOimEXawD2qBoLCFP3Skq5zi1XzzcMAJlfdSxz9hfymuJb+c
N8rB91HOdU9wQCwVKnkUtGWxUnMp0tT0uAZj5NzhNfyinf0JGAbP67HDzVZhGBHl
HTjPX0638yaiUx90cTnucX0N20SgCYct29dMSgcPl+W78D3Jw3xEJsHQPYS9ASe2
eONxG09F/qNw7w/RO5/6WYoV2EmdarMMxq52pPe2chtNMQdSyOUbcCcIZyk4QVFZ
1ZLl6jTnUPb+JoCx1uMxXvMek4NF/5IL0Wr9dw2gKXKVKoHDr6SYWrCONRw61A5Z
wx1V+kn73YX3USRlkufQv/ih6/xThYDAXDC9cwIDAQABAoIBAG3bCo7ljMQb6tel
CAUjL5Ilqz5a9ebOsONABRYLOclq4ePbatxawdJF7/sSLwZxKkIJnZtvr2Hkubxg
eOO8KC0YbVS9u39Rjc2QfobxHfsojpbWSuCJl+pvwinbkiUAUxXR7S/PtCPJKat/
fGdYCiMQ/tqnynh4vR4+/d5o12c0KuuQ22/MdEf3GOadUamRXS1ET9iJWqla1pJW
TmzrlkGAEnR5PPO2RMxbnZCYmj3dArxWAnB57W+bWYla0DstkDKtwg2j2ikNZpXB
nkZJJpxR76IYD1GxfwftqAKxujKcyfqB0dIKCJ0UmfOkauNWjexroNLwaAOC3Nud
XIxppAECgYEA1wJ9EH6A6CrSjdzUocF9LtQy1LCDHbdiQFHxM5/zZqIxraJZ8Gzh
Q0d8JeOjwPdG4zL9pHcWS7+x64Wmfn0+Qfh6/47Vy3v90PIL0AeZYshrVZyJ/s6X
YkgFK80KEuWtacqIZ1K2UJyCw81u/ynIl2doRsIbgkbNeN0opjmqVTMCgYEA3CkW
2fETWK1LvmgKFjG1TjOotVRIOUfy4iN0kznPm6DK2PgTF5DX5RfktlmA8i8WPmB7
YFOEdAWHf+RtoM/URa7EAGZncCWe6uggAcWqznTS619BJ63OmncpSWov5Byg90gJ
48qIMY4wDjE85ypz1bmBc2Iph974dtWeDtB7dsECgYAyKZh4EquMfwEkq9LH8lZ8
aHF7gbr1YeWAUB3QB49H8KtacTg+iYh8o97pEBUSXh6hvzHB/y6qeYzPAB16AUpX
Jdu8Z9ylXsY2y2HKJRu6GjxAewcO9bAH8/mQ4INrKT6uIdx1Dq0OXZV8jR9KVLtB
55RCfeLhIBesDR0Auw9sVQKBgB0xTZhkgP43LF35Ca1btgDClNJGdLUztx8JOIH1
HnQyY/NVIaL0T8xO2MLdJ131pGts+68QI/YGbaslrOuv4yPCQrcS3RBfzKy1Ttkt
TrLFhtoy7T7HqyeMOWtEq0kCCs3/PWB5EIoRoomfOcYlOOrUCDg2ge9EP4nyVVz9
hAGBAoGBAJXw/ufevxpBJJMSyULmVWYr34GwLC1OhSE6AVVt9JkIYnc5L4xBKTHP
QNKKJLmFmMsEqfxHUNWmpiHkm2E0p37Zehui3kywo+A4ybHPTua70ZWQfZhKxLUr
PvJa8JmwiCM7kO8zjOv+edY1mMWrbjAZH1YUbfcTHmST7S8vp0F3
-----END RSA PRIVATE KEY-----`

	TestClusterServerCert = `-----BEGIN CERTIFICATE-----
MIIDtzCCAp+gAwIBAgIUBLqh6ctGWVDUxFhxJX7m6S/bnrcwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLbXl2YXVsdC5jb20wIBcNMTYwNTAyMTYwOTI2WhgPMjA2
NjA0MjAxNTA5NTZaMBsxGTAXBgNVBAMTEGNlcnQubXl2YXVsdC5jb20wggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDY3gPB29kkdbu0mPO6J0efagQhSiXB
9OyDuLf5sMk6CVDWVWal5hISkyBmw/lXgF7qC2XFKivpJOrcGQd5Ep9otBqyJLzI
b0IWdXuPIrVnXDwcdWr86ybX2iC42zKWfbXgjzGijeAVpl0UJLKBj+fk5q6NvkRL
5FUL6TRV7Krn9mrmnrV9J5IqV15pTd9W2aVJ6IqWvIPCACtZKulqWn4707uy2X2W
1Stq/5qnp1pDshiGk1VPyxCwQ6yw3iEcgecbYo3vQfhWcv7Q8LpSIM9ZYpXu6OmF
+czqRZS9gERl+wipmmrN1MdYVrTuQem21C/PNZ4jo4XUk1SFx6JrcA+lAgMBAAGj
gfUwgfIwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBSe
Cl9WV3BjGCwmS/KrDSLRjfwyqjAfBgNVHSMEGDAWgBTrirz4lBtOoh5EVwDnjIlH
QXxxcjA7BggrBgEFBQcBAQQvMC0wKwYIKwYBBQUHMAKGH2h0dHA6Ly8xMjcuMC4w
LjE6ODIwMC92MS9wa2kvY2EwIQYDVR0RBBowGIIQY2VydC5teXZhdWx0LmNvbYcE
fwAAATAxBgNVHR8EKjAoMCagJKAihiBodHRwOi8vMTI3LjAuMC4xOjgyMDAvdjEv
cGtpL2NybDANBgkqhkiG9w0BAQsFAAOCAQEAWGholPN8buDYwKbUiDavbzjsxUIX
lU4MxEqOHw7CD3qIYIauPboLvB9EldBQwhgOOy607Yvdg3rtyYwyBFwPhHo/hK3Z
6mn4hc6TF2V+AUdHBvGzp2dbYLeo8noVoWbQ/lBulggwlIHNNF6+a3kALqsqk1Ch
f/hzsjFnDhAlNcYFgG8TgfE2lE/FckvejPqBffo7Q3I+wVAw0buqiz5QL81NOT+D
Y2S9LLKLRaCsWo9wRU1Az4Rhd7vK5SEMh16jJ82GyEODWPvuxOTI1MnzfnbWyLYe
TTp6YBjGMVf1I6NEcWNur7U17uIOiQjMZ9krNvoMJ1A/cxCoZ98QHgcIPg==
-----END CERTIFICATE-----`

	TestClusterServerKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA2N4DwdvZJHW7tJjzuidHn2oEIUolwfTsg7i3+bDJOglQ1lVm
peYSEpMgZsP5V4Be6gtlxSor6STq3BkHeRKfaLQasiS8yG9CFnV7jyK1Z1w8HHVq
/Osm19oguNsyln214I8xoo3gFaZdFCSygY/n5Oaujb5ES+RVC+k0Veyq5/Zq5p61
fSeSKldeaU3fVtmlSeiKlryDwgArWSrpalp+O9O7stl9ltUrav+ap6daQ7IYhpNV
T8sQsEOssN4hHIHnG2KN70H4VnL+0PC6UiDPWWKV7ujphfnM6kWUvYBEZfsIqZpq
zdTHWFa07kHpttQvzzWeI6OF1JNUhceia3APpQIDAQABAoIBAQCH3vEzr+3nreug
RoPNCXcSJXXY9X+aeT0FeeGqClzIg7Wl03OwVOjVwl/2gqnhbIgK0oE8eiNwurR6
mSPZcxV0oAJpwiKU4T/imlCDaReGXn86xUX2l82KRxthNdQH/VLKEmzij0jpx4Vh
bWx5SBPdkbmjDKX1dmTiRYWIn/KjyNPvNvmtwdi8Qluhf4eJcNEUr2BtblnGOmfL
FdSu+brPJozpoQ1QdDnbAQRgqnh7Shl0tT85whQi0uquqIj1gEOGVjmBvDDnL3GV
WOENTKqsmIIoEzdZrql1pfmYTk7WNaD92bfpN128j8BF7RmAV4/DphH0pvK05y9m
tmRhyHGxAoGBAOV2BBocsm6xup575VqmFN+EnIOiTn+haOvfdnVsyQHnth63fOQx
PNtMpTPR1OMKGpJ13e2bV0IgcYRsRkScVkUtoa/17VIgqZXffnJJ0A/HT67uKBq3
8o7RrtyK5N20otw0lZHyqOPhyCdpSsurDhNON1kPVJVYY4N1RiIxfut/AoGBAPHz
HfsJ5ZkyELE9N/r4fce04lprxWH+mQGK0/PfjS9caXPhj/r5ZkVMvzWesF3mmnY8
goE5S35TuTvV1+6rKGizwlCFAQlyXJiFpOryNWpLwCmDDSzLcm+sToAlML3tMgWU
jM3dWHx3C93c3ft4rSWJaUYI9JbHsMzDW6Yh+GbbAoGBANIbKwxh5Hx5XwEJP2yu
kIROYCYkMy6otHLujgBdmPyWl+suZjxoXWoMl2SIqR8vPD+Jj6mmyNJy9J6lqf3f
DRuQ+fEuBZ1i7QWfvJ+XuN0JyovJ5Iz6jC58D1pAD+p2IX3y5FXcVQs8zVJRFjzB
p0TEJOf2oqORaKWRd6ONoMKvAoGALKu6aVMWdQZtVov6/fdLIcgf0pn7Q3CCR2qe
X3Ry2L+zKJYIw0mwvDLDSt8VqQCenB3n6nvtmFFU7ds5lvM67rnhsoQcAOaAehiS
rl4xxoJd5Ewx7odRhZTGmZpEOYzFo4odxRSM9c30/u18fqV1Mm0AZtHYds4/sk6P
aUj0V+kCgYBMpGrJk8RSez5g0XZ35HfpI4ENoWbiwB59FIpWsLl2LADEh29eC455
t9Muq7MprBVBHQo11TMLLFxDIjkuMho/gcKgpYXCt0LfiNm8EZehvLJUXH+3WqUx
we6ywrbFCs6LaxaOCtTiLsN+GbZCatITL0UJaeBmTAbiw0KQjUuZPQ==
-----END RSA PRIVATE KEY-----`
)
