package tlsutil

import (
	"crypto/tls"
	"path/filepath"
	"testing"
)

func TestServerConfigGeneratesThenReusesCertificate(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		CertFile: filepath.Join(dir, "server.crt"),
		KeyFile:  filepath.Join(dir, "private", "server.key"),
		Hosts:    []string{"127.0.0.1", "localhost"},
	}

	cfg, fp1, err := ServerConfig(opts)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatal("expected TLS 1.2 minimum")
	}
	if len(fp1) != 64 {
		t.Fatalf("unexpected fingerprint %q", fp1)
	}

	_, fp2, err := ServerConfig(opts)
	if err != nil {
		t.Fatal(err)
	}
	if fp1 != fp2 {
		t.Fatal("existing certificate should be reused, fingerprint changed")
	}
}

func TestTLSHandshakeWithGeneratedCertificate(t *testing.T) {
	dir := t.TempDir()
	cfg, fp, err := ServerConfig(Options{
		CertFile: filepath.Join(dir, "c.crt"),
		KeyFile:  filepath.Join(dir, "c.key"),
		Hosts:    []string{"127.0.0.1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err == nil {
			_ = c.(*tls.Conn).Handshake()
			c.Close()
		}
	}()

	// 客户端用指纹固定证书，而不是依赖 CA
	conn, err := tls.Dial("tcp", ln.Addr().String(), &tls.Config{
		InsecureSkipVerify: true,
		VerifyConnection: func(cs tls.ConnectionState) error {
			if Fingerprint(cs.PeerCertificates[0].Raw) != fp {
				t.Error("fingerprint mismatch")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
}
