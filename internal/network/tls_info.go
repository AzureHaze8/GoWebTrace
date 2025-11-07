package network

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
)

// ExtractTLSInfo 从 TLS 连接状态中提取 TLS 信息
func ExtractTLSInfo(connState *tls.ConnectionState) *TLSInfo {
	if connState == nil {
		return nil
	}

	tlsInfo := &TLSInfo{
		Version:     tls.VersionName(connState.Version),
		CipherSuite: tls.CipherSuiteName(connState.CipherSuite),
	}

	if len(connState.PeerCertificates) > 0 {
		cert := connState.PeerCertificates[0]

		var certPublicKey string
		if pubKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey); err == nil {
			certPublicKey = hex.EncodeToString(pubKeyBytes)
		}

		var ipAddresses []string
		for _, ip := range cert.IPAddresses {
			ipAddresses = append(ipAddresses, ip.String())
		}

		var uris []string
		for _, uri := range cert.URIs {
			uris = append(uris, uri.String())
		}

		fingerprintSHA1 := sha1.Sum(cert.Raw)
		fingerprintSHA256 := sha256.Sum256(cert.Raw)

		tlsInfo.Certificate = &CertInfo{
			FingerprintSHA1:       hex.EncodeToString(fingerprintSHA1[:]),
			FingerprintSHA256:     hex.EncodeToString(fingerprintSHA256[:]),
			Subject:               cert.Subject.String(),
			Issuer:                cert.Issuer.String(),
			NotBefore:             cert.NotBefore,
			NotAfter:              cert.NotAfter,
			SerialNumber:          cert.SerialNumber.String(),
			PublicKeyAlgorithm:    cert.PublicKeyAlgorithm.String(),
			PublicKey:             certPublicKey,
			SignatureAlgorithm:    cert.SignatureAlgorithm.String(),
			Version:               cert.Version,
			OCSPStapling:          connState.OCSPResponse != nil,
			OCSPReponder:          cert.OCSPServer,
			DNSNames:              cert.DNSNames,
			EmailAddresses:        cert.EmailAddresses,
			IPAddresses:           ipAddresses,
			URIs:                  uris,
			CRLDistributionPoints: cert.CRLDistributionPoints,
			IssuingCertificateURL: cert.IssuingCertificateURL,
		}
	}

	return tlsInfo
}