package probe

import (
    "testing"
    "time"

    "github.com/wyw14/solo-037/internal/model"
)

func TestInvalidateRebuildsSourceIndex(t *testing.T) {
    collection := NewCollection()
    run := model.NewRunID()
    revisionA := model.NewRevisionID()
    revisionB := model.NewRevisionID()
    at := time.Now().UTC()

    collection.Add(model.NewEvidence(run, revisionA, "probe-old-1", model.EvidenceWaveTime, 0.1, at))
    collection.Add(model.NewEvidence(run, revisionA, "probe-old-2", model.EvidenceWaveTime, 0.2, at))
    collection.Add(model.NewEvidence(run, revisionB, "probe-new", model.EvidenceWaveTime, 0.3, at))

    collection.Invalidate(run, revisionB)

    samples := collection.Samples(run)
    if len(samples) != 1 || samples[0].SourceID != "probe-new" {
        t.Fatalf("retained samples mismatch: %+v", samples)
    }
    if got := collection.IndependentSources(run); got != 1 {
        t.Fatalf("stale sources still counted after revision switch: got %d, want 1", got)
    }
    if got := collection.Revision(run); got != revisionB {
        t.Fatalf("current revision not updated: got %v, want %v", got, revisionB)
    }
}

func TestInvalidateStableAcrossRepeatedSwitches(t *testing.T) {
    collection := NewCollection()
    run := model.NewRunID()
    at := time.Now().UTC()

    revisions := []model.RevisionID{model.NewRevisionID(), model.NewRevisionID(), model.NewRevisionID()}
    sources := []string{"probe-a", "probe-b", "probe-c"}
    for index, revision := range revisions {
        collection.Add(model.NewEvidence(run, revision, sources[index], model.EvidenceWaveTime, 0.1, at))
        collection.Invalidate(run, revision)
        if got := collection.IndependentSources(run); got != 1 {
            t.Fatalf("switch %d: independent sources = %d, want 1", index, got)
        }
        if got := len(collection.Samples(run)); got != 1 {
            t.Fatalf("switch %d: retained samples = %d, want 1", index, got)
        }
    }

    // Re-applying the current revision must be idempotent.
    collection.Invalidate(run, revisions[2])
    if got := collection.IndependentSources(run); got != 1 {
        t.Fatalf("repeated invalidate changed source count: %d", got)
    }

    // Switching to a revision with no retained evidence empties both views.
    collection.Invalidate(run, model.NewRevisionID())
    if got := collection.IndependentSources(run); got != 0 {
        t.Fatalf("unknown revision kept sources: %d", got)
    }
    if got := len(collection.Samples(run)); got != 0 {
        t.Fatalf("unknown revision kept samples: %d", got)
    }
}
