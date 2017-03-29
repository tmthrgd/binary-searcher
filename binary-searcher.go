// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

package searcher

import (
	"bytes"
	"sort"
)

type BinarySearcher struct {
	Data []byte
	Size int

	IncrementBytes func(base, data []byte)

	buffer []byte
}

func New(data []byte, size int) *BinarySearcher {
	return &BinarySearcher{
		Data: data,
		Size: size,

		IncrementBytes: incrementBytes,
	}
}

func (s *BinarySearcher) Len() int {
	return len(s.Data) / s.Size
}

func (s *BinarySearcher) Less(i, j int) bool {
	return bytes.Compare(s.Data[i*s.Size:(i+1)*s.Size], s.Data[j*s.Size:(j+1)*s.Size]) < 0
}

func (s *BinarySearcher) Swap(i, j int) {
	if s.buffer == nil {
		s.buffer = make([]byte, s.Size)
	}

	copy(s.buffer, s.Data[i*s.Size:(i+1)*s.Size])
	copy(s.Data[i*s.Size:(i+1)*s.Size], s.Data[j*s.Size:(j+1)*s.Size])
	copy(s.Data[j*s.Size:(j+1)*s.Size], s.buffer)
}

func (s *BinarySearcher) Sort() {
	sort.Sort(s)
}

func (s *BinarySearcher) Index(check []byte) int {
	if len(check) != s.Size {
		panic("invalid size")
	}

	return sort.Search(s.Len(), func(i int) bool {
		return bytes.Compare(s.Data[i*s.Size:(i+1)*s.Size], check) >= 0
	})
}

func (s *BinarySearcher) search(check []byte) (pos int, has bool) {
	pos = s.Index(check)
	has = pos*s.Size < len(s.Data) && bytes.Equal(check, s.Data[pos*s.Size:(pos+1)*s.Size])
	return
}

func (s *BinarySearcher) Contains(check []byte) bool {
	if s == nil {
		return false
	}

	_, has := s.search(check)
	return has
}

func (s *BinarySearcher) Insert(b []byte) bool {
	pos, has := s.search(b)
	if has {
		return false
	}

	s.Data = append(s.Data, b...)
	copy(s.Data[(pos+1)*s.Size:], s.Data[pos*s.Size:])
	copy(s.Data[pos*s.Size:(pos+1)*s.Size], b)
	return true
}

func (s *BinarySearcher) Remove(b []byte) bool {
	pos, has := s.search(b)
	if has {
		s.Data = append(s.Data[:pos*s.Size], s.Data[(pos+1)*s.Size:]...)
		return true
	}

	return false
}

func (s *BinarySearcher) InsertRange(base []byte, num int) {
	if num < 0 {
		panic("num is negative")
	}

	startPos := s.Index(base)
	var endPos int

	if startPos*s.Size == len(s.Data) {
		endPos = s.Len()
	} else {
		if s.buffer == nil {
			s.buffer = make([]byte, s.Size)
		}

		end := s.buffer
		copy(end, base)

		if addIntToBytes(end, num) {
			panic("overflow")
		}

		endPos = s.Index(end)
		if endPos-startPos == num {
			return
		}
	}

	if need := (s.Len() - (endPos - startPos) + num) * s.Size; cap(s.Data) < need {
		data := make([]byte, need, need+(need>>3) /*= need * 1.125*/)
		copy(data, s.Data[:startPos*s.Size])
		copy(data[(startPos+num)*s.Size:], s.Data[endPos*s.Size:])
		s.Data = data
	} else {
		s.Data = s.Data[:need]
		copy(s.Data[(startPos+num)*s.Size:], s.Data[endPos*s.Size:])
	}

	s.IncrementBytes(base, s.Data[startPos*s.Size:(startPos+num)*s.Size])
}

func (s *BinarySearcher) RemoveRange(base []byte, num int) {
	if num < 0 {
		panic("num is negative")
	}

	startPos := s.Index(base)
	if startPos*s.Size == len(s.Data) {
		return
	}

	if s.buffer == nil {
		s.buffer = make([]byte, s.Size)
	}

	end := s.buffer
	copy(end, base)

	var endPos int

	if addIntToBytes(end, num) {
		endPos = s.Len()
	} else {
		endPos = s.Index(end)
	}

	if startPos != endPos {
		s.Data = append(s.Data[:startPos*s.Size], s.Data[endPos*s.Size:]...)
	}
}

func (s *BinarySearcher) Clear() {
	if cap(s.Data) <= 4096 {
		s.Data = s.Data[:0]
	} else {
		s.Data = nil
	}
}
