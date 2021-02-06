package main

import (
	"github.com/segmentio/fasthash/fnv1a"
)

type StringKey string

func (key StringKey) Hash() uint64 {
	return fnv1a.HashString64(string(key))
}

func (key StringKey) Equal(other HamtKey) bool {
	otherKey, isPointerKey := other.(StringKey)
	return isPointerKey && string(otherKey) == string(key)
}
