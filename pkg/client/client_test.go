package client

import (
	"encoding/base64"
	"io/ioutil"

	"testing"

	"k8s.io/kubernetes/pkg/api"
)

var (
	fake_namespace string = "default"
)

func TestService_list(t *testing.T) {
	result, err := NamespaceScopedServices(fake_cf, fake_namespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v services found: %+v", len(result.Items), result)
}

func TestListRC(t *testing.T) {
	var ca, crt, key []byte = make([]byte, 2000), make([]byte, 2000), make([]byte, 2000)
	l1, _ := base64.StdEncoding.Decode(ca, []byte(`
LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURHakNDQWdLZ0F3SUJBZ0lKQUxuV3hZRmty
VnQrTUEwR0NTcUdTSWIzRFFFQkJRVUFNQkl4RURBT0JnTlYKQkFNVEIydDFZbVV0WTJFd0hoY05N
VFl3TVRJNU1UZzFOVEE1V2hjTk5ETXdOakUyTVRnMU5UQTVXakFTTVJBdwpEZ1lEVlFRREV3ZHJk
V0psTFdOaE1JSUJJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBUThBTUlJQkNnS0NBUUVBCjlVY2FP
Qmt6U0t5Q1VWY2pOTnQ5Z0pVSWc4REhMSmNVNVhPakY0d2lBb3p3Z25tdTF1RzRiYzNxRm9GcXo0
VC8KNFpBK1BlaFBiSmNBc2JHZkNpdHJ3VjJxNDlKWHR2Z3ZaT1E3U28rTGJSTGlwZnZXVHc2MDI5
TVNadS9lV3RIKwoyK3BNcVBhZFFjZnFIQTgzQkJLbm9FWFd3c3N2VEpld3FOVlY1RVhEY3YvREhr
SFlyMEp2SUJIWGFoQkhnTWFrClZoTnV3ME9OZ1drQVhHRk5senduOU1sZy9Jd0t1V2d2aE8zaklT
Y3ZnMG8wMmlselM3NmNhSEdvdEJranI5K0wKYWNmSkhqbHhKYi9qWkZPTEJLTUR1SkJ1ay9SSmRv
b0NwbXFaVm1WSEhlMFlzaWdXeFF2ZFZKQ3MzUEYvQXlmMAo0cDBHaDY2c2tHY2c1R05yaE5uVVZ3
SURBUUFCbzNNd2NUQWRCZ05WSFE0RUZnUVVoTVVNMWh3SUhCYVVGcVVpClRTc3JGWHhLM2FJd1Fn
WURWUjBqQkRzd09ZQVVoTVVNMWh3SUhCYVVGcVVpVFNzckZYeEszYUtoRnFRVU1CSXgKRURBT0Jn
TlZCQU1UQjJ0MVltVXRZMkdDQ1FDNTFzV0JaSzFiZmpBTUJnTlZIUk1FQlRBREFRSC9NQTBHQ1Nx
RwpTSWIzRFFFQkJRVUFBNElCQVFEbTVaYXlicjVNQndNbmdiV1YxWUdVNnRwVjllVzVnVkJabTBV
QjV6RUhIN0xZCkt1cDVJNTJtUmF6WCtMSnhsL2YvSUt2SnFReXFyRHBXOCtOL0tJQ1dodW5wTmt5
SDJ0M3hJNk9PWTdIMHBUWU4KdFIrY0JHN25MY2x0SnhZRjdVaGlSTnRBeWYwMjBLOVpyRTNMbmNB
TW54TUppYlhQaEE4MVdNd0JzR3ludlQ3ZgpSSVBGeXUxSi9DdlU5dEJkOC9LTXVBUzZzNWRsR1Ix
ZWlLTGJuNzBHQzhDc0dmTnNKaDhBcnJ2NnBlVVFFTjhnClIxbzhwNUlVM0IwWEdhZld1TG1IU3VS
UnM0ZU1yV0FJVzZXWkc4bWduR0tWRklBZTlNR3ZRVXJMeG11NFRwMk4Kc1k3QXdYUzVQSjJvdks1
WXkrelYvU2Z1ZFVCbFVMdlByQk9BZmVzYgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
`))
	l2, _ := base64.StdEncoding.Decode(crt, []byte(`
LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM3ekNDQWRlZ0F3SUJBZ0lKQUl0UEgyR3BL
UjV1TUEwR0NTcUdTSWIzRFFFQkJRVUFNQkl4RURBT0JnTlYKQkFNVEIydDFZbVV0WTJFd0hoY05N
VFl3TVRJNU1UZzFOVEV3V2hjTk1UY3dNVEk0TVRnMU5URXdXakFWTVJNdwpFUVlEVlFRRERBcHJk
V0psTFdGa2JXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDCkFRRUFy
OGV1dkFsYm1oSDJJdkJLaDdoSnduUU9qY0JpZ000TW5tQ2ZSZ0FUajZXVERzMVU1WVRnT1lpb0Zh
N08KeW51RmpLU1E2QmJtNmEwblBsamJXcFg2YmFTR1BETVd5THN4WEYwSTZQU2pTeW15bkg0RlZz
QWg2cGpPUjl3RgpQQnRHN3pGcnBtKzloV0lDZExJWFZuMGpORSthNXp5bG1udGZ6azhwR3krUzhx
YVBOQTh1aDNlV0Rrald6bVZkCjlTenFaSU1Xd01uMFJMYU1qZVNHQitKeUVFWjJmY1drUUJBSlZu
RGhaTXpOVHZOWGVDLzdDRjRuZ3kyOUdOdWYKNnc2QlB4cm1kWTZvcmE2NW9PMDcwcVdld0dPYkwz
b3c4aWtwZFdPaUZlalZUaEFicTRWMHdKR2xLcGJVaEJrYQp2MkdoMXJVQlQ2RTBZN3poUHd3VHNy
UDA1d0lEQVFBQm8wVXdRekFKQmdOVkhSTUVBakFBTUFzR0ExVWREd1FFCkF3SUY0REFwQmdOVkhS
RUVJakFnZ2dwcmRXSmxjbTVsZEdWemdoSnJkV0psY201bGRHVnpMbVJsWm1GMWJIUXcKRFFZSktv
WklodmNOQVFFRkJRQURnZ0VCQVBQSGhSajZCSkhRZURjMEVtVVNnc00wbG96biszQjl2T1RhK0Zp
cwpoL3lRNktKdWFUZW93Y1psdkw1L1orSmdab3hFWllBaEphVytUcHRjdnB2YkIreUhObXVLZ1RZ
WDZnRWE5RTdBClA3OGdNK2xoUnpybkxZWEc5Q3pTSS9VTWNCaU9UOXRWTE9vMWcrWW5BRVZPQUdq
TEZWelMxTytPQUcwTnp1cEgKWmkrckxCTVYvb0hHRU5uaXFjTnFhK25IaTdlTjFDSDFHWFJ6dFho
QWp0dmFjOHhQVGFFRE9oOHlOR2dlN2NXdApyc1VNemVjM0w5WXNQd1JWOE4rZlpPdEdwWWtEVGJB
QlNsRHRaRmZONXVHcEJFeUYvREZUV2lZbVFTTlJuY2xyCkRyZ1E4STFzRXQ2TVNzYWxvSGxxcjVO
cGVBWlhVbkg2cjRYdFROZ1NBN085QnFJPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
`))
	l3, _ := base64.StdEncoding.Decode(key, []byte(`
LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBcjhldXZBbGJt
aEgySXZCS2g3aEp3blFPamNCaWdNNE1ubUNmUmdBVGo2V1REczFVCjVZVGdPWWlvRmE3T3ludUZq
S1NRNkJibTZhMG5QbGpiV3BYNmJhU0dQRE1XeUxzeFhGMEk2UFNqU3lteW5INEYKVnNBaDZwak9S
OXdGUEJ0Rzd6RnJwbSs5aFdJQ2RMSVhWbjBqTkUrYTV6eWxtbnRmems4cEd5K1M4cWFQTkE4dQpo
M2VXRGtqV3ptVmQ5U3pxWklNV3dNbjBSTGFNamVTR0IrSnlFRVoyZmNXa1FCQUpWbkRoWk16TlR2
TlhlQy83CkNGNG5neTI5R051ZjZ3NkJQeHJtZFk2b3JhNjVvTzA3MHFXZXdHT2JMM293OGlrcGRX
T2lGZWpWVGhBYnE0VjAKd0pHbEtwYlVoQmthdjJHaDFyVUJUNkUwWTd6aFB3d1RzclAwNXdJREFR
QUJBb0lCQVFDV1d4MXhwaWQrZ0VLagozQnBFUVRTR2FqTlF1UEVJOERjRytlc0RPMm1BQThib2NH
WmY5T3EvQzgrS3pINEI0T3A1UlFMV0kxRGRMTWxXCm9UYndJejJVcjVoS3dnQ0pMdVduOWNSN1Qx
YW1Ja0t2YlhRdm9pVGd3RmdpUzJjRzVPWW1ENFhmVmRFczJJcjEKT3FpUHlIWm1kbmt2dGthSS8y
TVIyOUJ6ZXB3ZGhVTlp4bjFuUmlidzZRdVhZcnB6RER4TEVVcUh3Z0tSUGIyMworckxpWUpCcEh0
N3M3NWM3TGV4THBxY0dDMmhpQ1lEWDNUdk9qWDNnZWdUWndjenhadUVOWlVxVytTc3VySG9pCjF0
Snp6YVhpV050eExKS3owMS9MdnpaS1RMWG5ab1FhZkl6enhFWERhdkFnSDY4U0NZaWp2SGVvajlL
WUMzdmkKTENHSS9BZ1JBb0dCQU9Md2I0a2grMXRuTEhOT0hTbHB4TlZWVHh6ajhEREtDVVIwUjhN
bVZEUG1kSzJaUkpiRQpPcXV2cXMxZUVtdHZOZmtuQVFwSjZWaFRNeDFKOHFzTUwxemF0MHM1Q3o0
c0JHUEkveWhCQmo3bExDL1V3VUJNClNMcC96RUxkM3kxMGEwWHBMTjFEZ2FnRkFsQmJjeWk2VkRZ
K2lEaXErWG5Na0ZQZVdBeWF5YitiQW9HQkFNWksKSXU3cjNmSkVDQ2Zwc0dzR3UxRmpmejhHYnJw
U1kxZW56dTFVQkpTbUlOZmM5U01zQ2ZDbnl0ZkFkbjNvSzUwWQpCcldVSy83MGloRThwMmNtdStt
ZUV2VmtGSkF6Wk1vNWJ3c1NhWUhRZG96a1h5MTB6ODJoaW9vVE5SRFdpN1lXCnp3bEZYSVVObkdR
dGkycWkrLy96bUY1cXhTUnJlR29pTDJBbDZNS2xBb0dBSUlzR0U0NHg4MXVLUGthM1c5YjEKQ1Bx
Z3k2M01KZEx6SFVQbmZvNmlpSWJGdUpkQUJMYkRDeGUzMEpkcTRNa2oza1MvbTNBWjdEYVNIK01Z
ejNxQgoyRGp0Qy9aMExFZzNvTytUMTN2cm4xMVJ4dElsbkVqVUxScGllanhDNHN2TkRrdHZ5Wk1D
cHN1QkYzRGx4TE5qCm5CdS9nUkRUa2FuS0VKelQxNHJpMzMwQ2dZQjUyYnNIMlRmbGxYcnhsQUNP
ZEcrTnZ6ZkZzaDAvTUR2TjlOZ0wKTFpNM0NiT3JFeVFzL2ZZSnhnSzNmSlVVSUNVVS8xdTRINXE5
ck9aZlcwZC84dVNNMWsvT0xqY1l1Z3hZM054cApFR0ozbkhRTmRwVXFhTnIrQVNRU1gyVS91S2ZZ
T01IM2I0RkFYakhadWNjdnU0SmlNZjVUSHdlUXJ0NHJVbUNNCmxCOHA1UUtCZ0U2czNmeWhISHdQ
d2FHK2RCUkI3dGVJVDgvRFJCb2YzNnJTMC81OWVOR2N0QlJXRVlFYUhrTTYKNGJ4Nm9QTi80OEg4
d2daN2JEMTlCT3ZPK29EMWhIWW00WnBRK3VkVGhYVVM0S21nUU93ZGJ6Nkl6V0ZybldNSQpBZGtY
SDV5V2xqbkJpU1lnc1BVVFhvVXdXcmZlSnFQdzhmcE9jSTFoSmtSaGJZS2NPc0J4Ci0tLS0tRU5E
IFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
`))
	t.Log(l1, l2, l3)
	cc := K8sClientConfig{
		ClusterID: "test",
		NameSpace: "default",
		//Server:               "https://192.168.1.230:443",
		Server: "https://172.17.4.99",
		//CertificateAuthority: "/home/johnsenxu/ssl/ca.pem",
		CertificateAuthorityData: ca[:l1],
		//ClientCertificate:    "/home/johnsenxu/ssl/admin.pem",
		ClientCertificateData: crt[:l2],
		//ClientKey:            "/home/johnsenxu/ssl/admin-key.pem",
		ClientKeyData: key[:l3],
	}
	var err error
	var rc *api.ReplicationControllerList
	opts := api.ListOptions{}
	rc, err = cc.UnversionedClient().ReplicationControllers(cc.NameSpace).List(opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rc)

}

func testConf(t *testing.T) {
	conf := K8sClientConfig{
		ClusterID: "test",
		NameSpace: "default",
		Server:    "http://localhost:8080",
	}
	var err error
	var rc *api.ReplicationController
	conf.DeleteRc("fx3-nginx")
	rc, err = conf.CreateRc("/home/johnsenxu/test/fx3-nginx.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rc)
	rc, err = conf.GetRc("fx3-nginx")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rc)

	rc, err = conf.ScaleRc("fx3-nginx", 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rc)

}

func testg(t *testing.T) {
	data, _ := ioutil.ReadFile("/home/johnsenxu/test/fx3-nginx.yaml")
	conf := K8sClientConfig{
		ClusterID: "test",
		NameSpace: "default",
		Server:    "http://localhost:8080",
	}
	rc, err := conf.CreateRcByInput(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rc)

}
