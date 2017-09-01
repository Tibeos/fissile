package helm

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// annotate by recursively adding comments or conditions to each node
func annotate(node Node, comment bool, index int) int {
	index++
	if comment {
		node.Apply(Comment(fmt.Sprintf("comment %d", index)))
	} else {
		node.Apply(Condition(fmt.Sprintf("if condition %d", index)))
	}
	switch node.(type) {
	case *List:
		for _, node := range node.(*List).nodes {
			index = annotate(node, comment, index)
		}
	case *Object:
		for _, namedNode := range node.(*Object).nodes {
			index = annotate(namedNode.node, comment, index)
		}
	}
	return index
}
func addComments(node Node)   { annotate(node, true, 0) }
func addConditions(node Node) { annotate(node, false, 0) }

func equal(t *testing.T, config *Object, expect string) {
	buffer := &bytes.Buffer{}
	NewEncoder(buffer).Encode(config)
	assert.Equal(t, expect, buffer.String())
}

func TestHelmScalar(t *testing.T) {
	root := NewObject()
	root.Add("Scalar", NewScalar("42"))

	equal(t, root, `---
Scalar: 42
`)

	addComments(root)
	equal(t, root, `---
# comment 1
# comment 2
Scalar: 42
`)

	addConditions(root)
	equal(t, root, `---
# comment 1
{{- if condition 1 }}
# comment 2
{{- if condition 2 }}
Scalar: 42
{{- end }}
{{- end }}
`)
}

func TestHelmList(t *testing.T) {
	list := NewList()
	list.Add(NewScalar("1"))
	list.Add(NewScalar("2"))
	list.Add(NewScalar("3"))

	root := NewObject()
	root.Add("List", list)

	equal(t, root, `---
List:
- 1
- 2
- 3
`)

	addComments(root)
	equal(t, root, `---
# comment 1
# comment 2
List:
# comment 3
- 1
# comment 4
- 2
# comment 5
- 3
`)

	addConditions(root)
	equal(t, root, `---
# comment 1
{{- if condition 1 }}
# comment 2
{{- if condition 2 }}
List:
# comment 3
{{- if condition 3 }}
- 1
{{- end }}
# comment 4
{{- if condition 4 }}
- 2
{{- end }}
# comment 5
{{- if condition 5 }}
- 3
{{- end }}
{{- end }}
{{- end }}
`)
}

func TestHelmObject(t *testing.T) {
	obj := NewObject()
	obj.Add("foo", NewScalar("1"))
	obj.Add("bar", NewScalar("2"))
	obj.Add("baz", NewScalar("3"))

	root := NewObject()
	root.Add("Object", obj)

	equal(t, root, `---
Object:
  foo: 1
  bar: 2
  baz: 3
`)

	addComments(root)
	equal(t, root, `---
# comment 1
# comment 2
Object:
  # comment 3
  foo: 1
  # comment 4
  bar: 2
  # comment 5
  baz: 3
`)

	addConditions(root)
	equal(t, root, `---
# comment 1
{{- if condition 1 }}
# comment 2
{{- if condition 2 }}
Object:
  # comment 3
  {{- if condition 3 }}
  foo: 1
  {{- end }}
  # comment 4
  {{- if condition 4 }}
  bar: 2
  {{- end }}
  # comment 5
  {{- if condition 5 }}
  baz: 3
  {{- end }}
{{- end }}
{{- end }}
`)
}

func TestHelmListOfList(t *testing.T) {
	list1 := NewList()
	list1.Add(NewScalar("1"))
	list1.Add(NewScalar("2"))

	list2 := NewList()
	list2.Add(list1)
	list2.Add(NewScalar("x"))
	list2.Add(NewScalar("y"))

	list3 := NewList()
	list3.Add(list2)
	list3.Add(NewScalar("foo"))
	list3.Add(NewScalar("bar"))

	root := NewObject()
	root.Add("List", list3)

	equal(t, root, `---
List:
- - - 1
    - 2
  - x
  - y
- foo
- bar
`)

	addComments(root)
	equal(t, root, `---
# comment 1
# comment 2
List:
# comment 3
- # comment 4
  - # comment 5
    - 1
    # comment 6
    - 2
  # comment 7
  - x
  # comment 8
  - y
# comment 9
- foo
# comment 10
- bar
`)

	addConditions(root)
	equal(t, root, `---
# comment 1
{{- if condition 1 }}
# comment 2
{{- if condition 2 }}
List:
# comment 3
{{- if condition 3 }}
- # comment 4
  {{- if condition 4 }}
  - # comment 5
    {{- if condition 5 }}
    - 1
    {{- end }}
    # comment 6
    {{- if condition 6 }}
    - 2
    {{- end }}
  {{- end }}
  # comment 7
  {{- if condition 7 }}
  - x
  {{- end }}
  # comment 8
  {{- if condition 8 }}
  - y
  {{- end }}
{{- end }}
# comment 9
{{- if condition 9 }}
- foo
{{- end }}
# comment 10
{{- if condition 10 }}
- bar
{{- end }}
{{- end }}
{{- end }}
`)
}

func TestHelmObjectOfObject(t *testing.T) {
	obj1 := NewObject()
	obj1.Add("One", NewScalar("1"))
	obj1.Add("Two", NewScalar("2"))

	obj2 := NewObject()
	obj2.Add("OneTwo", obj1)
	obj2.Add("X", NewScalar("x"))
	obj2.Add("Y", NewScalar("y"))

	obj3 := NewObject()
	obj3.Add("XY", obj2)
	obj3.Add("Foo", NewScalar("foo"))
	obj3.Add("Bar", NewScalar("bar"))

	root := NewObject()
	root.Add("Object", obj3)

	equal(t, root, `---
Object:
  XY:
    OneTwo:
      One: 1
      Two: 2
    X: x
    Y: y
  Foo: foo
  Bar: bar
`)

	addComments(root)
	equal(t, root, `---
# comment 1
# comment 2
Object:
  # comment 3
  XY:
    # comment 4
    OneTwo:
      # comment 5
      One: 1
      # comment 6
      Two: 2
    # comment 7
    X: x
    # comment 8
    Y: y
  # comment 9
  Foo: foo
  # comment 10
  Bar: bar
`)

	addConditions(root)
	equal(t, root, `---
# comment 1
{{- if condition 1 }}
# comment 2
{{- if condition 2 }}
Object:
  # comment 3
  {{- if condition 3 }}
  XY:
    # comment 4
    {{- if condition 4 }}
    OneTwo:
      # comment 5
      {{- if condition 5 }}
      One: 1
      {{- end }}
      # comment 6
      {{- if condition 6 }}
      Two: 2
      {{- end }}
    {{- end }}
    # comment 7
    {{- if condition 7 }}
    X: x
    {{- end }}
    # comment 8
    {{- if condition 8 }}
    Y: y
    {{- end }}
  {{- end }}
  # comment 9
  {{- if condition 9 }}
  Foo: foo
  {{- end }}
  # comment 10
  {{- if condition 10 }}
  Bar: bar
  {{- end }}
{{- end }}
{{- end }}
`)
}

func TestHelmObjectOfList(t *testing.T) {
	list := NewList()
	list.Add(NewScalar("1"))
	list.Add(NewScalar("2"))
	list.Add(NewScalar("3"))

	obj := NewObject()
	obj.Add("List", list)

	root := NewObject()
	root.Add("Object", obj)

	equal(t, root, `---
Object:
  List:
  - 1
  - 2
  - 3
`)

	addComments(root)
	equal(t, root, `---
# comment 1
# comment 2
Object:
  # comment 3
  List:
  # comment 4
  - 1
  # comment 5
  - 2
  # comment 6
  - 3
`)

	addConditions(root)
	equal(t, root, `---
# comment 1
{{- if condition 1 }}
# comment 2
{{- if condition 2 }}
Object:
  # comment 3
  {{- if condition 3 }}
  List:
  # comment 4
  {{- if condition 4 }}
  - 1
  {{- end }}
  # comment 5
  {{- if condition 5 }}
  - 2
  {{- end }}
  # comment 6
  {{- if condition 6 }}
  - 3
  {{- end }}
  {{- end }}
{{- end }}
{{- end }}
`)
}

func TestHelmListOfObject(t *testing.T) {
	obj := NewObject()
	obj.Add("Foo", NewScalar("foo"))
	obj.Add("Bar", NewScalar("bar"))
	obj.Add("Baz", NewScalar("baz"))

	list := NewList()
	list.Add(obj)

	root := NewObject()
	root.Add("Object", list)

	equal(t, root, `---
Object:
- Foo: foo
  Bar: bar
  Baz: baz
`)

	addComments(root)
	equal(t, root, `---
# comment 1
# comment 2
Object:
# comment 3
- # comment 4
  Foo: foo
  # comment 5
  Bar: bar
  # comment 6
  Baz: baz
`)

	addConditions(root)
	equal(t, root, `---
# comment 1
{{- if condition 1 }}
# comment 2
{{- if condition 2 }}
Object:
# comment 3
{{- if condition 3 }}
- # comment 4
  {{- if condition 4 }}
  Foo: foo
  {{- end }}
  # comment 5
  {{- if condition 5 }}
  Bar: bar
  {{- end }}
  # comment 6
  {{- if condition 6 }}
  Baz: baz
  {{- end }}
{{- end }}
{{- end }}
{{- end }}
`)
}

func TestHelmMultiLineComment(t *testing.T) {
	root := NewObject()
	root.Add("Scalar", NewScalar("42", Comment("Many\n\nlines")))

	equal(t, root, `---
# Many
#
# lines
Scalar: 42
`)

	// list of list
	list1 := NewList()
	list1.Add(NewScalar("42", Comment("Many\n\nlines")))

	list2 := NewList()
	list2.Add(list1)
	list2.Add(NewScalar("foo"))

	root = NewObject()
	root.Add("List", list2)

	equal(t, root, `---
List:
- # Many
  #
  # lines
  - 42
- foo
`)
}

func TestHelmWrapLongComments(t *testing.T) {
	root := NewObject()
	obj := NewObject()
	word := "1"
	for i := len(word) + 1; i < 7; i++ {
		word += strconv.Itoa(i)
		root.Add(fmt.Sprintf("Key%d", i), NewScalar("~", Comment(strings.Repeat(word+" ", 10))))
		if i < 5 {
			obj.Add(fmt.Sprintf("Key%d", i), NewScalar("~", Comment(strings.Repeat(word+" ", 5))))
		}
	}

	obj.Add("Very", NewScalar("Long", Comment(strings.Repeat(strings.Repeat("x", 50)+" ", 2))))
	root.Add("Very", NewScalar("Long", Comment(strings.Repeat(strings.Repeat("x", 50)+" ", 2))))
	root.Add("Nested", obj)

	expect := `---
# 12 12 12 12 12 12 12
# 12 12 12
Key2: ~
# 123 123 123 123 123
# 123 123 123 123 123
Key3: ~
# 1234 1234 1234 1234
# 1234 1234 1234 1234
# 1234 1234
Key4: ~
# 12345 12345 12345
# 12345 12345 12345
# 12345 12345 12345
# 12345
Key5: ~
# 123456 123456 123456
# 123456 123456 123456
# 123456 123456 123456
# 123456
Key6: ~
# xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
# xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Very: Long
Nested:
          # 12 12 12 12
          # 12
          Key2: ~
          # 123 123 123
          # 123 123
          Key3: ~
          # 1234 1234
          # 1234 1234
          # 1234
          Key4: ~
          # xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          # xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          Very: Long
`
	buffer := &bytes.Buffer{}
	NewEncoder(buffer, Indent(10), Wrap(24)).Encode(root)
	assert.Equal(t, expect, buffer.String())
}

func TestHelmIndent(t *testing.T) {
	obj1 := NewObject()
	obj1.Add("Foo", NewScalar("Bar", Comment("Baz")))

	list1 := NewList()
	list1.Add(obj1)
	list1.Add(NewScalar("1"))

	list2 := NewList()
	list2.Add(NewScalar("abc"))
	list2.Add(NewScalar("xyz"))

	list1.Add(list2)
	list1.Add(NewScalar("2"))
	list1.Add(NewScalar("3"))

	obj2 := NewObject()
	obj2.Add("List", list1)

	obj3 := NewObject()
	obj3.Add("Foo", NewScalar("1"))
	obj3.Add("Bar", NewScalar("2"))

	obj2.Add("Meta", obj3)

	root := NewObject()
	root.Add("Object", obj2)

	expect := `---
Object:
    List:
      - # Baz
        Foo: Bar
      - 1
      -   - abc
          - xyz
      - 2
      - 3
    Meta:
        Foo: 1
        Bar: 2
`
	buffer := &bytes.Buffer{}
	NewEncoder(buffer, Indent(4)).Encode(root)
	assert.Equal(t, expect, buffer.String())
}

func TestHelmEncoderModifier(t *testing.T) {
	obj := NewObject()
	obj.Add("foo", NewScalar("1"))
	obj.Add("bar", NewScalar("2"))
	obj.Add("baz", NewScalar("3"))

	root := NewObject()
	root.Add("Object", obj)

	expect := `---
Object:
  foo: 1
  bar: 2
  baz: 3
---
Object:
    foo: 1
    bar: 2
    baz: 3
`

	buffer := &bytes.Buffer{}
	enc := NewEncoder(buffer)
	enc.Encode(root)
	enc.Apply(Indent(4))
	enc.Encode(root)
	assert.Equal(t, expect, buffer.String())
}
