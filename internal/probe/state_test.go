package probe

import (
    "testing"
    "time"

    "github.com/wyw14/solo-037/internal/model"
)

func sample(run model.RunID, revision model.RevisionID, source string, at time.Time) model.Evidence {
    return model.NewEvidence(run, revision, source, model.EvidenceWaveTime, 0.01, at)
}

func TestInvalidateRebuildsSourceIndexFromRetainedEvidence(t *testing.T) {
    collection := NewCollection()
    run := model.NewRunID()
    rev1 := model.NewRevisionID()
    rev2 := model.NewRevisionID()
    rev3 := model.NewRevisionID()
    at := time.Unix(1700000000, 0)

    collection.Add(sample(run, rev1, "probe-a", at))
    collection.Add(sample(run, rev1, "probe-b", at.Add(time.Millisecond)))
    if got := collection.IndependentSources(run); got != 2 {
        t.Fatalf("initial source count = %d, want 2", got)
    }

    // Switch to rev2: only the rev2 probe must count, old probes are gone.
    collection.Add(sample(run, rev2, "probe-c", at.Add(2*time.Millisecond)))
    collection.Invalidate(run, rev2)
    if got := len(collection.Samples(run)); got != 1 {
        t.Fatalf("retained samples = %d, want 1", got)
    }
    if got := collection.IndependentSources(run); got != 1 {
        t.Fatalf("source count after first switch = %d, want 1", got)
    }
    if got := collection.Revision(run); got != rev2 {
        t.Fatalf("revision after switch = %s, want %s", got, rev2)
    }

    // Repeated invalidation against the same revision must stay stable.
    for i := 0; i < 3; i++ {
        collection.Invalidate(run, rev2)
        if got := collection.IndependentSources(run); got != 1 {
            t.Fatalf("source count drifted on repeat %d: got %d, want 1", i, got)
        }
        if got := len(collection.Samples(run)); got != 1 {
            t.Fatalf("sample count drifted on repeat %d: got %d, want 1", i, got)
        }
    }

    // Switch again, reusing a source id seen in the retired revision.
    collection.Add(sample(run, rev3, "probe-a", at.Add(3*time.Millisecond)))
    collection.Add(sample(run, rev3, "probe-d", at.Add(4*time.Millisecond)))
    collection.Invalidate(run, rev3)
    if got := collection.IndependentSources(run); got != 2 {
        t.Fatalf("source count after second switch = %d, want 2", got)
    }
    if got := len(collection.Samples(run)); got != 2 {
        t.Fatalf("retained samples after second switch = %d, want 2", got)
    }
    if got := collection.Revision(run); got != rev3 {
        t.Fatalf("revision after second switch = %s, want %s", got, rev3)
    }
    collection.Invalidate(run, rev3)
    if got := collection.IndependentSources(run); got != 2 {
        t.Fatalf("source count not idempotent after second switch: got %d, want 2", got)
    }
}

func TestInvalidateClearsRunWhenNoEvidenceRetained(t *testing.T) {
    collection := NewCollection()
    run := model.NewRunID()
    at := time.Unix(1700000000, 0)

    collection.Add(sample(run, model.NewRevisionID(), "probe-a", at))
    collection.Invalidate(run, model.NewRevisionID())

    if got := collection.IndependentSources(run); got != 0 {
        t.Fatalf("source count after clearing = %d, want 0", got)
    }
    if got := len(collection.Samples(run)); got != 0 {
        t.Fatalf("samples after clearing = %d, want 0", got)
    }
    if got := collection.Revision(run); got != "" {
        t.Fatalf("revision after clearing = %q, want empty", got)
    }

    // The run can collect evidence again after being cleared.
    rev := model.NewRevisionID()
    collection.Add(sample(run, rev, "probe-b", at.Add(time.Millisecond)))
    if got := collection.IndependentSources(run); got != 1 {
        t.Fatalf("source count after re-add = %d, want 1", got)
    }
    if got := collection.Revision(run); got != rev {
        t.Fatalf("revision after re-add = %s, want %s", got, rev)
    }
}
