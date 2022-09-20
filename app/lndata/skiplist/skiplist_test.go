package skiplist

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkiplist(t *testing.T) {
	cases := map[string]func(*testing.T){
		"insert":  testInsert,
		"delete":  testDelete,
		"search":  testSearch,
		"length":  testLength,
		"iterate": testIterate,
		"head":    testHead,
		"tail":    testTail,
	}

	for name, tcase := range cases {
		t.Run(name, tcase)
	}
}

func mustCreateSkiplist(t *testing.T, data map[uint64]interface{}) *Skiplist {
	list := New()
	for key, value := range data {
		_, ok := list.Insert(key, value)
		require.True(t, ok)
	}
	return list
}

func testInsert(t *testing.T) {
	data := map[uint64]interface{}{
		42: "key no.1",
		3:  "key no.2",
		17: "key no.3",
		93: "key no.4",
		10: "key no.5",
		8:  "key no.6",
		13: "key no.7",
	}

	list := mustCreateSkiplist(t, data)

	cases := map[string]struct {
		key           uint64
		value         interface{}
		exists        bool
		returnedValue interface{}
	}{
		"key exists": {
			key:           17,
			value:         "inconsequential",
			exists:        true,
			returnedValue: data[17],
		},
		"smallest existing key": {
			key:           3,
			value:         "inconsequential II - the sequel",
			exists:        true,
			returnedValue: data[3],
		},
		"largest existing key": {
			key:           93,
			value:         "inconsequential III - reborn",
			exists:        true,
			returnedValue: "key no.4",
		},
		"key does not exist": {
			key:   0,
			value: "key no.8",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			node, ok := list.Insert(tc.key, tc.value)

			assert.NotNil(t, node)
			switch tc.exists {
			case true:
				assert.False(t, ok)
				assert.EqualValues(t, tc.returnedValue, node.Value())
			default:
				assert.True(t, ok)
				assert.EqualValues(t, tc.value, node.Value())
			}
		})
	}
}

func testDelete(t *testing.T) {
	data := map[uint64]interface{}{
		31:   "data no.1",
		4:    "data no.2",
		57:   "data no.3",
		9000: "data no.4",
	}

	list := mustCreateSkiplist(t, data)

	cases := []struct {
		name          string
		key           uint64
		exists        bool
		returnedValue interface{}
	}{
		{
			name: "key smaller than min",
			key:  1,
		},
		{
			name: "key larger than max",
			key:  9001,
		},
		{
			name: "key does not exist",
			key:  500,
		},
		{
			name:          "key exists",
			key:           31,
			exists:        true,
			returnedValue: data[31],
		},
		{
			name:          "largest key",
			key:           9000,
			exists:        true,
			returnedValue: data[9000],
		},
		{
			name:          "smallest key",
			key:           4,
			exists:        true,
			returnedValue: data[4],
		},
		{
			name:          "last key",
			key:           57,
			exists:        true,
			returnedValue: data[57],
		},
		{
			name: "empty skiplist",
			key:  42,
		},
	}

	for _, tc := range cases {
		result := list.Delete(tc.key)

		switch tc.exists {
		case true:
			assert.NotNil(t, result)
			assert.EqualValues(t, tc.returnedValue, result.Value())
		default:
			assert.Nil(t, result)
		}
	}
}

func testSearch(t *testing.T) {
	data := map[uint64]interface{}{
		42: "key no.1",
		3:  "key no.2",
		17: "key no.3",
		93: "key no.4",
		10: "key no.5",
		8:  "key no.6",
		13: "key no.7",
	}

	list := mustCreateSkiplist(t, data)

	cases := map[string]struct {
		key           uint64
		exists        bool
		returnedValue interface{}
	}{
		"key smaller than min": {
			key: 1,
		},
		"key larger than max": {
			key: 101,
		},
		"key does not exist": {
			key: 18,
		},
		"smallest key": {
			key: 3, exists: true, returnedValue: data[3],
		},
		"key exists": {
			key: 17, exists: true, returnedValue: data[17],
		},
		"largest key": {
			key: 93, exists: true, returnedValue: data[93],
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := list.Search(tc.key)

			switch tc.exists {
			case true:
				assert.NotNil(t, result)
				assert.EqualValues(t,
					tc.returnedValue, result.Value())
			default:
				assert.Nil(t, result)
			}
		})
	}
}

func testHead(t *testing.T) {
	data := map[uint64]interface{}{
		42: "a",
		13: "b",
		17: "c",
		3:  "d",
	}

	cases := map[string]struct {
		data map[uint64]interface{}
		test func(*testing.T, *Skiplist)
	}{
		"non-empty skiplist": {
			data: data,
			test: func(t *testing.T, sl *Skiplist) {
				var firstKey uint64 = 3

				head := sl.Head()

				assert.NotNil(t, head)
				assert.EqualValues(t, firstKey, head.Key())
				assert.EqualValues(t,
					data[firstKey], head.Value())
			},
		},
		"empty skiplist": {
			data: map[uint64]interface{}{},
			test: func(t *testing.T, sl *Skiplist) {
				head := sl.Head()

				assert.Nil(t, head)
			},
		},
		"after delete": {
			data: data,
			test: func(t *testing.T, sl *Skiplist) {
				var beforeMin, afterMin uint64 = 3, 13

				before := sl.Head()
				require.NotNil(t, before)
				require.EqualValues(t, beforeMin, before.Key())

				deleted := sl.Delete(beforeMin)
				require.NotNil(t, deleted)
				require.EqualValues(t, beforeMin, deleted.Key())

				after := sl.Head()
				assert.NotNil(t, after)
				assert.EqualValues(t, afterMin, after.Key())
			},
		},
		"after insert": {
			data: data,
			test: func(t *testing.T, sl *Skiplist) {
				var beforeMin uint64 = 3
				newKey, newValue := uint64(1), "new value"

				before := sl.Head()
				require.NotNil(t, before)
				require.EqualValues(t, beforeMin, before.Key())

				inserted, ok := sl.Insert(newKey, newValue)
				require.True(t, ok)
				require.NotNil(t, inserted)

				after := sl.Head()
				assert.NotNil(t, after)
				assert.EqualValues(t, newKey, after.Key())
			},
		},
	}

	for name, tc := range cases {
		list := mustCreateSkiplist(t, tc.data)
		t.Run(name, func(t *testing.T) {
			tc.test(t, list)
		})
	}
}

func testTail(t *testing.T) {
	data := map[uint64]interface{}{
		42: "a",
		13: "b",
		17: "c",
		3:  "d",
	}

	cases := map[string]struct {
		data map[uint64]interface{}
		test func(*testing.T, *Skiplist)
	}{
		"non-empty skiplist": {
			data: data,
			test: func(t *testing.T, sl *Skiplist) {
				var lastKey uint64 = 42

				tail := sl.Tail()

				assert.NotNil(t, tail)
				assert.EqualValues(t, lastKey, tail.Key())
				assert.EqualValues(t,
					data[lastKey], tail.Value())
			},
		},
		"empty skiplist": {
			data: map[uint64]interface{}{},
			test: func(t *testing.T, sl *Skiplist) {
				tail := sl.Tail()

				assert.Nil(t, tail)
			},
		},
		"after delete": {
			data: data,
			test: func(t *testing.T, sl *Skiplist) {
				var beforeMax, afterMax uint64 = 42, 17

				before := sl.Tail()
				require.NotNil(t, before)
				require.EqualValues(t, beforeMax, before.Key())

				deleted := sl.Delete(beforeMax)
				require.NotNil(t, deleted)
				require.EqualValues(t, beforeMax, deleted.Key())

				after := sl.Tail()
				assert.NotNil(t, after)
				assert.EqualValues(t, afterMax, after.Key())
			},
		},
		"after insert": {
			data: data,
			test: func(t *testing.T, sl *Skiplist) {
				var beforeMax uint64 = 42
				newKey, newValue := uint64(91), "new value"

				before := sl.Tail()
				require.NotNil(t, before)
				require.EqualValues(t, beforeMax, before.Key())

				inserted, ok := sl.Insert(newKey, newValue)
				require.True(t, ok)
				require.NotNil(t, inserted)

				after := sl.Tail()
				assert.NotNil(t, after)
				assert.EqualValues(t, newKey, after.Key())
			},
		},
	}

	for name, tc := range cases {
		list := mustCreateSkiplist(t, tc.data)
		t.Run(name, func(t *testing.T) {
			tc.test(t, list)
		})
	}
}

type testItem struct {
	key   uint64
	value interface{}
}

func testLength(t *testing.T) {
	cases := map[string]struct {
		items          []testItem
		expectedLength uint64
	}{
		"normal": {
			items: []testItem{
				{key: 7},
				{key: 6},
				{key: 5},
				{key: 4},
				{key: 3},
				{key: 2},
				{key: 1},
			},
			expectedLength: 7,
		},
		"empty skiplist": {
			expectedLength: 0,
		},
		"after duplicate insertion attempts": {
			items: []testItem{
				{key: 7},
				{key: 6},
				{key: 5},
				{key: 4},
				{key: 3},
				{key: 2},
				{key: 1},
				{key: 7},
				{key: 3},
				{key: 0},
			},
			expectedLength: 8,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			list := New()
			for _, item := range tc.items {
				list.Insert(item.key, item.value)
			}

			assert.EqualValues(t, tc.expectedLength, list.Length())
		})
	}
}

func testIterate(t *testing.T) {
	allItems := []testItem{
		{key: 2, value: "data no.1"},
		{key: 3, value: "data no.2"},
		{key: 4, value: "data no.3"},
		{key: 5, value: "data no.4"},
		{key: 7, value: "data no.5"},
		{key: 19, value: "data no.6"},
	}

	data := map[uint64]interface{}{}
	for _, item := range allItems {
		data[item.key] = item.value
	}

	list := mustCreateSkiplist(t, data)

	cases := map[string]struct {
		list     *Skiplist
		startKey uint64
		stopKey  uint64
		expected []testItem
	}{
		"empty skiplist": {
			list: New(),
		},
		"all nodes visited in order": {
			list:     list,
			expected: allItems[:],
		},
		"visitor stops on first node": {
			list:     list,
			startKey: 1,
			stopKey:  2,
			expected: allItems[:1],
		},
		"visitor starts on existing key": {
			list:     list,
			startKey: 4,
			expected: allItems[2:],
		},
		"visitor starts on non-existent key": {
			list:     list,
			startKey: 1,
			expected: allItems[:],
		},
		"visitor stops before end": {
			list:     list,
			stopKey:  5,
			expected: allItems[:4],
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var visited []testItem
			count := tc.list.Iterate(tc.startKey, func(n *Node) bool {
				assert.NotNil(t, n)

				nodeKey := n.Key()
				visited = append(visited, testItem{
					key:   nodeKey,
					value: n.Value(),
				})

				return !(nodeKey == tc.stopKey)
			})

			assert.EqualValues(t, len(tc.expected), count)
			assert.Equal(t, tc.expected, visited)
		})
	}
}

func TestNode(t *testing.T) {
	data := map[uint64]interface{}{
		42: "key no.1",
		3:  "key no.2",
		17: "key no.3",
		93: "key no.4",
		10: "key no.5",
		8:  "key no.6",
		13: "key no.7",
	}

	cases := map[string]func(*testing.T, *Skiplist){
		"key": func(t *testing.T, sl *Skiplist) {
			var targetKey uint64 = 17

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			assert.EqualValues(t, targetKey, target.Key())
		},
		"value": func(t *testing.T, sl *Skiplist) {
			var targetKey uint64 = 17

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			assert.EqualValues(t, data[targetKey], target.Value())
		},
		"set": func(t *testing.T, sl *Skiplist) {
			var targetKey uint64 = 17
			oldValue, newValue := data[targetKey], "new value"

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			assert.EqualValues(t, oldValue, target.Value())

			target.Set(newValue)
			assert.EqualValues(t, newValue, target.Value())

			newTarget := sl.Search(targetKey)
			assert.EqualValues(t, newValue, newTarget.Value())
		},
		"previous": func(t *testing.T, sl *Skiplist) {
			var targetKey, prevKey uint64 = 17, 13

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			prev := sl.Search(prevKey)
			require.NotNil(t, prev)

			nodePrev := target.Previous()
			assert.NotNil(t, nodePrev)
			assert.EqualValues(t, prev, nodePrev)
		},
		"previous after delete": func(t *testing.T, sl *Skiplist) {
			var targetKey, before, after uint64 = 17, 13, 10

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			beforePrev := target.Previous()
			assert.NotNil(t, beforePrev)
			assert.EqualValues(t, before, beforePrev.Key())

			deleted := sl.Delete(before)
			require.NotNil(t, deleted)

			afterPrev := target.Previous()
			assert.NotNil(t, afterPrev)
			assert.EqualValues(t, after, afterPrev.Key())
		},
		"previous after insert": func(t *testing.T, sl *Skiplist) {
			var targetKey, before uint64 = 17, 13
			newKey, newVal := uint64(14), "new value"

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			beforePrev := target.Previous()
			assert.NotNil(t, beforePrev)
			assert.EqualValues(t, before, beforePrev.Key())

			inserted, ok := sl.Insert(newKey, newVal)
			require.True(t, ok)
			require.NotNil(t, inserted)

			afterPrev := target.Previous()
			assert.NotNil(t, afterPrev)
			assert.EqualValues(t, newKey, afterPrev.Key())
		},
		"previous of first is nil": func(t *testing.T, sl *Skiplist) {
			var targetKey uint64 = 3

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			assert.Nil(t, target.Previous())
		},
		"next": func(t *testing.T, sl *Skiplist) {
			var targetKey, nextKey uint64 = 17, 42

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			next := sl.Search(nextKey)
			require.NotNil(t, next)

			nodeNext := target.Next()
			assert.NotNil(t, nodeNext)
			assert.EqualValues(t, next, nodeNext)
		},
		"next after delete": func(t *testing.T, sl *Skiplist) {
			var targetKey, before, after uint64 = 17, 42, 93

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			beforeNext := target.Next()
			assert.NotNil(t, beforeNext)
			assert.EqualValues(t, before, beforeNext.Key())

			deleted := sl.Delete(before)
			require.NotNil(t, deleted)

			afterNext := target.Next()
			assert.NotNil(t, afterNext)
			assert.EqualValues(t, after, afterNext.Key())
		},
		"next after insert": func(t *testing.T, sl *Skiplist) {
			var targetKey, before uint64 = 17, 42
			newKey, newVal := uint64(23), "new value"

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			beforeNext := target.Next()
			assert.NotNil(t, beforeNext)
			assert.EqualValues(t, before, beforeNext.Key())

			inserted, ok := sl.Insert(newKey, newVal)
			require.True(t, ok)
			require.NotNil(t, inserted)

			afterNext := target.Next()
			assert.NotNil(t, afterNext)
			assert.EqualValues(t, newKey, afterNext.Key())
		},
		"next of last is nil": func(t *testing.T, sl *Skiplist) {
			var targetKey uint64 = 93

			target := sl.Search(targetKey)
			require.NotNil(t, target)

			assert.Nil(t, target.Next())
		},
	}

	for name, test := range cases {
		list := mustCreateSkiplist(t, data)
		t.Run(name, func(t *testing.T) {
			test(t, list)
		})
	}
}
