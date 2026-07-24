/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package serialize_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/internal/serialize"
	"k8s.io/klog/v2/internal/test"
)

// point conforms to fmt.Stringer interface as it implements the String() method
type point struct {
	x int
	y int
}

// we now have a value receiver
func (p point) String() string {
	return fmt.Sprintf("x=%d, y=%d", p.x, p.y)
}

type dummyStruct struct {
	key   string
	value string
}

func (d *dummyStruct) MarshalLog() interface{} {
	return map[string]string{
		"key-data":   d.key,
		"value-data": d.value,
	}
}

type dummyStructWithStringMarshal struct {
	key   string
	value string
}

func (d *dummyStructWithStringMarshal) MarshalLog() interface{} {
	return fmt.Sprintf("%s=%s", d.key, d.value)
}

// Test that kvListFormat works as advertised.
func TestKvListFormat(t *testing.T) {
	var emptyPoint *point
	var testKVList = []struct {
		keysValues []interface{}
		want       string
	}{
		{
			keysValues: []interface{}{"data", &dummyStruct{key: "test", value: "info"}},
			want:       ` data={"key-data":"test","value-data":"info"}`,
		},
		{
			keysValues: []interface{}{"data", &dummyStructWithStringMarshal{key: "test", value: "info"}},
			want:       ` data="test=info"`,
		},
		{
			keysValues: []interface{}{"pod", "kubedns"},
			want:       " pod=\"kubedns\"",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "update", true},
			want:       " pod=\"kubedns\" update=true",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "spec", struct {
				X int
				Y string
				N time.Time
			}{X: 76, Y: "strval", N: time.Date(2006, 1, 2, 15, 4, 5, .067890e9, time.UTC)}},
			want: ` pod="kubedns" spec={"X":76,"Y":"strval","N":"2006-01-02T15:04:05.06789Z"}`,
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "values", []int{8, 6, 7, 5, 3, 0, 9}},
			want:       " pod=\"kubedns\" values=[8,6,7,5,3,0,9]",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "values", []string{"deployment", "svc", "configmap"}},
			want:       ` pod="kubedns" values=["deployment","svc","configmap"]`,
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "bytes", []byte("test case for byte array")},
			want:       " pod=\"kubedns\" bytes=\"test case for byte array\"",
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "bytes", []byte("��=� ⌘")},
			want:       " pod=\"kubedns\" bytes=\"\\ufffd\\ufffd=\\ufffd \\u2318\"",
		},
		{
			keysValues: []interface{}{"multiLineString", `Hello world!
	Starts with tab.
  Starts with spaces.
No whitespace.`,
				"pod", "kubedns",
			},
			want: ` multiLineString=<
	Hello world!
		Starts with tab.
	  Starts with spaces.
	No whitespace.
 > pod="kubedns"`,
		},
		{
			keysValues: []interface{}{"pod", "kubedns", "maps", map[string]int{"three": 4}},
			want:       ` pod="kubedns" maps={"three":4}`,
		},
		{
			keysValues: []interface{}{"pod", klog.KRef("kube-system", "kubedns"), "status", "ready"},
			want:       " pod=\"kube-system/kubedns\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", klog.KRef("", "kubedns"), "status", "ready"},
			want:       " pod=\"kubedns\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", klog.KObj(test.KMetadataMock{Name: "test-name", NS: "test-ns"}), "status", "ready"},
			want:       " pod=\"test-ns/test-name\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", klog.KObj(test.KMetadataMock{Name: "test-name", NS: ""}), "status", "ready"},
			want:       " pod=\"test-name\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", klog.KObj(nil), "status", "ready"},
			want:       " pod=\"\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", klog.KObj((*test.PtrKMetadataMock)(nil)), "status", "ready"},
			want:       " pod=\"\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pod", klog.KObj((*test.KMetadataMock)(nil)), "status", "ready"},
			want:       " pod=\"\" status=\"ready\"",
		},
		{
			keysValues: []interface{}{"pods", klog.KObjs([]test.KMetadataMock{
				{
					Name: "kube-dns",
					NS:   "kube-system",
				},
				{
					Name: "mi-conf",
				},
			})},
			want: ` pods=[{"name":"kube-dns","namespace":"kube-system"},{"name":"mi-conf"}]`,
		},
		{
			keysValues: []interface{}{"point-1", point{100, 200}, "point-2", emptyPoint},
			want:       " point-1=\"x=100, y=200\" point-2=\"<panic: value method k8s.io/klog/v2/internal/serialize_test.point.String called using nil *point pointer>\"",
		},
		{
			keysValues: []interface{}{struct{ key string }{key: "k1"}, "value"},
			want:       " {k1}=\"value\"",
		},
		{
			keysValues: []interface{}{1, "test"},
			want:       " %!s(int=1)=\"test\"",
		},
		{
			keysValues: []interface{}{map[string]string{"k": "key"}, "value"},
			want:       " map[k:key]=\"value\"",
		},
	}

	for _, d := range testKVList {
		b := &bytes.Buffer{}
		serialize.FormatKVs(b, d.keysValues)
		if b.String() != d.want {
			t.Errorf("KVListFormat error:\n got:\n\t%s\nwant:\t%s", b.String(), d.want)
		}
	}
}
