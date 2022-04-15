package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/feng081212/fabric-protos-go/peer"
	"github.com/feng081212/fabric-sdk-go/client"
	"github.com/feng081212/fabric-sdk-go/fabric/chaincode/ccpackager/lifecycle"
	"github.com/feng081212/fabric-sdk-go/fabric/endpoints"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"fmt"
)

var zcyAdminCertificate = "-----BEGIN CERTIFICATE-----\nMIICSTCCAe+gAwIBAgIIcXpsNp1IIAUwCgYIKoZIzj0EAwIwXDELMAkGA1UEBhMC\nQ04xFTATBgNVBAkMDOi9rOWhmOihl+mBkzEMMAoGA1UEChMDemN5MQswCQYDVQQL\nEwJjYTEbMBkGA1UEAxMSY2EuemN5Lmxhemllc3QuY29tMB4XDTIyMDMwNzAxNTgx\nOFoXDTIzMDMwNzAxNTgxOFowYjELMAkGA1UEBhMCQ04xFTATBgNVBAkMDOi9rOWh\nmOihl+mBkzEMMAoGA1UEChMDemN5MQ4wDAYDVQQLEwVhZG1pbjEeMBwGA1UEAxMV\nYWRtaW4uemN5Lmxhemllc3QuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\nYaJhZji8d1G+TGtouZTYo09KGIOkvpHFepKuNhfwfgLZoTekwsG+XZ3CeKUWcv+u\n9tAsMT06nnk/UKYYrAHFGaOBlDCBkTAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYw\nFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBQ+dbT/yU8ZMli0q5xPuRna\nDPknVzAfBgNVHSMEGDAWgBSl7LYUHsWCV/0BygRSz7tIaB6fuTAgBgNVHREEGTAX\nghVhZG1pbi56Y3kubGF6aWVzdC5jb20wCgYIKoZIzj0EAwIDSAAwRQIgKWkTlFse\nn3+EX3UPpQspcWf6kXTDNFuB2skhtDYElLoCIQCTtWKpV2qKXeuFZYqhb1ab7O4j\nanzDF/RNgXFDwZsd4Q==\n-----END CERTIFICATE-----"
var zcyAdminPrivateKey = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIOW7lxJE/LLInhP3HkgmmRKwjYaDKI1uN/lNBwaEQbQJoAoGCCqGSM49\nAwEHoUQDQgAEYaJhZji8d1G+TGtouZTYo09KGIOkvpHFepKuNhfwfgLZoTekwsG+\nXZ3CeKUWcv+u9tAsMT06nnk/UKYYrAHFGQ==\n-----END EC PRIVATE KEY-----"
var zcyTlsCertificate = "-----BEGIN CERTIFICATE-----\nMIICQjCCAemgAwIBAgIIbuPEzfdurcgwCgYIKoZIzj0EAwIwXDELMAkGA1UEBhMC\nQ04xFTATBgNVBAkMDOi9rOWhmOihl+mBkzEMMAoGA1UEChMDemN5MQswCQYDVQQL\nEwJjYTEbMBkGA1UEAxMSY2EuemN5Lmxhemllc3QuY29tMB4XDTIyMDMwNzAxNTY0\nNVoXDTIzMDMwNzAxNTY0NVowXjELMAkGA1UEBhMCQ04xFTATBgNVBAkMDOi9rOWh\nmOihl+mBkzEMMAoGA1UEChMDemN5MQwwCgYDVQQLEwN0bHMxHDAaBgNVBAMTE3Rs\ncy56Y3kubGF6aWVzdC5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARxXOin\nOF/JU7m1Z9zMr0Mq+A2Gdop8xLTpYSvmA/Y8EY1JH5T0dLsja+aLR7KbMbJgHDDk\nS9SvXOrKmRVXIRioo4GSMIGPMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggr\nBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYEFNmezA3S7/Ap1/9znISzkscUcMji\nMB8GA1UdIwQYMBaAFKXsthQexYJX/QHKBFLPu0hoHp+5MB4GA1UdEQQXMBWCE3Rs\ncy56Y3kubGF6aWVzdC5jb20wCgYIKoZIzj0EAwIDRwAwRAIgF14+VeRUkwZvN/hs\nDO/ZYvuALWNH07FqCSz9OrOKdzkCIF2XRXqL+FgsWnBLr4TBMb5WklMe8IrEDM8S\n01ZEZtV6\n-----END CERTIFICATE-----"
var zcyTlsPrivateKey = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIEx0AIevciM0wThfXns85ibC037gmVPn5wkMrqpkVG16oAoGCCqGSM49\nAwEHoUQDQgAEcVzopzhfyVO5tWfczK9DKvgNhnaKfMS06WEr5gP2PBGNSR+U9HS7\nI2vmi0eymzGyYBww5EvUr1zqypkVVyEYqA==\n-----END EC PRIVATE KEY-----"
var zcyRootCaCertificate = "-----BEGIN CERTIFICATE-----\nMIIB/jCCAaSgAwIBAgIIPBTBNG5/4mwwCgYIKoZIzj0EAwIwVTELMAkGA1UEBhMC\nQ04xDzANBgNVBAkMBui9rOWhmDEMMAoGA1UEChMDemN5MQ0wCwYDVQQLEwRyb290\nMRgwFgYDVQQDEw96Y3kubGF6aWVzdC5jb20wHhcNMjIwMzA3MDE1MzQ1WhcNMjMw\nMzA3MDE1MzQ1WjBVMQswCQYDVQQGEwJDTjEPMA0GA1UECQwG6L2s5aGYMQwwCgYD\nVQQKEwN6Y3kxDTALBgNVBAsTBHJvb3QxGDAWBgNVBAMTD3pjeS5sYXppZXN0LmNv\nbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABP0Krn+QF9qyh72cDSuNSUI4gIj8\nT00gY+IVqcfyG22rdcVZgVppMZWik1TOQza3CR8z+XGnPLCJNYy2U8NbkXWjXjBc\nMA4GA1UdDwEB/wQEAwIBhjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQG1K03\nX4hFTS7hv8ojubgWwvpD9zAaBgNVHREEEzARgg96Y3kubGF6aWVzdC5jb20wCgYI\nKoZIzj0EAwIDSAAwRQIgasp5FqXrVMHBSW2ub4UcCTd9R8u5uNIxta2K/csawrAC\nIQDCjVqwBMpdjm/+Q+8TYeDsvSI6Gbn+gri0lKZ1Y3WD0g==\n-----END CERTIFICATE-----"
var zcyIntermediateCaCertificate = "-----BEGIN CERTIFICATE-----\nMIICLDCCAdGgAwIBAgIILzkxq6twACMwCgYIKoZIzj0EAwIwVTELMAkGA1UEBhMC\nQ04xDzANBgNVBAkMBui9rOWhmDEMMAoGA1UEChMDemN5MQ0wCwYDVQQLEwRyb290\nMRgwFgYDVQQDEw96Y3kubGF6aWVzdC5jb20wHhcNMjIwMzA3MDE1NTA2WhcNMjMw\nMzA3MDE1NTA2WjBcMQswCQYDVQQGEwJDTjEVMBMGA1UECQwM6L2s5aGY6KGX6YGT\nMQwwCgYDVQQKEwN6Y3kxCzAJBgNVBAsTAmNhMRswGQYDVQQDExJjYS56Y3kubGF6\naWVzdC5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATfEdk7lIrxE7Ss1qhy\ngNcDMof1zUza1085sNTSLZjKGg72PmlE3rTlsjzGCXzHJSS6Iip0sQTF/xxzqm8Q\nQUvZo4GDMIGAMA4GA1UdDwEB/wQEAwIBhjAPBgNVHRMBAf8EBTADAQH/MB0GA1Ud\nDgQWBBSl7LYUHsWCV/0BygRSz7tIaB6fuTAfBgNVHSMEGDAWgBQG1K03X4hFTS7h\nv8ojubgWwvpD9zAdBgNVHREEFjAUghJjYS56Y3kubGF6aWVzdC5jb20wCgYIKoZI\nzj0EAwIDSQAwRgIhAKBL3WApAxtgrEPWdPw8/9L255JrLYovFa5vnNAvKXxyAiEA\nrlZT6OCp2o14wUQvgivfjrSLatUX4XlkVPt21Q4NXKw=\n-----END CERTIFICATE-----"

var zctAdminCertificate = "-----BEGIN CERTIFICATE-----\nMIICSDCCAe+gAwIBAgIIZ+NsNTD2S38wCgYIKoZIzj0EAwIwXDELMAkGA1UEBhMC\nQ04xFTATBgNVBAkMDOiJr+a4muihl+mBkzEMMAoGA1UEChMDemN0MQswCQYDVQQL\nEwJjYTEbMBkGA1UEAxMSY2EuemN0Lmxhemllc3QuY29tMB4XDTIyMDMwNzAzMDMx\nNFoXDTIzMDMwNzAzMDMxNFowYjELMAkGA1UEBhMCQ04xFTATBgNVBAkMDOiJr+a4\nmuihl+mBkzEMMAoGA1UEChMDemN0MQ4wDAYDVQQLEwVhZG1pbjEeMBwGA1UEAxMV\nYWRtaW4uemN0Lmxhemllc3QuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\nUHkmBtV1u0tDQRC8WfpjPIKoW0UzRuep0x7y+XMCAlS5ajFkjjfYhdWM1ch3wNJ6\nSdbHR6KS4EQjO0xleImdHKOBlDCBkTAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYw\nFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBSKTK/+QiyCq0UnK2YsyPrm\nxNJlETAfBgNVHSMEGDAWgBSf8RMMRpD1eFJEBu+BhP8+ZzGfOzAgBgNVHREEGTAX\nghVhZG1pbi56Y3QubGF6aWVzdC5jb20wCgYIKoZIzj0EAwIDRwAwRAIgGQOZJNdt\nPuelZizE6XJD/ZMr9oSZDug4u8lcILhC5VgCIFzWqtvE7jzfEpCejEWdU7ffQvRP\nBMfmaqORw+0hQZpZ\n-----END CERTIFICATE-----"
var zctAdminPrivateKey = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIOworZaB7Y/z1pgLOdWdJCmI12fXRGhoI5mjd016W75FoAoGCCqGSM49\nAwEHoUQDQgAEUHkmBtV1u0tDQRC8WfpjPIKoW0UzRuep0x7y+XMCAlS5ajFkjjfY\nhdWM1ch3wNJ6SdbHR6KS4EQjO0xleImdHA==\n-----END EC PRIVATE KEY-----"
var zctTlsCertificate = "-----BEGIN CERTIFICATE-----\nMIICQjCCAemgAwIBAgIIMR8n4dLGS24wCgYIKoZIzj0EAwIwXDELMAkGA1UEBhMC\nQ04xFTATBgNVBAkMDOiJr+a4muihl+mBkzEMMAoGA1UEChMDemN0MQswCQYDVQQL\nEwJjYTEbMBkGA1UEAxMSY2EuemN0Lmxhemllc3QuY29tMB4XDTIyMDMwNzAzMDIx\nMloXDTIzMDMwNzAzMDIxMlowXjELMAkGA1UEBhMCQ04xFTATBgNVBAkMDOiJr+a4\nmuihl+mBkzEMMAoGA1UEChMDemN0MQwwCgYDVQQLEwN0bHMxHDAaBgNVBAMTE3Rs\ncy56Y3QubGF6aWVzdC5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAQSiY3i\naILZtF6RBB5lZ0BiQYHcCAQLwCrvNibj30BFgtG/09Haa05yQCXWfG41yiJL/XSp\nK1PmigdqpibFiAQ7o4GSMIGPMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggr\nBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYEFNZ62/LljS8aDmlRaVXRRFTcP2eC\nMB8GA1UdIwQYMBaAFJ/xEwxGkPV4UkQG74GE/z5nMZ87MB4GA1UdEQQXMBWCE3Rs\ncy56Y3QubGF6aWVzdC5jb20wCgYIKoZIzj0EAwIDRwAwRAIgI8OYwd0lLVaXsFBE\naRJIu4BG39iBxUHxPdzfbdf/jE0CIDWotVgRhGAHA87z3qlw5olRDvXG1pi7JTgm\nEdBdV96J\n-----END CERTIFICATE-----"
var zctRootCaCertificate = "-----BEGIN CERTIFICATE-----\nMIICCTCCAbCgAwIBAgIIcT52i3qolR8wCgYIKoZIzj0EAwIwWzELMAkGA1UEBhMC\nQ04xFTATBgNVBAkMDOiJr+a4muihl+mBkzEMMAoGA1UEChMDemN0MQ0wCwYDVQQL\nEwRyb290MRgwFgYDVQQDEw96Y3QubGF6aWVzdC5jb20wHhcNMjIwMzA3MDI1OTA1\nWhcNMjMwMzA3MDI1OTA1WjBbMQswCQYDVQQGEwJDTjEVMBMGA1UECQwM6Imv5ria\n6KGX6YGTMQwwCgYDVQQKEwN6Y3QxDTALBgNVBAsTBHJvb3QxGDAWBgNVBAMTD3pj\ndC5sYXppZXN0LmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABDRuArurJ7A8\nysQkHleQevX4hzJdqd5kIpLWbAyK/Znc+IF9+gMqb/lx4Ufido7Xzh3/u1YymEjR\nCI+76xeo9OOjXjBcMA4GA1UdDwEB/wQEAwIBhjAPBgNVHRMBAf8EBTADAQH/MB0G\nA1UdDgQWBBRgyyQL2hHlFS1D6cq9bT+XXxhoqTAaBgNVHREEEzARgg96Y3QubGF6\naWVzdC5jb20wCgYIKoZIzj0EAwIDRwAwRAIgb6FKvpbuhsG88EpNhDMSN72jfkeA\nw+twvPX79wZBQeACIFqBUPjvGKWA2+nerR7Yz3dkRAOKh7L/oGK0c1NoyweU\n-----END CERTIFICATE-----"
var zctIntermediateCaCertificate = "-----BEGIN CERTIFICATE-----\nMIICMDCCAdegAwIBAgIIOcDMwI19AxAwCgYIKoZIzj0EAwIwWzELMAkGA1UEBhMC\nQ04xFTATBgNVBAkMDOiJr+a4muihl+mBkzEMMAoGA1UEChMDemN0MQ0wCwYDVQQL\nEwRyb290MRgwFgYDVQQDEw96Y3QubGF6aWVzdC5jb20wHhcNMjIwMzA3MDMwMDEz\nWhcNMjMwMzA3MDMwMDEzWjBcMQswCQYDVQQGEwJDTjEVMBMGA1UECQwM6Imv5ria\n6KGX6YGTMQwwCgYDVQQKEwN6Y3QxCzAJBgNVBAsTAmNhMRswGQYDVQQDExJjYS56\nY3QubGF6aWVzdC5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASknL/DQjoa\nuMzR8uBSUd353f/djOSHJTeafsTVlf2BrzHrWfoxZp2BbaaP8BXSOeVmKkPdcNUf\n+05EI1uGZImwo4GDMIGAMA4GA1UdDwEB/wQEAwIBhjAPBgNVHRMBAf8EBTADAQH/\nMB0GA1UdDgQWBBSf8RMMRpD1eFJEBu+BhP8+ZzGfOzAfBgNVHSMEGDAWgBRgyyQL\n2hHlFS1D6cq9bT+XXxhoqTAdBgNVHREEFjAUghJjYS56Y3QubGF6aWVzdC5jb20w\nCgYIKoZIzj0EAwIDRwAwRAIgBTvHazeY3OqTas7fmdIZOyG8fdw8oP+bATXLEvgY\nHs4CIBebLngZnDTKS7ZcWBM/fiX6vOHZakhwqUFg84pj3xKZ\n-----END CERTIFICATE-----"

var zcmAdminCertificate = "-----BEGIN CERTIFICATE-----\nMIICGzCCAcGgAwIBAgIIdXvbLj+z5oAwCgYIKoZIzj0EAwIwRTELMAkGA1UEBhMC\nQ04xDDAKBgNVBAoTA3pjbTELMAkGA1UECxMCY2ExGzAZBgNVBAMTEmNhLnpjbS5s\nYXppZXN0LmNvbTAeFw0yMjAzMjIwMjAxMzJaFw0yMzAzMjIwMjAxMzJaMEsxCzAJ\nBgNVBAYTAkNOMQwwCgYDVQQKEwN6Y20xDjAMBgNVBAsTBWFkbWluMR4wHAYDVQQD\nExVhZG1pbi56Y20ubGF6aWVzdC5jb20wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\nAAQ0qLSs2E9cJ4L1VlZlPYx8g8lvlzC2iAwc4r7bV2rFLsEgLEJrpAldkb/jdkzd\nZZM6ANJ4RfxnPKok+SjHkDrTo4GUMIGRMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUE\nFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYEFDOCVHPokEhJWUlv4a4n\nWxGtK39/MB8GA1UdIwQYMBaAFGZfZ7nufLbEP5052wF1rsTEcjX2MCAGA1UdEQQZ\nMBeCFWFkbWluLnpjbS5sYXppZXN0LmNvbTAKBggqhkjOPQQDAgNIADBFAiEA1gKP\n57r0+8Cw/MWCr+MLVc1QYRCgXednXH34XHcZ0zECIBAl4hAQ6Mvo8haGvBfLf1gf\nWZqHpA+eQ5suoizHrP36\n-----END CERTIFICATE-----"
var zcmAdminPrivateKey = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIARngXIFuk/JXhKr/wHb1Y/nH6nhwnpb5hiDZ4SG7OssoAoGCCqGSM49\nAwEHoUQDQgAENKi0rNhPXCeC9VZWZT2MfIPJb5cwtogMHOK+21dqxS7BICxCa6QJ\nXZG/43ZM3WWTOgDSeEX8ZzyqJPkox5A60w==\n-----END EC PRIVATE KEY-----"
var zcmPeerCertificate = "-----BEGIN CERTIFICATE-----\nMIICGDCCAb6gAwIBAgIIR9BYzDR5Dz0wCgYIKoZIzj0EAwIwRTELMAkGA1UEBhMC\nQ04xDDAKBgNVBAoTA3pjbTELMAkGA1UECxMCY2ExGzAZBgNVBAMTEmNhLnpjbS5s\nYXppZXN0LmNvbTAeFw0yMjAzMjIwMjA3MTNaFw0yMzAzMjIwMjA3MTNaMEkxCzAJ\nBgNVBAYTAkNOMQwwCgYDVQQKEwN6Y20xDTALBgNVBAsTBHBlZXIxHTAbBgNVBAMT\nFHBlZXIuemNtLmxhemllc3QuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE\nFMCMhr+KaNYWWSp14oyRVYkFJ0jIOe8KD3AFczXbEEbBLGyryfwcj4rJ3jOfHTxb\ncADgLHUwoK5Fy0KzcN04B6OBkzCBkDAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYw\nFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBR4yl8dzvFPmQDaqOBW1/uL\nLvrRsTAfBgNVHSMEGDAWgBRmX2e57ny2xD+dOdsBda7ExHI19jAfBgNVHREEGDAW\nghRwZWVyLnpjbS5sYXppZXN0LmNvbTAKBggqhkjOPQQDAgNIADBFAiEAygrq3+cM\nTdh3fCiPdO1PYRGERqm4aPLUxlT3XZ8GS54CIHqsyLuwWeu8OTHlTwAzfmMnEK9x\n9y5JabjELbsgxa/d\n-----END CERTIFICATE-----"
var zcmPeerPrivateKey = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEIBPgIIQAt/t4+TdL3xox/eO3Q24WcGXq/y7k8e49S5UloAoGCCqGSM49\nAwEHoUQDQgAEFMCMhr+KaNYWWSp14oyRVYkFJ0jIOe8KD3AFczXbEEbBLGyryfwc\nj4rJ3jOfHTxbcADgLHUwoK5Fy0KzcN04Bw==\n-----END EC PRIVATE KEY-----"
var zcmRootCaCertificate = "-----BEGIN CERTIFICATE-----\nMIIB3DCCAYKgAwIBAgIIQXEL1ThahDcwCgYIKoZIzj0EAwIwRDELMAkGA1UEBhMC\nQ04xDDAKBgNVBAoTA3pjbTENMAsGA1UECxMEcm9vdDEYMBYGA1UEAxMPemNtLmxh\nemllc3QuY29tMB4XDTIyMDMyMjAxNTkxOVoXDTIzMDMyMjAxNTkxOVowRDELMAkG\nA1UEBhMCQ04xDDAKBgNVBAoTA3pjbTENMAsGA1UECxMEcm9vdDEYMBYGA1UEAxMP\nemNtLmxhemllc3QuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEpAPVDhF0\nY4bPPH8TBoKz14ALbUp43OvFLefXPqhiwhCiI3yurf/rzt6lqOwmyDj8F7+55Fj1\n+hVuXJQKWZwAcaNeMFwwDgYDVR0PAQH/BAQDAgGGMA8GA1UdEwEB/wQFMAMBAf8w\nHQYDVR0OBBYEFIrNazSOuMcySy2hUWgMsmFTMfz4MBoGA1UdEQQTMBGCD3pjbS5s\nYXppZXN0LmNvbTAKBggqhkjOPQQDAgNIADBFAiBNxpcKV2FMl6F9w+aNVPxYRQqB\n+twpfB0tgt78lDdVfQIhANIhg5JP0ZWc9lx3ROEwhmmexvQmMYx+cicj0zY375Q5\n-----END CERTIFICATE-----"
var zcmIntermediateCaCertificate = "-----BEGIN CERTIFICATE-----\nMIICIjCCAcigAwIBAgIIObR1Vb9AA8cwCgYIKoZIzj0EAwIwRDELMAkGA1UEBhMC\nQ04xDDAKBgNVBAoTA3pjbTENMAsGA1UECxMEcm9vdDEYMBYGA1UEAxMPemNtLmxh\nemllc3QuY29tMB4XDTIyMDMyMjAyMDA0OVoXDTIzMDMyMjAyMDA0OVowRTELMAkG\nA1UEBhMCQ04xDDAKBgNVBAoTA3pjbTELMAkGA1UECxMCY2ExGzAZBgNVBAMTEmNh\nLnpjbS5sYXppZXN0LmNvbTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABGEzKy22\nD/bJoFeKp29iMINFYe5n3HWe7bHZ1cMprRe8oBb1pEdMMMTBxIIpUYW1cQePNlqh\n8HSbd+KOcTZjmxmjgaIwgZ8wDgYDVR0PAQH/BAQDAgHuMB0GA1UdJQQWMBQGCCsG\nAQUFBwMBBggrBgEFBQcDAjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRmX2e5\n7ny2xD+dOdsBda7ExHI19jAfBgNVHSMEGDAWgBSKzWs0jrjHMkstoVFoDLJhUzH8\n+DAdBgNVHREEFjAUghJjYS56Y20ubGF6aWVzdC5jb20wCgYIKoZIzj0EAwIDSAAw\nRQIgX69W6hl6Ga5zLBdfWOntSRmlrIpAL++pR+TLgrwgEMICIQCsuCWdnJZsgPF6\nk2NWLB5mLUVZYM6aMrYENqNiG14F0A==\n-----END CERTIFICATE-----"

func TestCreateConsortium(t *testing.T) {

	zcyOrganization := GetZcyOrganization()

	zctOrganization := GetZctOrganization()

	consortium := client.DefaultConsortium("zcl")
	consortium.Organizations = []*client.Organization{zcyOrganization, zctOrganization}
	consortium.AddPolicy("Readers", "ANY Readers")
	consortium.AddPolicy("Writers", "ANY Writers")
	consortium.AddPolicy("Admins", "ANY Admins")

	ordererEndpoints := client.DefaultOrdererEndpoints()
	ordererEndpoints.AddOrganization(zcyOrganization, zctOrganization)
	ordererEndpoints.Orderers = []*client.OrdererEndpoint{{
		Host:          "tls.zcy.laziest.com",
		Port:          7050,
		ClientTlsCert: zcyTlsCertificate,
		ServerTlsCert: zcyTlsCertificate,
	}, {
		Host:          "tls.zct.laziest.com",
		Port:          9050,
		ClientTlsCert: zctTlsCertificate,
		ServerTlsCert: zctTlsCertificate,
	}}
	ordererEndpoints.AddPolicy(client.ReadersPolicyKey, "ANY Readers")
	ordererEndpoints.AddPolicy(client.WritersPolicyKey, "ANY Writers")
	ordererEndpoints.AddPolicy(client.AdminsPolicyKey, "ANY Admins")
	ordererEndpoints.AddPolicy(client.BlockValidationPolicyKey, "ANY BlockValidation")

	consortium.OrdererEndpoints = ordererEndpoints

	res, e := consortium.BuildConfigGroup()
	if e != nil {
		panic(e)
	}
	fmt.Println(res)

	block, e := consortium.GenesisBlock()
	if e != nil {
		panic(e)
	}

	e = writeFile("/Users/jianfengjin/workspace/gitee/fabric/fabric/cmd/orderer/system-genesis-block/genesis.block", client.ProtoMarshalIgnoreError(block), 0o640)
	if e != nil {
		panic(e)
	}
}

func TestPrintOrgConfig(t *testing.T) {
	organization := GetZcyOrganization()

	og, err := organization.BuildConfigGroup()
	if err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(og, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestQueryChannel(t *testing.T) {

	res, err := GetZctPeerClient().QueryChannels()
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestCreateChannelWithBlock(t *testing.T) {

	configReader, _ := os.Open("/Users/jianfengjin/workspace/fabric/workspaces/chanzcy/channel.tx")
	configTx, _ := ioutil.ReadAll(configReader)
	chConfig, _ := client.GetConfigUpdateFromEnvelope(configTx)
	res, err := GetOrdererClientForZcy().CreateChannelWithBlock("chan", chConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestCreateChannel(t *testing.T) {

	zcyOrganization := GetZcyOrganization()

	zctOrganization := GetZctOrganization()

	ch := client.DefaultChannel("zcl", "chan")
	ch.AddPolicy("Readers", "ANY Readers")
	ch.AddPolicy("Writers", "ANY Writers")
	ch.AddPolicy("Admins", "ANY Admins")

	application := client.DefaultApplication()
	application.AddPolicy(client.ReadersPolicyKey, "ANY Readers")
	application.AddPolicy(client.WritersPolicyKey, "ANY Writers")
	application.AddPolicy(client.AdminsPolicyKey, "ANY Admins")
	application.AddPolicy("LifecycleEndorsement", "ANY Writers")
	application.AddPolicy("Endorsement", "ANY Writers")
	application.Organizations = []*client.Organization{zcyOrganization, zctOrganization}

	ch.Application = application
	ch.Organizations = []*client.Organization{zcyOrganization, zctOrganization}

	res, err := GetOrdererClientForZcy().CreateChannel(ch.ChannelID, ch)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestOrgJoinChannel(t *testing.T) {

	zctOrganization := GetZctOrganization()

	res, err := GetOrdererClientForZcy().AddOrganizationalToChannel("chan", zctOrganization)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestOrgJoinChannel2(t *testing.T) {

	organization := GetZcmOrganization()

	res, err := GetOrdererClientForZcy().AddOrganizationalToChannel("kfc", organization)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestOrgExitChannel(t *testing.T) {

	res, err := GetOrdererClientForZcy().DeleteOrganizationalToChannel("kfc", "zcm")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestSetAnchorPeerWithBlock(t *testing.T) {
	configReader, _ := os.Open("/Users/jianfengjin/workspace/fabric/workspaces/chanzcy/channelAnchor.tx")
	configTx, _ := ioutil.ReadAll(configReader)
	chConfig, _ := client.GetConfigUpdateFromEnvelope(configTx)

	res, err := GetOrdererClientForZcy().SetAnchorPeerWithBlock("zcy", chConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestSetAnchorPeer(t *testing.T) {
	res, err := GetOrdererClientForZct().SetAnchorPeer("zct", "kfc", &client.AnchorPeer{
		Host: "tls.zct.laziest.com",
		Port: 9051,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestSetAnchorPeer2(t *testing.T) {
	res, err := GetOrdererClientForZcy().SetAnchorPeer("zcy", "kfc", &client.AnchorPeer{
		Host: "tls.zcy.laziest.com",
		Port: 7051,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestSetAnchorPeer3(t *testing.T) {
	res, err := GetOrdererClientForZcm().SetAnchorPeer("zcm", "kfc", &client.AnchorPeer{
		Host: "tls.zcm.laziest.com",
		Port: 8051,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func TestPeerJoinChannel(t *testing.T) {

	c := GetZctPeerClient()

	err := c.JoinChannel("kfc", GetOrdererClientForZct())

	if err != nil {
		panic(err)
	}
}

func TestPeerJoinChannel2(t *testing.T) {

	c := GetZcyPeerClient()

	err := c.JoinChannel("kfc", GetOrdererClientForZcy())

	if err != nil {
		panic(err)
	}
}

func TestPeerJoinChannel3(t *testing.T) {

	c := GetZcmPeerClient()

	err := c.JoinChannel("kfc", GetOrdererClientForZcy())

	if err != nil {
		panic(err)
	}
}

func TestGetGenesisBlock(t *testing.T) {
	block, err := GetOrdererClientForZcy().GenesisBlock("channel")
	if err != nil {
		panic(err)
	}
	fmt.Println(block.Header)
}

func TestGetConfigBlock(t *testing.T) {

	block, err := GetOrdererClientForZcy().GetConfigBlock("chan1")
	if err != nil {
		panic(err)
	}
	c, err := client.GetConfigUpdateFromEnvelope(block.Data.Data[0])
	if err != nil {
		panic(err)
	}
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestDeleteOrgFromConsortium(t *testing.T) {
	res, e := GetOrdererClientForZcy().DeleteOrganizationalFromConsortium("zcl", "zcm")
	if e != nil {
		panic(e)
	}
	fmt.Println(res)
}

func TestAddOrgToConsortium(t *testing.T) {

	organization := GetZcmOrganization()

	res, e := GetOrdererClientForZcy().AddOrganizationalToConsortium("zcl", organization)
	if e != nil {
		panic(e)
	}
	fmt.Println(res)
}

func TestInstallScc(t *testing.T) {

	mapV := make(map[string]interface{})
	mapV["address"] = "10.201.109.131:9998"
	mapV["dial_timeout"] = "10s"
	mapV["tls_required"] = false
	mapV["client_auth_required"] = false

	desc := &lifecycle.Descriptor{
		//Path:  "/Users/jianfengjin/workspace/fabric/workspaces/chaincode/chaincode-pkg/connection.json",
		Type:  peer.ChaincodeSpec_EXTERNAL,
		Label: "scc",
		Value: mapV,
	}
	ccPkg, err := lifecycle.NewCCPackage(desc)

	packageID := lifecycle.ComputePackageID("scc", ccPkg)
	fmt.Println(packageID)

	resp, err := GetZcyPeerClient().InstallChainCodePackage(ccPkg)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func TestGetInstalledCCPackage(t *testing.T) {

	packageID := "scc:d6756afe9a7e0a06d1b203addd8b9f0a711472cf6d9d10d1d8c5253c03b5df15"
	fmt.Println(packageID)

	resp, err := GetZcyPeerClient().GetInstalledChainCodePackageByID(packageID)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func TestGetInstalledCC(t *testing.T) {

	resp, err := GetZcyPeerClient().GetInstalledChainCodePackage()
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)

	b, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestApproveChainCode(t *testing.T) {

	request := &client.ApproveChaincodeRequest{
		Name:            "scc",
		Version:         "1",
		PackageID:       "scc:913bf4652697766a54cdb13fef398d8f0763f615f0139004104783186d9f7b73",
		Sequence:        2,
		SignaturePolicy: "AND('zcy.peer')",
		//EndorsementPlugin: "escc",
		//ValidationPlugin:  "vscc",
		//InitRequired:      true,
	}
	resp, err := GetZcyPeerClient().ApproveChainCode("kfc", request, GetOrdererClientForZcy())
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func TestQueryApprovedChaincodeDefinition(t *testing.T) {
	resp, err := GetZcyPeerClient().QueryApprovedChaincodeDefinition("kfc", "scc", 0)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	b, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestCheckCommitReadiness(t *testing.T) {
	resp, err := GetZcyPeerClient().CheckCommitReadiness("kfc", &client.CheckChaincodeCommitReadinessRequest{
		Name:            "scc",
		Sequence:        1,
		Version:         "1",
		SignaturePolicy: "AND('zcy.admin')",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)

	b, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestCommitChainCode(t *testing.T) {
	resp, err := GetPeersClient().CommitChainCode("kfc", &client.CommitChaincodeRequest{
		Name:            "scc",
		Sequence:        2,
		Version:         "1",
		SignaturePolicy: "AND('zcy.peer')",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func TestQueryCommitted(t *testing.T) {
	resp, err := GetZcyPeerClient().QueryCommitted("kfc", "scc")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	b, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestQueryCommittedForChannel(t *testing.T) {
	resp, err := GetZcyPeerClient().QueryCommittedOfChannel("kfc")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	b, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestInvokeChainCode(t *testing.T) {
	args := [][]byte{[]byte("ADD"), []byte("name"), []byte("jinjianfeng1")}
	resp, err := GetPeersClient().InvokeChainCode("kfc", "scc", false, args)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	b, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func TestQueryChainCode(t *testing.T) {
	args := [][]byte{[]byte("GET"), []byte("name")}
	resp, err := GetZcyPeerClient().QueryChainCode("kfc", "scc", false, args)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func TestQueryChainCode2(t *testing.T) {
	args := [][]byte{[]byte("HISTORY"), []byte("name")}
	resp, err := GetZcyPeerClient().QueryChainCode("kfc", "scc", false, args)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)

	var m []map[string]interface{}

	err = json.Unmarshal(resp.Payload, &m)
	if err != nil {
		panic(err)
	}

	for _, o := range m {
		v := o["value"].(string)
		fmt.Println(v)

		vv := o["valueo"].(string)
		vvv, _ := base64.StdEncoding.DecodeString(vv)
		var mm map[string]interface{}
		_ = json.Unmarshal(vvv, &mm)
		fmt.Println(mm)
	}

}

func GetZcyPeerClient() *client.PeerClient {

	zcyAdminUser, _ := client.GetUser("admin", "zcy", zcyAdminCertificate, zcyAdminPrivateKey)

	p := client.GetPeer("zcy", "tls.zcy.laziest.com", "tls.zcy.laziest.com:7051", zcyIntermediateCaCertificate)

	return &client.PeerClient{
		Peer:   p,
		Signer: zcyAdminUser,
	}
}

func GetZctPeerClient() *client.PeerClient {

	zctAdminUser, _ := client.GetUser("admin", "zct", zctAdminCertificate, zctAdminPrivateKey)

	p := client.GetPeer("zct", "tls.zct.laziest.com", "tls.zct.laziest.com:9051", zctIntermediateCaCertificate)

	return &client.PeerClient{
		Peer:   p,
		Signer: zctAdminUser,
	}
}

func GetZcmPeerClient() *client.PeerClient {

	zcmAdminUser, _ := client.GetUser("admin", "zcm", zcmAdminCertificate, zcmAdminPrivateKey)

	p := client.GetPeer("zcm", "tls.zcm.laziest.com", "tls.zcm.laziest.com:8051", zcmIntermediateCaCertificate)

	return &client.PeerClient{
		Peer:   p,
		Signer: zcmAdminUser,
	}
}

func GetPeersClient() *client.PeersClient {

	zctAdminUser, _ := client.GetUser("admin", "zct", zctAdminCertificate, zctAdminPrivateKey)
	zcyAdminUser, _ := client.GetUser("admin", "zcy", zcyAdminCertificate, zcyAdminPrivateKey)

	fmt.Println(zctAdminUser)
	fmt.Println(zcyAdminUser)

	zctPeer := client.GetPeer("zct", "tls.zct.laziest.com", "tls.zct.laziest.com:9051", zctIntermediateCaCertificate)
	zcyPeer := client.GetPeer("zcy", "tls.zcy.laziest.com", "tls.zcy.laziest.com:7051", zcyIntermediateCaCertificate)
	//zcmPeer := client.GetPeer("zcm", "tls.zcm.laziest.com", "tls.zcm.laziest.com:8051", zcmIntermediateCaCertificate)

	return &client.PeersClient{
		Peers:   []*endpoints.Peer{zcyPeer, zctPeer},
		Signer:  zcyAdminUser,
		Orderer: *GetOrdererClientForZcy(),
	}
}

func GetOrdererClientForZcy() *client.OrdererClient {

	zcyAdminUser, _ := client.GetUser("admin", "zcy", zcyAdminCertificate, zcyAdminPrivateKey)

	o := client.GetOrderer("tls.zcy.laziest.com", "tls.zcy.laziest.com:7050", zcyIntermediateCaCertificate)

	return &client.OrdererClient{
		Orderer: o,
		Signers: []client.Signer{zcyAdminUser},
		Signer:  zcyAdminUser,
	}
}

func GetOrdererClientForZct() *client.OrdererClient {

	zctAdminUser, _ := client.GetUser("admin", "zct", zctAdminCertificate, zctAdminPrivateKey)

	o := client.GetOrderer("tls.zct.laziest.com", "tls.zct.laziest.com:9050", zctIntermediateCaCertificate)

	return &client.OrdererClient{
		Orderer: o,
		Signers: []client.Signer{zctAdminUser},
		Signer:  zctAdminUser,
	}
}

func GetOrdererClientForZcm() *client.OrdererClient {

	zcmAdminUser, _ := client.GetUser("admin", "zcm", zcmAdminCertificate, zcmAdminPrivateKey)

	o := client.GetOrderer("tls.zcy.laziest.com", "tls.zcy.laziest.com:7050", zcyIntermediateCaCertificate)

	return &client.OrdererClient{
		Orderer: o,
		Signers: []client.Signer{zcmAdminUser},
		Signer:  zcmAdminUser,
	}
}

func GetZcyOrganization() *client.Organization {
	zcyConfig := &client.MspConfig{
		MspID:                "zcy",
		Cacerts:              []string{zcyRootCaCertificate},
		Intermediatecerts:    []string{zcyIntermediateCaCertificate},
		TlsCACerts:           []string{zcyRootCaCertificate},
		TlsIntermediateCerts: []string{zcyIntermediateCaCertificate},

		NodeOUs: &client.NodeOUs{
			Enable:                  true,
			ClientOUIdentifierCert:  zcyIntermediateCaCertificate,
			PeerOUIdentifierCert:    zcyIntermediateCaCertificate,
			AdminOUIdentifierCert:   zcyIntermediateCaCertificate,
			OrdererOUIdentifierCert: zcyIntermediateCaCertificate,
		},
	}

	zcyOrganization := client.DefaultOrganization("zcy")
	zcyOrganization.MspConfig = zcyConfig
	zcyOrganization.OrdererEndpoints = []string{"tls.zcy.laziest.com:7050"}
	zcyOrganization.AnchorPeers = []*client.AnchorPeer{{
		Host: "tls.zcy.laziest.com",
		Port: 7051,
	}}

	return zcyOrganization
}

func GetZctOrganization() *client.Organization {
	zctConfig := &client.MspConfig{
		MspID:                "zct",
		Cacerts:              []string{zctRootCaCertificate},
		Intermediatecerts:    []string{zctIntermediateCaCertificate},
		TlsCACerts:           []string{zctRootCaCertificate},
		TlsIntermediateCerts: []string{zctIntermediateCaCertificate},

		NodeOUs: &client.NodeOUs{
			Enable:                  true,
			ClientOUIdentifierCert:  zctIntermediateCaCertificate,
			PeerOUIdentifierCert:    zctIntermediateCaCertificate,
			AdminOUIdentifierCert:   zctIntermediateCaCertificate,
			OrdererOUIdentifierCert: zctIntermediateCaCertificate,
		},
	}

	zctOrganization := client.DefaultOrganization("zct")
	zctOrganization.MspConfig = zctConfig
	zctOrganization.OrdererEndpoints = []string{"tls.zct.laziest.com:9050"}
	zctOrganization.AnchorPeers = []*client.AnchorPeer{{
		Host: "tls.zct.laziest.com",
		Port: 9051,
	}}

	return zctOrganization
}

func GetZcmOrganization() *client.Organization {
	mspID := "zcm"
	config := &client.MspConfig{
		MspID:                mspID,
		Cacerts:              []string{zcmRootCaCertificate},
		Intermediatecerts:    []string{zcmIntermediateCaCertificate},
		TlsCACerts:           []string{zcmRootCaCertificate},
		TlsIntermediateCerts: []string{zcmIntermediateCaCertificate},

		NodeOUs: &client.NodeOUs{
			Enable:                  true,
			ClientOUIdentifierCert:  zcmIntermediateCaCertificate,
			PeerOUIdentifierCert:    zcmIntermediateCaCertificate,
			AdminOUIdentifierCert:   zcmIntermediateCaCertificate,
			OrdererOUIdentifierCert: zcmIntermediateCaCertificate,
		},
	}

	organization := client.DefaultOrganization(mspID)
	organization.MspConfig = config
	//organization.OrdererEndpoints = []string{"tls.zcm.laziest.com:7070"}

	return organization
}

// jsonToMap allocates a map[string]interface{}, unmarshals a JSON document into it
// and returns it, or error
func jsonToMap(marshaled []byte) (map[string]interface{}, error) {
	tree := make(map[string]interface{})
	d := json.NewDecoder(bytes.NewReader(marshaled))
	d.UseNumber()
	err := d.Decode(&tree)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling intermediate JSON: %s", err)
	}
	return tree, nil
}

func writeFile(filename string, data []byte, perm os.FileMode) error {
	dirPath := filepath.Dir(filename)
	exists, err := dirExists(dirPath)
	if err != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(dirPath, 0o750)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filename, data, perm)
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
