package diagnose

import "time"

// Stance is the Conclusion's three-way claim. Values are the closed set
// holds, does_not_hold, insufficient_evidence.
type Stance string

const (
	StanceHolds        Stance = "holds"
	StanceDoesNotHold  Stance = "does_not_hold"
	StanceInsufficient Stance = "insufficient_evidence"
)

// Operator is a Predicate comparison. Values are the closed set lt, le, gt, ge, eq.
type Operator string

const (
	OpLT Operator = "lt"
	OpLE Operator = "le"
	OpGT Operator = "gt"
	OpGE Operator = "ge"
	OpEQ Operator = "eq"
)

// Window is one already-cut interval on one host. start and end use the same
// wall clock as Record.at; end is exclusive.
type Window struct {
	ID    string
	Host  string
	Start time.Time
	End   time.Time
}

// WindowCite is the Evidence form of a Window: identity and span, no host.
type WindowCite struct {
	ID    string
	Start time.Time
	End   time.Time
}

// Subject is the Linux-task identity a Conclusion is about.
type Subject struct {
	Host      string
	Pid       int32
	Tid       int32
	Starttime uint64
}

// Conclusion is a typed claim: kind, the Subject, and a stance. Prose is not stored.
type Conclusion struct {
	Kind    string
	Subject Subject
	Stance  Stance
}

// RecordSnap is the observation a citation used, not every Record in the Window.
type RecordSnap struct {
	Domain  string
	At      time.Time
	Payload any
}

// MissingSource is a required source that was absent.
type MissingSource struct {
	Source string
	Reason string
}

// Citation is one rests-on entry. Exactly one of Record, Window, EvidenceID, or
// Missing is set.
type Citation struct {
	Record     *RecordSnap
	Window     *WindowCite
	EvidenceID string
	Missing    *MissingSource
}

// Predicate is an optional comparison recorded on Evidence.
type Predicate struct {
	Observed float64
	Operator Operator
	Bound    float64
}

// Evidence is one item a Conclusion rests on. Kind is opaque and namespaced by
// collection domain; domain-free packages do not branch on it.
type Evidence struct {
	ID        string
	Kind      string
	Payload   any
	Predicate *Predicate
	Citations []Citation
}

// Report is the auditable document: one Conclusion, the Evidence it rests on,
// and the Window it covers. Immutable; a later run is a different Report.
type Report struct {
	ID         string
	ProducedAt time.Time
	Host       string
	Analyzer   string
	Window     Window
	Conclusion Conclusion
	Evidence   []Evidence
}
