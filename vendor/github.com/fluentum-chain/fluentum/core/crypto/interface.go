package crypto

// Signer defines the interface for cryptographic signing algorithms
// Features (e.g., quantum_signing) should implement this interface
// and register their signer at runtime.
type Signer interface {
	GenerateKey() ([]byte, []byte) // private, public
	Sign(privateKey []byte, message []byte) []byte
	Verify(publicKey []byte, message []byte, signature []byte) bool
	Name() string
}

// signerRegistry holds all registered signers by name
var signerRegistry = make(map[string]Signer)

// currentSigner is the active signer used by the node
var currentSigner Signer

// RegisterSigner registers a new signer implementation by name
func RegisterSigner(name string, s Signer) {
	signerRegistry[name] = s
}

// SetActiveSigner sets the active signer by name
func SetActiveSigner(name string) bool {
	s, ok := signerRegistry[name]
	if ok {
		currentSigner = s
	}
	return ok
}

// GetSigner returns the currently active signer
func GetSigner() Signer {
	return currentSigner
}

// ListSigners returns the names of all registered signers
func ListSigners() []string {
	names := make([]string, 0, len(signerRegistry))
	for name := range signerRegistry {
		names = append(names, name)
	}
	return names
}

// --- Default ECDSA Signer (example stub) ---
// In production, replace with actual ECDSA implementation

type ECDSASigner struct{}

func NewECDSASigner() *ECDSASigner { return &ECDSASigner{} }

func (e *ECDSASigner) GenerateKey() ([]byte, []byte)                                  { return []byte("priv"), []byte("pub") }
func (e *ECDSASigner) Sign(privateKey []byte, message []byte) []byte                  { return []byte("sig") }
func (e *ECDSASigner) Verify(publicKey []byte, message []byte, signature []byte) bool { return true }
func (e *ECDSASigner) Name() string                                                   { return "ecdsa" }

// --- Example: Register default signer at init ---
func init() {
	RegisterSigner("ecdsa", NewECDSASigner())
	currentSigner = signerRegistry["ecdsa"]
}

/*
USAGE FOR FEATURES (e.g., quantum_signing):

// Implement the Signer interface in your feature package
// Register your signer at runtime:
import "github.com/fluentum-chain/fluentum/core/crypto"

crypto.RegisterSigner("dilithium", NewDilithiumSigner())
crypto.SetActiveSigner("dilithium") // To activate
*/
