package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/iron-io/functions/api/models"
	"net/url"
	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
)

const tmpBolt = "/tmp/func_test_bolt.db"

func TestDatastore(t *testing.T) {
	os.Remove(tmpBolt)
	u, err := url.Parse("bolt://" + tmpBolt)
	if err != nil {
		t.Fatalf("failed to parse url:", err)
	}
	ds, err := New(u)
	if err != nil {
		t.Fatalf("failed to create bolt datastore:", err)
	}
	datastoretest.Test(t, ds)
}
