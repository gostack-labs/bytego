package bytego

import (
	"testing"
)

func Test_binder_bindDefault(t *testing.T) {
	type struct2 struct {
		S  string `default:"string1"`
		S2 string `default:"string2"`
		B  bool   `default:"true"`
		I  int    `default:"999"`
	}
	type struct1 struct {
		Int         int  `default:"1"`
		IntPtr      *int `default:"2"`
		IntPtr2     *int
		Bool        bool `default:"true"`
		Struct2     struct2
		Struct3     *struct2 `default:"new"`
		Struct4     *struct2
		IntSlice    []int    `default:"100,200,300"`
		StringSlice []string `default:"abc,def,ghi,jk"`
	}

	t.Run("bind default test", func(t *testing.T) {
		b := &binder{}
		v := &struct1{}
		if err := b.bindDefault(v, "default"); err != nil {
			t.Errorf("binder.bindDefault() error = %v", err)
		}
		if v.Int != 1 {
			t.Error("bind int error")
		}
		if v.IntPtr == nil || *v.IntPtr != 2 {
			t.Error("bind intptr error")
		}
		if v.IntPtr2 != nil {
			t.Error("bind intptr2 error")
		}
		if v.Bool == false {
			t.Error("bind bool error")
		}
		if v.Struct2.S != "string1" || v.Struct2.S2 != "string2" || v.Struct2.B != true || v.Struct2.I != 999 {
			t.Error("bind embeddedd struct error")
		}
		if v.Struct3 == nil || v.Struct3.S != "string1" {
			t.Error("bind embedded struct new ptr error")
		}
		if v.Struct4 != nil {
			t.Error("bind embedded struct ptr error")
		}
		if len(v.IntSlice) != 3 || v.IntSlice[0] != 100 || v.IntSlice[1] != 200 || v.IntSlice[2] != 300 {
			t.Error("bind int slice error")
		}
		if len(v.StringSlice) != 4 || v.StringSlice[0] != "abc" || v.StringSlice[1] != "def" || v.StringSlice[2] != "ghi" || v.StringSlice[3] != "jk" {
			t.Error("bind string slice error")
		}
	})
}
