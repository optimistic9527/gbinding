package gbinding

import (
	"context"
	"mime/multipart"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-playground/assert/v2"
)

func Test_toArgTypeEnum(t *testing.T) {
	t.Run("*http.Request", func(t *testing.T) {
		typeInfo := toArgTypeEnum(reflect.TypeOf(&http.Request{}))
		assert.Equal(t, typeInfo.argTypeEnum, hrArg)
	})

	t.Run("http.ResponseWriter", func(t *testing.T) {
		typeInfo := toArgTypeEnum(reflect.TypeOf((*http.ResponseWriter)(nil)).Elem())
		assert.Equal(t, typeInfo.argTypeEnum, rwArg)
	})

	t.Run("context.Context", func(t *testing.T) {
		typeInfo := toArgTypeEnum(reflect.TypeOf((*context.Context)(nil)).Elem())
		assert.Equal(t, typeInfo.argTypeEnum, ctxArg)
	})

	t.Run("CustomizeStruct", func(t *testing.T) {
		typeInfo := toArgTypeEnum(reflect.TypeOf(CustomerStruct{}))
		assert.Equal(t, typeInfo.argTypeEnum, customizeStructArg)
	})

	t.Run("customizeStructPrtArg", func(t *testing.T) {
		t.Run("normal", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf(&CustomerStruct{}))
			assert.Equal(t, typeInfo.argTypeEnum, customizeStructPrtArg)
		})
		t.Run("notStruct", func(t *testing.T) {
			defer func() {
				i := recover()
				assert.Equal(t, i.(string), "expect struct prt but get *int")
			}()
			toArgTypeEnum(reflect.TypeOf((*int)(nil)))
		})

	})
	t.Run("*multipart.fileHeader", func(t *testing.T) {
		typeInfo := toArgTypeEnum(reflect.TypeOf(&multipart.FileHeader{}))
		assert.Equal(t, typeInfo.argTypeEnum, fileHeader)
	})

	t.Run("*multipart.Form", func(t *testing.T) {
		typeInfo := toArgTypeEnum(reflect.TypeOf(&multipart.Form{}))
		assert.Equal(t, typeInfo.argTypeEnum, multiFile)
	})

	t.Run("basic", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf((*int)(nil)).Elem())
			assert.Equal(t, typeInfo.argTypeEnum, basicArg)
		})
		t.Run("bool", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf((*bool)(nil)).Elem())
			assert.Equal(t, typeInfo.argTypeEnum, basicArg)
		})
		t.Run("float", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf((*float64)(nil)).Elem())
			assert.Equal(t, typeInfo.argTypeEnum, basicArg)
		})
		t.Run("string", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf((*string)(nil)).Elem())
			assert.Equal(t, typeInfo.argTypeEnum, basicArg)
		})
	})

	t.Run("basicSlice", func(t *testing.T) {
		t.Run("int", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf([]int{}))
			assert.Equal(t, typeInfo.argTypeEnum, basicSliceArg)
		})
		t.Run("bool", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf([]bool{}))
			assert.Equal(t, typeInfo.argTypeEnum, basicSliceArg)
		})
		t.Run("float", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf([]float64{}))
			assert.Equal(t, typeInfo.argTypeEnum, basicSliceArg)
		})
		t.Run("string", func(t *testing.T) {
			typeInfo := toArgTypeEnum(reflect.TypeOf([]string{}))
			assert.Equal(t, typeInfo.argTypeEnum, basicSliceArg)
		})

		t.Run("notBasicType", func(t *testing.T) {
			defer func() {
				i := recover()
				assert.Equal(t, i.(string), "only support basic type slice,but this slice elem type is gbinding.CustomerStruct")
			}()
			toArgTypeEnum(reflect.TypeOf([]CustomerStruct{}))
		})
	})
}

type CustomerStruct struct {
}

func Test_setBasicSlice(t *testing.T) {
	type args struct {
		slice    reflect.Value
		elemKind reflect.Kind
		value    []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setBasicSlice(tt.args.slice, tt.args.elemKind, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("setBasicSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setBasicValue(t *testing.T) {
	type args struct {
		field reflect.Value
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setBasicValue(tt.args.field, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("setBasicValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
