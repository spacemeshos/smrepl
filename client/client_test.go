package client

import "testing"

func TestDefault(t *testing.T) {
	testClient(t, DefaultGRPCServer, DefaultSecureConnection)
}

func TestPublic(t *testing.T) {
	testClient(t, "api-21.spacemesh.io:443", true)
}

func testClient(t *testing.T, server string, secure bool) {
	client := newGRPCClient(server, secure)

	err := client.Connect()
	if err != nil {
		t.Fatal(err)
	}
	str := client.ServerInfo()
	t.Log(str)
	client.getNodeServiceClient()
	ni, err := client.NodeInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ni.Version, ni.Build)
	t.Fail()
}
