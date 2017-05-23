package amclient_test

import (
	"context"
	"fmt"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
)

func ExampleReverse() {
	c, _ := amclient.New(nil, "http://127.0.0.1:32804", "test", "test")
	resp, _ := c.Transfer.Approve(context.TODO(), &amclient.TransferApproveRequest{
		Type:      "standard",
		Directory: "/asdf",
	})
	fmt.Printf("%d", resp.StatusCode)
}
