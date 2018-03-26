package consumer

import (
	"context"
	"testing"
)

func TestStorageInMemoryImpl(t *testing.T) {
	ctx := context.Background()
	s := newStorageInMemory()

	id, _ := s.GetResearchObject(ctx, "foo")
	if have, want := id, ""; have != want {
		t.Fatalf("GetResearchObject(); have `%s` want `%s`", have, want)
	}

	_ = s.AssociateResearchObject(ctx, "foo", "bar")

	id, _ = s.GetResearchObject(ctx, "foo")
	if have, want := id, "bar"; have != want {
		t.Fatalf("GetResearchObject(); have `%s` want `%s`", have, want)
	}
}
