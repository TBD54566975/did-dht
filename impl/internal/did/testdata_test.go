package did

import (
	"embed"
	"testing"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata
	testData embed.FS
)

const (
	vector1PublicKeyJWK1 string = "vector-1-public-key-jwk-1.json"
	vector1DIDDocument   string = "vector-1-did-document.json"
	vector1DNSRecords    string = "vector-1-dns-records.json"

	vector2PublicKeyJWK2 string = "vector-2-public-key-jwk-2.json"
	vector2DIDDocument   string = "vector-2-did-document.json"
	vector2DNSRecords    string = "vector-2-dns-records.json"

	vector3PublicKeyJWK1 string = "vector-3-public-key-jwk-1.json"
	vector3PublicKeyJWK2 string = "vector-3-public-key-jwk-2.json"
	vector3DIDDocument   string = "vector-3-did-document.json"
	vector3DNSRecords    string = "vector-3-dns-records.json"
)

func getTestData(fileName string) ([]byte, error) {
	return testData.ReadFile("testdata/" + fileName)
}

// retrieveTestVectorAs retrieves a test vector from the testdata folder and unmarshals it into the given interface
func retrieveTestVectorAs(t *testing.T, fileName string, output any) {
	t.Helper()
	testDataBytes, err := getTestData(fileName)
	require.NoError(t, err)
	err = json.Unmarshal(testDataBytes, output)
	require.NoError(t, err)
}
