package client

import (
	"testing"
	"time"
)

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

func TestFriendlyString(t *testing.T) {
	strs := []string{"2021-01-22T18-10-10.581Z", "2021-01-22T16-20-26.818Z", "2021-01-22T06-20-26.818Z"}
	for _, str := range strs {
		t.Log(friendlyTime(str))
	}
	t.Fail()
}

func TestTime(t *testing.T) {
	strs := []string{"2021-01-22T18-10-10.581Z", "2021-01-22T16-20-26.818Z"}
	for _, str := range strs {
		tx, err := time.Parse("2006-01-02T15-04-05.000Z", str)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(tx.String())
	}
	t.Fail()
}
