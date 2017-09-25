package killable

import "time"

type kContext struct {
	k Killable
}

// Deadline always returns the time.Time zero value and false
func (c *kContext) Deadline() (time.Time, bool) {
	var t time.Time
	return t, false
}

// Done return the killable.Killable's Dying() channel
func (c *kContext) Done() <-chan struct{} {
	return c.k.Dying()
}

// Err returns an error if the killable is dying, otherwise it returns nill
func (c *kContext) Err() error {
	if c.k.isDying() {
		return c.k.Err()
	}
	return nil
}

// Value always returns nil
func (c *kContext) Value(key interface{}) interface{} {
	return nil
}
