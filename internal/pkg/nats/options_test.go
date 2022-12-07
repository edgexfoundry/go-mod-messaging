//
// Copyright (c) 2022 One Track Consulting
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

//go:build include_nats_messaging

package nats

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v3/internal/pkg"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-messaging/v3/pkg/types"
)

func TestCreateClientConfiguration(t *testing.T) {
	type args struct {
		messageBusConfig types.MessageBusConfig
	}
	tests := []struct {
		name    string
		args    args
		want    ClientConfig
		wantErr bool
	}{
		{
			"Successfully load all configurations",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.Username:                "TestUser",
					pkg.Password:                "TestPassword",
					pkg.ClientId:                "TestClientID",
					pkg.Durable:                 "durable-1",
					pkg.Subject:                 "stream/#",
					pkg.Format:                  "format",
					pkg.AutoProvision:           "true",
					pkg.RetryOnFailedConnect:    "true",
					pkg.ConnectTimeout:          "7",
					pkg.QueueGroup:              "group-1",
					pkg.DefaultPubRetryAttempts: "12",
					pkg.Deliver:                 "set-deliver",
				}}},
			ClientConfig{
				BrokerURL: "tcp://example.com:9090",
				ClientOptions: ClientOptions{
					Username:                "TestUser",
					Password:                "TestPassword",
					ClientId:                "TestClientID",
					AutoProvision:           true,
					RetryOnFailedConnect:    true,
					Durable:                 "durable-1",
					Subject:                 "stream/#",
					Format:                  "format",
					ConnectTimeout:          7,
					QueueGroup:              "group-1",
					DefaultPubRetryAttempts: 12,
					Deliver:                 "set-deliver",
				},
			},
			false,
		},
		{
			"Does not over write host configuration with optional properties",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.Username:                "TestUser",
					pkg.Password:                "TestPassword",
					pkg.ClientId:                "TestClientID",
					pkg.KeepAlive:               "3",
					pkg.Retained:                "true",
					pkg.CleanSession:            "false",
					pkg.ConnectTimeout:          "7",
					pkg.DefaultPubRetryAttempts: "2",
				}}},
			ClientConfig{
				BrokerURL: "tcp://example.com:9090",
				ClientOptions: ClientOptions{
					Username:                "TestUser",
					Password:                "TestPassword",
					ClientId:                "TestClientID",
					ConnectTimeout:          7,
					DefaultPubRetryAttempts: 2,
					Format:                  "nats",
				}},
			false,
		},
		{
			"Invalid URL",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "   ", Port: 999999999999, Protocol: "    "},
				Optional: map[string]string{
					// Other valid configurations
					"ClientId": "TestClientID",
				}}},
			ClientConfig{},
			true,
		},
		{
			"Invalid Int",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.ConnectTimeout: "abc",
					// Other valid configurations
				}}},
			ClientConfig{},
			true,
		},
		{
			"Invalid Bool",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.RetryOnFailedConnect: "abc",
				}}},
			ClientConfig{},
			true,
		},
		{
			"Unknown configuration",
			args{types.MessageBusConfig{
				PublishHost: types.HostInfo{Host: "example.com", Port: 9090, Protocol: "tcp"},
				Optional: map[string]string{
					pkg.Username:       "TestUser",
					pkg.Password:       "TestPassword",
					pkg.ClientId:       "TestClientID",
					pkg.ConnectTimeout: "7",
					"Unknown config":   "Something random",
				}}},
			ClientConfig{
				BrokerURL: "tcp://example.com:9090",
				ClientOptions: ClientOptions{
					Username:                "TestUser",
					Password:                "TestPassword",
					ClientId:                "TestClientID",
					ConnectTimeout:          7,
					DefaultPubRetryAttempts: 2,
					Format:                  "nats",
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateClientConfiguration(tt.args.messageBusConfig)
			if tt.wantErr {
				require.Error(t, err, "CreateClientConfiguration() error = %v, wantErr %v", err, tt.wantErr)
				return // End test for expected errors
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClientConfig_ConnectOpt(t *testing.T) {
	tests := []struct {
		name    string
		sut     ClientConfig
		wantLen int
		wantErr assert.ErrorAssertionFunc
	}{
		{"default", ClientConfig{}, 2, assert.NoError},
		{"ClientId", ClientConfig{ClientOptions: ClientOptions{Username: uuid.NewString()}}, 3, assert.NoError},
		{"Username", ClientConfig{ClientOptions: ClientOptions{Username: uuid.NewString()}}, 3, assert.NoError},
		{"TLS", ClientConfig{
			BrokerURL: "tls://broker",
			ClientOptions: ClientOptions{
				TlsConfigurationOptions: pkg.TlsConfigurationOptions{
					SkipCertVerify: true,
					CertPEMBlock: `-----BEGIN CERTIFICATE-----
MIID0DCCArigAwIBAgIBATANBgkqhkiG9w0BAQUFADB/MQswCQYDVQQGEwJGUjET
MBEGA1UECAwKU29tZS1TdGF0ZTEOMAwGA1UEBwwFUGFyaXMxDTALBgNVBAoMBERp
bWkxDTALBgNVBAsMBE5TQlUxEDAOBgNVBAMMB0RpbWkgQ0ExGzAZBgkqhkiG9w0B
CQEWDGRpbWlAZGltaS5mcjAeFw0xNDAxMjgyMDM2NTVaFw0yNDAxMjYyMDM2NTVa
MFsxCzAJBgNVBAYTAkZSMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJ
bnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFDASBgNVBAMMC3d3dy5kaW1pLmZyMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvpnaPKLIKdvx98KW68lz8pGa
RRcYersNGqPjpifMVjjE8LuCoXgPU0HePnNTUjpShBnynKCvrtWhN+haKbSp+QWX
SxiTrW99HBfAl1MDQyWcukoEb9Cw6INctVUN4iRvkn9T8E6q174RbcnwA/7yTc7p
1NCvw+6B/aAN9l1G2pQXgRdYC/+G6o1IZEHtWhqzE97nY5QKNuUVD0V09dc5CDYB
aKjqetwwv6DFk/GRdOSEd/6bW+20z0qSHpa3YNW6qSp+x5pyYmDrzRIR03os6Dau
ZkChSRyc/Whvurx6o85D6qpzywo8xwNaLZHxTQPgcIA5su9ZIytv9LH2E+lSwwID
AQABo3sweTAJBgNVHRMEAjAAMCwGCWCGSAGG+EIBDQQfFh1PcGVuU1NMIEdlbmVy
YXRlZCBDZXJ0aWZpY2F0ZTAdBgNVHQ4EFgQU+tugFtyN+cXe1wxUqeA7X+yS3bgw
HwYDVR0jBBgwFoAUhMwqkbBrGp87HxfvwgPnlGgVR64wDQYJKoZIhvcNAQEFBQAD
ggEBAIEEmqqhEzeXZ4CKhE5UM9vCKzkj5Iv9TFs/a9CcQuepzplt7YVmevBFNOc0
+1ZyR4tXgi4+5MHGzhYCIVvHo4hKqYm+J+o5mwQInf1qoAHuO7CLD3WNa1sKcVUV
vepIxc/1aHZrG+dPeEHt0MdFfOw13YdUc2FH6AqEdcEL4aV5PXq2eYR8hR4zKbc1
fBtuqUsvA8NWSIyzQ16fyGve+ANf6vXvUizyvwDrPRv/kfvLNa3ZPnLMMxU98Mvh
PXy3PkB8++6U4Y3vdk2Ni2WYYlIls8yqbM4327IKmkDc2TimS8u60CT47mKU7aDY
cbTV5RDkrlaYwm5yqlTIglvCv7o=
-----END CERTIFICATE-----`,
					KeyPEMBlock: `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvpnaPKLIKdvx98KW68lz8pGaRRcYersNGqPjpifMVjjE8LuC
oXgPU0HePnNTUjpShBnynKCvrtWhN+haKbSp+QWXSxiTrW99HBfAl1MDQyWcukoE
b9Cw6INctVUN4iRvkn9T8E6q174RbcnwA/7yTc7p1NCvw+6B/aAN9l1G2pQXgRdY
C/+G6o1IZEHtWhqzE97nY5QKNuUVD0V09dc5CDYBaKjqetwwv6DFk/GRdOSEd/6b
W+20z0qSHpa3YNW6qSp+x5pyYmDrzRIR03os6DauZkChSRyc/Whvurx6o85D6qpz
ywo8xwNaLZHxTQPgcIA5su9ZIytv9LH2E+lSwwIDAQABAoIBAFml8cD9a5pMqlW3
f9btTQz1sRL4Fvp7CmHSXhvjsjeHwhHckEe0ObkWTRsgkTsm1XLu5W8IITnhn0+1
iNr+78eB+rRGngdAXh8diOdkEy+8/Cee8tFI3jyutKdRlxMbwiKsouVviumoq3fx
OGQYwQ0Z2l/PvCwy/Y82ffq3ysC5gAJsbBYsCrg14bQo44ulrELe4SDWs5HCjKYb
EI2b8cOMucqZSOtxg9niLN/je2bo/I2HGSawibgcOdBms8k6TvsSrZMr3kJ5O6J+
77LGwKH37brVgbVYvbq6nWPL0xLG7dUv+7LWEo5qQaPy6aXb/zbckqLqu6/EjOVe
ydG5JQECgYEA9kKfTZD/WEVAreA0dzfeJRu8vlnwoagL7cJaoDxqXos4mcr5mPDT
kbWgFkLFFH/AyUnPBlK6BcJp1XK67B13ETUa3i9Q5t1WuZEobiKKBLFm9DDQJt43
uKZWJxBKFGSvFrYPtGZst719mZVcPct2CzPjEgN3Hlpt6fyw3eOrnoECgYEAxiOu
jwXCOmuGaB7+OW2tR0PGEzbvVlEGdkAJ6TC/HoKM1A8r2u4hLTEJJCrLLTfw++4I
ddHE2dLeR4Q7O58SfLphwgPmLDezN7WRLGr7Vyfuv7VmaHjGuC3Gv9agnhWDlA2Q
gBG9/R9oVfL0Dc7CgJgLeUtItCYC31bGT3yhV0MCgYEA4k3DG4L+RN4PXDpHvK9I
pA1jXAJHEifeHnaW1d3vWkbSkvJmgVf+9U5VeV+OwRHN1qzPZV4suRI6M/8lK8rA
Gr4UnM4aqK4K/qkY4G05LKrik9Ev2CgqSLQDRA7CJQ+Jn3Nb50qg6hFnFPafN+J7		
7juWln08wFYV4Atpdd+9XQECgYBxizkZFL+9IqkfOcONvWAzGo+Dq1N0L3J4iTIk
w56CKWXyj88d4qB4eUU3yJ4uB4S9miaW/eLEwKZIbWpUPFAn0db7i6h3ZmP5ZL8Q
qS3nQCb9DULmU2/tU641eRUKAmIoka1g9sndKAZuWo+o6fdkIb1RgObk9XNn8R4r
psv+aQKBgB+CIcExR30vycv5bnZN9EFlIXNKaeMJUrYCXcRQNvrnUIUBvAO8+jAe
CdLygS5RtgOLZib0IVErqWsP3EI1ACGuLts0vQ9GFLQGaN1SaMS40C9kvns1mlDu
LhIhYpJ8UsCVt5snWo2N+M+6ANh5tpWdQnEK6zILh4tRbuzaiHgb
-----END RSA PRIVATE KEY-----`,
				},
			}}, 3, assert.NoError},
		{"NATS+TLS", ClientConfig{
			BrokerURL: "nats://broker",
			ClientOptions: ClientOptions{
				TlsConfigurationOptions: pkg.TlsConfigurationOptions{
					SkipCertVerify: true,
					CertPEMBlock: `-----BEGIN CERTIFICATE-----
MIID0DCCArigAwIBAgIBATANBgkqhkiG9w0BAQUFADB/MQswCQYDVQQGEwJGUjET
MBEGA1UECAwKU29tZS1TdGF0ZTEOMAwGA1UEBwwFUGFyaXMxDTALBgNVBAoMBERp
bWkxDTALBgNVBAsMBE5TQlUxEDAOBgNVBAMMB0RpbWkgQ0ExGzAZBgkqhkiG9w0B
CQEWDGRpbWlAZGltaS5mcjAeFw0xNDAxMjgyMDM2NTVaFw0yNDAxMjYyMDM2NTVa
MFsxCzAJBgNVBAYTAkZSMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJ
bnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFDASBgNVBAMMC3d3dy5kaW1pLmZyMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvpnaPKLIKdvx98KW68lz8pGa
RRcYersNGqPjpifMVjjE8LuCoXgPU0HePnNTUjpShBnynKCvrtWhN+haKbSp+QWX
SxiTrW99HBfAl1MDQyWcukoEb9Cw6INctVUN4iRvkn9T8E6q174RbcnwA/7yTc7p
1NCvw+6B/aAN9l1G2pQXgRdYC/+G6o1IZEHtWhqzE97nY5QKNuUVD0V09dc5CDYB
aKjqetwwv6DFk/GRdOSEd/6bW+20z0qSHpa3YNW6qSp+x5pyYmDrzRIR03os6Dau
ZkChSRyc/Whvurx6o85D6qpzywo8xwNaLZHxTQPgcIA5su9ZIytv9LH2E+lSwwID
AQABo3sweTAJBgNVHRMEAjAAMCwGCWCGSAGG+EIBDQQfFh1PcGVuU1NMIEdlbmVy
YXRlZCBDZXJ0aWZpY2F0ZTAdBgNVHQ4EFgQU+tugFtyN+cXe1wxUqeA7X+yS3bgw
HwYDVR0jBBgwFoAUhMwqkbBrGp87HxfvwgPnlGgVR64wDQYJKoZIhvcNAQEFBQAD
ggEBAIEEmqqhEzeXZ4CKhE5UM9vCKzkj5Iv9TFs/a9CcQuepzplt7YVmevBFNOc0
+1ZyR4tXgi4+5MHGzhYCIVvHo4hKqYm+J+o5mwQInf1qoAHuO7CLD3WNa1sKcVUV
vepIxc/1aHZrG+dPeEHt0MdFfOw13YdUc2FH6AqEdcEL4aV5PXq2eYR8hR4zKbc1
fBtuqUsvA8NWSIyzQ16fyGve+ANf6vXvUizyvwDrPRv/kfvLNa3ZPnLMMxU98Mvh
PXy3PkB8++6U4Y3vdk2Ni2WYYlIls8yqbM4327IKmkDc2TimS8u60CT47mKU7aDY
cbTV5RDkrlaYwm5yqlTIglvCv7o=
-----END CERTIFICATE-----`,
					KeyPEMBlock: `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvpnaPKLIKdvx98KW68lz8pGaRRcYersNGqPjpifMVjjE8LuC
oXgPU0HePnNTUjpShBnynKCvrtWhN+haKbSp+QWXSxiTrW99HBfAl1MDQyWcukoE
b9Cw6INctVUN4iRvkn9T8E6q174RbcnwA/7yTc7p1NCvw+6B/aAN9l1G2pQXgRdY
C/+G6o1IZEHtWhqzE97nY5QKNuUVD0V09dc5CDYBaKjqetwwv6DFk/GRdOSEd/6b
W+20z0qSHpa3YNW6qSp+x5pyYmDrzRIR03os6DauZkChSRyc/Whvurx6o85D6qpz
ywo8xwNaLZHxTQPgcIA5su9ZIytv9LH2E+lSwwIDAQABAoIBAFml8cD9a5pMqlW3
f9btTQz1sRL4Fvp7CmHSXhvjsjeHwhHckEe0ObkWTRsgkTsm1XLu5W8IITnhn0+1
iNr+78eB+rRGngdAXh8diOdkEy+8/Cee8tFI3jyutKdRlxMbwiKsouVviumoq3fx
OGQYwQ0Z2l/PvCwy/Y82ffq3ysC5gAJsbBYsCrg14bQo44ulrELe4SDWs5HCjKYb
EI2b8cOMucqZSOtxg9niLN/je2bo/I2HGSawibgcOdBms8k6TvsSrZMr3kJ5O6J+
77LGwKH37brVgbVYvbq6nWPL0xLG7dUv+7LWEo5qQaPy6aXb/zbckqLqu6/EjOVe
ydG5JQECgYEA9kKfTZD/WEVAreA0dzfeJRu8vlnwoagL7cJaoDxqXos4mcr5mPDT
kbWgFkLFFH/AyUnPBlK6BcJp1XK67B13ETUa3i9Q5t1WuZEobiKKBLFm9DDQJt43
uKZWJxBKFGSvFrYPtGZst719mZVcPct2CzPjEgN3Hlpt6fyw3eOrnoECgYEAxiOu
jwXCOmuGaB7+OW2tR0PGEzbvVlEGdkAJ6TC/HoKM1A8r2u4hLTEJJCrLLTfw++4I
ddHE2dLeR4Q7O58SfLphwgPmLDezN7WRLGr7Vyfuv7VmaHjGuC3Gv9agnhWDlA2Q
gBG9/R9oVfL0Dc7CgJgLeUtItCYC31bGT3yhV0MCgYEA4k3DG4L+RN4PXDpHvK9I
pA1jXAJHEifeHnaW1d3vWkbSkvJmgVf+9U5VeV+OwRHN1qzPZV4suRI6M/8lK8rA
Gr4UnM4aqK4K/qkY4G05LKrik9Ev2CgqSLQDRA7CJQ+Jn3Nb50qg6hFnFPafN+J7		
7juWln08wFYV4Atpdd+9XQECgYBxizkZFL+9IqkfOcONvWAzGo+Dq1N0L3J4iTIk
w56CKWXyj88d4qB4eUU3yJ4uB4S9miaW/eLEwKZIbWpUPFAn0db7i6h3ZmP5ZL8Q
qS3nQCb9DULmU2/tU641eRUKAmIoka1g9sndKAZuWo+o6fdkIb1RgObk9XNn8R4r
psv+aQKBgB+CIcExR30vycv5bnZN9EFlIXNKaeMJUrYCXcRQNvrnUIUBvAO8+jAe
CdLygS5RtgOLZib0IVErqWsP3EI1ACGuLts0vQ9GFLQGaN1SaMS40C9kvns1mlDu
LhIhYpJ8UsCVt5snWo2N+M+6ANh5tpWdQnEK6zILh4tRbuzaiHgb
-----END RSA PRIVATE KEY-----`,
				},
			}}, 3, assert.NoError},
		{"TLS error", ClientConfig{
			BrokerURL: "tls://broker",
			ClientOptions: ClientOptions{
				TlsConfigurationOptions: pkg.TlsConfigurationOptions{
					SkipCertVerify: true,
					CertPEMBlock:   "not PEM",
					KeyPEMBlock:    "not PEM",
				},
			}}, 0, assert.Error},
		{"NKEYS error", ClientConfig{
			BrokerURL: "nats://broker",
			ClientOptions: ClientOptions{
				NKeySeedFile: "/fake/file/path.seed",
			}}, 0, assert.Error},
		// the NKeyOptionFromSeed option will validate the seed file on initialization
		{"NKEYS", ClientConfig{
			BrokerURL: "nats://broker",
			ClientOptions: ClientOptions{
				NKeySeedFile: "testdata/nkey.seed",
			}}, 3, assert.NoError},
		// the UserCredentials option does not validate the file passed on initialization
		{"Credentials", ClientConfig{
			BrokerURL: "nats://broker",
			ClientOptions: ClientOptions{
				CredentialsFile: "/fake/file/credentials",
			}}, 3, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.sut.ConnectOpt()
			if !tt.wantErr(t, err, fmt.Sprintf("ConnectOpt()")) {
				return
			}
			assert.Equalf(t, tt.wantLen, len(got), "ConnectOpt()")
		})
	}
}
