package dilithium

type Mode3 struct{}
type PrivateKey struct{}
type PublicKey struct{}

func GenerateKey() (PrivateKey, PublicKey) { return PrivateKey{}, PublicKey{} }

// Add other required functions
