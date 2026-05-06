package firestoresave_test

type fakeEntity struct {
	version int
}

func (e *fakeEntity) Version() int     { return e.version }
func (e *fakeEntity) SetVersion(v int) { e.version = v }

type fakeRecord struct {
	ID      string
	Version int
}

func (r *fakeRecord) GetVersion() int { return r.Version }
