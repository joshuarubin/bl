package selfcert

import (
	"crypto/ecdsa"
	"crypto/x509"
	"net"
	"testing"
)

func TestNewCert(t *testing.T) {
	cert, err := NewCert("::1", "localhost")
	if err != nil {
		t.Fatal(err)
	}

	if len(cert.Certificate) == 0 {
		t.Fatal("empty certificate")
	}

	xc, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := xc.PublicKey.(*ecdsa.PublicKey); !ok {
		t.Fatal("invalid cert generated")
	}

	if len(xc.Issuer.Organization) != 1 {
		t.Fatal("invalid cert generated")
	}

	if xc.Issuer.Organization[0] != organization {
		t.Fatal("invalid cert generated")
	}

	if len(xc.DNSNames) != 1 {
		t.Fatal("invalid cert generated")
	}

	if xc.DNSNames[0] != "localhost" {
		t.Fatal("invalid cert generated")
	}

	if len(xc.IPAddresses) != 1 {
		t.Fatal("invalid cert generated")
	}

	if !net.ParseIP("::1").Equal(xc.IPAddresses[0]) {
		t.Fatal("invalid cert generated")
	}
}
