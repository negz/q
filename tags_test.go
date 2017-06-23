package q

import (
	"reflect"
	"testing"
)

var tagsTests = []struct {
	tags map[string]string
}{
	{tags: map[string]string{"orbit": "heliosynchronous", "altitude": "low"}},
	{tags: map[string]string{"orbit": "geostationary", "altitude": "high"}},
	{tags: map[string]string{}},
	{tags: nil},
}

func TestTags(t *testing.T) {
	tags := &Tags{}

	// Ensure we can test membership of uninitialised tags.
	if tags.Contains("empty", "tag") {
		t.Error("tags.Contains(nil, tags): want false, got true")
	}

	// Ensure we can get uninitialised tags.
	if got := tags.Get(); got != nil {
		t.Errorf("tags.Get(): want nil, got %v", got)
	}

	// Ensure we can 'remove' from uninitialised tags.
	tags.Remove("nil", "tag")

	t.Run("Add", func(t *testing.T) {
		for _, tt := range tagsTests {
			for k, v := range tt.tags {
				// Add the tag twice to ensure we don't get duplicates.
				tags.Add(k, v)
				tags.Add(k, v)
			}
		}
	})

	t.Run("Contains", func(t *testing.T) {
		for _, tt := range tagsTests {
			for k, v := range tt.tags {
				if !tags.Contains(k, v) {
					t.Errorf("tags.Contains(%v, %v): want true, got false", k, v)
				}
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		want := make(map[Tag]bool)
		got := make(map[Tag]bool)

		for _, tt := range tagsTests {
			for k, v := range tt.tags {
				want[Tag{k, v}] = true
			}
		}

		for _, tag := range tags.Get() {
			got[tag] = true
		}

		// Compare the tags after converting them to 'sets'.
		// This allows to us to test that they have the same membership
		// without concern for their order.
		if !reflect.DeepEqual(want, got) {
			t.Errorf("tags.Get():\nwant %+#v\ngot%+#v", want, got)
		}
	})

	t.Run("Remove", func(t *testing.T) {
		for _, tt := range tagsTests {
			for k, v := range tt.tags {
				// Remove the tag twice to ensure we don't fail to remove non-existent tags.
				tags.Remove(k, v)
				tags.Remove(k, v)

				if tags.Contains(k, v) {
					t.Errorf("tags.Contains(%v, %v): want false, got true", k, v)
				}
			}
		}
	})

	// Create new tags to ensure AddMap can initialise them.
	tags = &Tags{}
	t.Run("AddMap", func(t *testing.T) {
		for _, tt := range tagsTests {
			tags.AddMap(tt.tags)

			for k, v := range tt.tags {
				if !tags.Contains(k, v) {
					t.Errorf("tags.Contains(%v, %v): want true, got false", k, v)
				}
			}
		}
	})
}
