package gbinding

import (
	"fmt"
	"log"
	"reflect"

	"github.com/spf13/cast"
)

type argTypeEnum string

const (
	hrArg                 argTypeEnum = "*http.request"
	rwArg                 argTypeEnum = "http.ResponseWriter"
	ctxArg                argTypeEnum = "context.Context"
	customizeStructArg    argTypeEnum = "customizeStruct"
	customizeStructPrtArg argTypeEnum = "customizeStructPrtArg"
	fileHeader            argTypeEnum = "*multipart.fileHeader"
	multiFile             argTypeEnum = "*multipart.Form"
	basicArg              argTypeEnum = "basic"
	basicSliceArg         argTypeEnum = "basicSlice"
)

type argTypeInfo struct {
	argType     reflect.Type
	argTypeEnum argTypeEnum
}

func (a *argTypeInfo) GetBasicType() reflect.Type {
	if a.argTypeEnum == customizeStructPrtArg {
		return a.argType.Elem()
	}
	return a.argType
}

func (a *argTypeInfo) String() string {
	return string(a.argTypeEnum)
}

func (a *argTypeInfo) ValidSecondArgType() bool {
	return a.argTypeEnum != hrArg && a.argTypeEnum != ctxArg
}

func (a *argTypeInfo) ValidFirstArgType() bool {
	return a.argTypeEnum == hrArg || a.argTypeEnum == ctxArg
}

func (a *argTypeInfo) IsHttpRequest() bool {
	return a.argTypeEnum == hrArg
}

func (a *argTypeInfo) IsResponseWriter() bool {
	return a.argTypeEnum == rwArg
}

func (a *argTypeInfo) IsCustomizeStructBind() bool {
	return a.argTypeEnum == customizeStructArg || a.argTypeEnum == customizeStructPrtArg
}

func (a *argTypeInfo) IsCustomizeStruct() bool {
	return a.argTypeEnum == customizeStructArg
}

func (a *argTypeInfo) IsCustomizeStructPrt() bool {
	return a.argTypeEnum == customizeStructPrtArg
}

func (a *argTypeInfo) IsBasic() bool {
	return a.argTypeEnum == basicArg
}

func (a *argTypeInfo) IsBasicSlice() bool {
	return a.argTypeEnum == basicSliceArg
}

func setBasicValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.Bool:
		e, err := cast.ToBoolE(value)
		if err != nil {
			return err
		}
		field.SetBool(e)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e, err := cast.ToInt64E(value)
		if err != nil {
			return err
		}
		field.SetInt(e)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e, err := cast.ToUint64E(value)
		if err != nil {
			return err
		}
		field.SetUint(e)
	case reflect.Float32, reflect.Float64:
		e, err := cast.ToFloat64E(value)
		if err != nil {
			return err
		}
		field.SetFloat(e)
	case reflect.String:
		field.SetString(value)
	default:
		return fmt.Errorf("kind:%v can't set", field.Kind())
	}
	return nil
}

func setBasicSlice(slice reflect.Value, elemKind reflect.Kind, value []string) error {
	switch elemKind {
	case reflect.Bool:
		for i := range value {
			v := value[i]
			e, err := cast.ToBoolE(v)
			if err != nil {
				return err
			}
			index := slice.Index(i)
			index.SetBool(e)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for i := range value {
			v := value[i]
			e, err := cast.ToInt64E(v)
			if err != nil {
				return err
			}
			index := slice.Index(i)
			index.SetInt(e)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for i := range value {
			v := value[i]
			e, err := cast.ToUint64E(v)
			if err != nil {
				return err
			}
			index := slice.Index(i)
			index.SetUint(e)
		}
	case reflect.Float32, reflect.Float64:
		for i := range value {
			v := value[i]
			e, err := cast.ToFloat64E(v)
			if err != nil {
				return err
			}
			index := slice.Index(i)
			index.SetFloat(e)
		}
	case reflect.String:
		for i := range value {
			v := value[i]
			index := slice.Index(i)
			index.SetString(v)
		}
	default:
		return fmt.Errorf("kind:%v can't set", elemKind)
	}
	return nil
}

func toArgTypeEnum(arg reflect.Type) *argTypeInfo {
	result := &argTypeInfo{
		argType: arg,
	}
	switch arg {
	case httpRequestType:
		result.argTypeEnum = hrArg
	case rsWriterType:
		result.argTypeEnum = rwArg
	case contextType:
		result.argTypeEnum = ctxArg
	default:
		switch arg.Kind() {
		case reflect.Ptr:
			if arg.Elem().Kind() != reflect.Struct {
				log.Panicf("expect struct prt but get %s", arg.String())
			}
			switch arg {
			case fileHeaderType:
				result.argTypeEnum = fileHeader
			case multiFileType:
				result.argTypeEnum = multiFile
			default:
				result.argTypeEnum = customizeStructPrtArg
			}
		case reflect.Struct:
			result.argTypeEnum = customizeStructArg
		case reflect.Slice:
			elem := arg.Elem()
			switch elem.Kind() {
			case reflect.Bool, reflect.Float32, reflect.Float64, reflect.String,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			default:
				log.Panicf("only support basic type slice,but this slice elem type is %s", elem.String())
			}
			result.argTypeEnum = basicSliceArg
		case reflect.Bool, reflect.Float32, reflect.Float64, reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result.argTypeEnum = basicArg
		default:
			log.Panicf("unsupport arg type %s", arg.String())
		}
	}
	return result
}
