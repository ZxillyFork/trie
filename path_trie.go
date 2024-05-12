package trie

// PathTrie is a trie of paths with string keys and interface{} values.

// PathTrie is a trie of string keys and interface{} values. Internal nodes
// have nil values so stored nil values cannot be distinguished and are
// excluded from walks. By default, PathTrie will segment keys by forward
// slashes with PathSegmenter (e.g. "/a/b/c" -> "/a", "/b", "/c"). A custom
// StringSegmenter may be used to customize how strings are segmented into
// nodes. A classic trie might segment keys by rune (i.e. unicode points).
type PathTrie[T any] struct {
	segmenter StringSegmenter // key segmenter, must not cause heap allocs
	Value     *T
	Children  map[string]*PathTrie[T]
}

// PathTrieConfig for building a path trie with different segmenter
type PathTrieConfig struct {
	Segmenter StringSegmenter
}

// NewPathTrie allocates and returns a new *PathTrie.
func NewPathTrie[T any]() *PathTrie[T] {
	return &PathTrie[T]{
		segmenter: PathSegmenter,
	}
}

// NewPathTrieWithConfig allocates and returns a new *PathTrie with the given *PathTrieConfig
func NewPathTrieWithConfig[T any](config *PathTrieConfig) *PathTrie[T] {
	segmenter := PathSegmenter
	if config != nil && config.Segmenter != nil {
		segmenter = config.Segmenter
	}

	return &PathTrie[T]{
		segmenter: segmenter,
	}
}

// newPathTrieFromTrie returns new trie while preserving its config
func (trie *PathTrie[T]) newPathTrie() *PathTrie[T] {
	return &PathTrie[T]{
		segmenter: trie.segmenter,
	}
}

// Get returns the Value stored at the given key. Returns nil for internal
// nodes or for nodes with a Value of nil.
func (trie *PathTrie[T]) Get(key string) T {
	node := trie
	for part, i := trie.segmenter(key, 0); part != ""; part, i = trie.segmenter(key, i) {
		node = node.Children[part]
		if node == nil {
			return *new(T)
		}
	}
	return *node.Value
}

// Put inserts the Value into the trie at the given key, replacing any
// existing items. It returns true if the put adds a new Value, false
// if it replaces an existing Value.
// Note that internal nodes have nil values so a stored nil Value will not
// be distinguishable and will not be included in Walks.
func (trie *PathTrie[T]) Put(key string, value T) {
	node := trie
	for part, i := trie.segmenter(key, 0); part != ""; part, i = trie.segmenter(key, i) {
		child := node.Children[part]
		if child == nil {
			if node.Children == nil {
				node.Children = map[string]*PathTrie[T]{}
			}
			child = trie.newPathTrie()
			node.Children[part] = child
		}
		node = child
	}
	node.Value = &value

}

// Walk iterates over each key/Value stored in the trie and calls the given
// walker function with the key and Value. If the walker function returns
// an error, the walk is aborted.
// The traversal is depth first with no guaranteed order.
func (trie *PathTrie[T]) Walk(walker WalkFunc[T]) error {
	return trie.walk("", walker)
}

// WalkPath iterates over each key/Value in the path in trie from the root to
// the node at the given key, calling the given walker function for each
// key/Value. If the walker function returns an error, the walk is aborted.
func (trie *PathTrie[T]) WalkPath(key string, walker WalkFunc[T]) error {
	// Get root Value if one exists.
	if trie.Value != nil {
		if err := walker("", *trie.Value); err != nil {
			return err
		}
	}
	for part, i := trie.segmenter(key, 0); ; part, i = trie.segmenter(key, i) {
		if trie = trie.Children[part]; trie == nil {
			return nil
		}
		if trie.Value != nil {
			var k string
			if i == -1 {
				k = key
			} else {
				k = key[0:i]
			}
			if err := walker(k, *trie.Value); err != nil {
				return err
			}
		}
		if i == -1 {
			break
		}
	}
	return nil
}

func (trie *PathTrie[T]) walk(key string, walker WalkFunc[T]) error {
	if trie.Value != nil {
		if err := walker(key, *trie.Value); err != nil {
			return err
		}
	}
	for part, child := range trie.Children {
		if err := child.walk(key+part, walker); err != nil {
			return err
		}
	}
	return nil
}

// Merge merge empty nodes
// if a node has no values, assign its part to the parent node
func (trie *PathTrie[T]) Merge() {
	for part, child := range trie.Children {
		child.Merge()
		if child.Value == nil {
			for cPart, cChild := range child.Children {
				trie.Children[part+cPart] = cChild
			}
			delete(trie.Children, part)
		}
	}
}

func (trie *PathTrie[T]) RecursiveDirectChildren() map[string]*PathTrie[T] {
	children := map[string]*PathTrie[T]{}
	for part, child := range trie.Children {
		if child.Value != nil {
			children[part] = child
			continue
		}
		for cPart, cChild := range child.RecursiveDirectChildren() {
			children[part+cPart] = cChild
		}
	}
	return children
}
