package types

// Common types used across ABCI
type ValidatorUpdate struct {
	PubKey PubKey
	Power  int64
}

type PubKey struct {
	Type string
	Data []byte
}

type Event struct {
	Type       string
	Attributes []EventAttribute
}

type EventAttribute struct {
	Key   string
	Value string
	Index bool
}

// Helper methods for EventAttribute
func (ea EventAttribute) String() string {
	return ea.Key + "=" + ea.Value
}

// Helper methods for Event
func (e Event) String() string {
	if len(e.Attributes) == 0 {
		return e.Type
	}
	
	result := e.Type + "("
	for i, attr := range e.Attributes {
		if i > 0 {
			result += ", "
		}
		result += attr.String()
	}
	result += ")"
	return result
} 