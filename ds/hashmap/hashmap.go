package hashmap

type HashMap struct {
	data map[string]map[string][]byte
}

// New creates a new hashmap.
func New() *HashMap {
	return &HashMap{
		data: make(map[string]map[string][]byte),
	}
}

// Put Puts value at a key location under a specified map. It initializes an empty map if name does not exist.
// It returns zero if there is already a value at specified key, returns 1 otherwise.
func (h *HashMap) Put(name, key string, value []byte) (result int) {
	if !h.exists(name) {
		h.data[name] = make(map[string][]byte)
	}

	if _, ok := h.data[name][key]; !ok {
		result = 1
	}

	h.data[name][key] = value

	return result
}

// Get returns the value associated with key within specific map.
func (h *HashMap) Get(name, key string) []byte {
	if !h.exists(name) {
		return nil
	}

	return h.data[name][key]
}

// PutIfAbsent Puts value at a key location under a specified map if it does not exist.
// It returns zero if there is already a value at specified key, returns 1 otherwise.
func (h *HashMap) PutIfAbsent(name, key string, value []byte) (result int) {
	if !h.exists(name) {
		h.data[name] = make(map[string][]byte)
	}

	if _, ok := h.data[name][key]; !ok {
		h.data[name][key] = value
		result = 1
	}

	return result
}

// Remove removes value specified by key from a map. It ignores if key is not in the map.
// It returns 1 if removal is successful, returns 0 otherwise.
func (h *HashMap) Remove(name, key string) int {
	if !h.exists(name) {
		return 0
	}

	if _, ok := h.data[name][key]; ok {
		delete(h.data[name], key)
		return 1

	}

	return 0
}

// Clear removes all the element within map.
// It returns 1 if map exists, returns 0 otherwise.
func (h *HashMap) Clear(name string) int {
	if !h.exists(name) {
		return 0
	}
	delete(h.data, name)

	return 1
}

func (h *HashMap) exists(key string) bool {
	if _, ok := h.data[key]; ok {
		return true
	}
	return false
}
