package trie

// Trier exposes the Trie structure capabilities.
type Trier[T any] interface {
	Get(key string) T
	Put(key string, value T) bool
	Delete(key string) bool
	Walk(walker WalkFunc[T]) error
	WalkPath(key string, walker WalkFunc[T]) error
}
