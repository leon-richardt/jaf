package main

type Set struct {
	_map map[interface{}]struct{}
}

func NewSet() *Set {
	set := &Set{}
	set._map = make(map[interface{}]struct{})
	return set
}

func (set *Set) Contains(value interface{}) bool {
	_, ok := set._map[value]
	return ok
}

func (set *Set) Insert(value interface{}) bool {
	if set.Contains(value) {
		return false
	}

	set._map[value] = struct{}{}
	return true
}

func (set *Set) Remove(value interface{}) bool {
	if !set.Contains(value) {
		return false
	}

	delete(set._map, value)
	return true
}
