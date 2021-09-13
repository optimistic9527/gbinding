package gbinding

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

type argsInfo struct {
	queryName   string
	fileName    string
	pathNames   []string
	headerNames []string
	cookieNames []string

	filedNameIsEqual func(fieldName, inputName string) bool
	args             []*argTypeInfo

	//要携带上下文那种
	//customerFieldBind map[string]func(c *gin.Context, fieldValue reflect.Value) error
}

//checkBindValue  检查绑定的值是否合法
func (a *argsInfo) checkBindValue(bindingTypeInfo *argTypeInfo) {
	switch bindingTypeInfo.argTypeEnum {
	case customizeStructArg, customizeStructPrtArg:
		structBasicType := bindingTypeInfo.GetBasicType()
		validValue := reflect.New(structBasicType).Elem()
		allNeedCheckField := make([]string, 0)
		allNeedCheckField = append(allNeedCheckField, a.pathNames...)
		allNeedCheckField = append(allNeedCheckField, a.headerNames...)
		allNeedCheckField = append(allNeedCheckField, a.cookieNames...)
		a.checkFieldValid(structBasicType, validValue, allNeedCheckField)
	case basicSliceArg:
		if len(a.queryName) == 0 {
			log.Panicf("BasicSlice arg must set queryName")
		}
	case basicArg:
		if len(a.queryName) == 0 && len(a.pathNames) == 0 && len(a.headerNames) == 0 && len(a.cookieNames) == 0 {
			log.Panicf("Basic arg must set one of name  (queryName,pathNames,headerNames,cookieNames)")
		}
	}
}

//checkFieldValid 当绑定要struct上时，需要检查用户设置的name所匹配的字段是否存在，是否能设置
func (a *argsInfo) checkFieldValid(structType reflect.Type, structValue reflect.Value, fieldNames []string) {
	//所有的参数都绑定到struct上,
	if len(fieldNames) != 0 {
		//检查这些字段是否存在
		for i := range fieldNames {
			value := fieldNames[i]
			_, ok := structType.FieldByNameFunc(func(s string) bool {
				return a.filedNameIsEqual(s, value)
			})
			if !ok {
				log.Panicf("struct:%s field:%s no found,please check", structType.String(), value)
			}
			if !structValue.FieldByName(value).CanSet() {
				log.Panicf("struct:%s field:%s can't set,please check is export", structType.String(), value)
			}
		}
	}
}

func (a *argsInfo) binding(gctx *gin.Context, argInfo *argTypeInfo, argValues []reflect.Value) error {
	switch argInfo.argTypeEnum {
	case fileHeader:
		file, err := gctx.FormFile(a.fileName)
		if err != nil {
			return err
		}
		argValues = append(argValues, reflect.ValueOf(file))
	case multiFile:
		form, err := gctx.MultipartForm()
		if err != nil {
			return err
		}
		argValues = append(argValues, reflect.ValueOf(form))
	case customizeStructArg, customizeStructPrtArg:
		var elemValuePrt reflect.Value
		if argInfo.argTypeEnum == customizeStructArg {
			elemValuePrt = reflect.New(argInfo.argType)
		} else {
			elemValuePrt = reflect.New(argInfo.argType.Elem())
		}

		if err := gctx.ShouldBind(elemValuePrt.Interface()); err != nil {
			return err
		}

		elemValue := elemValuePrt.Elem()
		//如果需要设置绑定uri数据，也帮忙绑定了
		pathNames := a.pathNames
		for i := range pathNames {
			filedValue := elemValue.FieldByNameFunc(func(s string) bool {
				return a.filedNameIsEqual(s, pathNames[i])
			})
			if err := setBasicValue(filedValue, gctx.Param(pathNames[i])); err != nil {
				return err
			}
		}

		//如果需要设置绑定header数据，也帮忙绑定了
		headerNames := a.headerNames
		for i := range headerNames {
			filedValue := elemValue.FieldByNameFunc(func(s string) bool {
				return a.filedNameIsEqual(s, headerNames[i])
			})

			if err := setBasicValue(filedValue, gctx.GetHeader(headerNames[i])); err != nil {
				return nil
			}
		}

		cookieNames := a.cookieNames
		for i := range cookieNames {
			filedValue := elemValue.FieldByNameFunc(func(s string) bool {
				return a.filedNameIsEqual(s, cookieNames[i])
			})
			cookie, err := gctx.Cookie(headerNames[i])
			if err != nil {
				return err
			}
			if err := setBasicValue(filedValue, cookie); err != nil {
				return nil
			}
		}

		//用户是需要接收结构体
		if argInfo.argTypeEnum == customizeStructArg {
			argValues = append(argValues, elemValuePrt.Elem())
		} else {
			argValues = append(argValues, elemValuePrt)
		}

	case basicArg:
		var (
			data  string
			exist bool
		)
		//先从url后面取
		queryName := a.queryName
		if queryName != "" {
			data, exist = gctx.GetQuery(queryName)
			//没有就从post
			if !exist {
				data, exist = gctx.GetPostForm(queryName)
			}
		}

		//再没有就从uri上获取
		if !exist && len(a.pathNames) > 0 {
			data = gctx.Param(a.pathNames[0])
			exist = data != ""
		}

		//支持从header上面获取
		if !exist && len(a.headerNames) > 0 {
			data = gctx.GetHeader(a.headerNames[0])
			exist = data != ""
		}

		//最后还是没有的话,支持从cookie上面获取
		if !exist && len(a.cookieNames) > 0 {
			data, _ = gctx.Cookie(a.cookieNames[0])
			exist = data != ""
		}
		if !exist {
			return fmt.Errorf("try to binding %s,but can't get from url,uri,header,cookie. please check you option name set", argInfo.argType.String())
		}
		value := reflect.New(argInfo.argType).Elem()
		if err := setBasicValue(value, data); err != nil {
			return err
		}
		argValues = append(argValues, value)
		return nil
	case basicSliceArg:
		var (
			data  []string
			exist bool
		)
		queryName := a.queryName
		data, exist = gctx.GetQueryArray(queryName)
		if !exist {
			data, exist = gctx.GetPostFormArray(queryName)
		}
		if !exist {
			return fmt.Errorf("try to binding %s,but can't get from url,uri,header,cookie. please check you option name set", argInfo.argType.String())
		}
		slice := reflect.MakeSlice(argInfo.argType, len(data), len(data))
		if err := setBasicSlice(slice, argInfo.argType.Elem().Kind(), data); err != nil {
			return err
		}
		argValues = append(argValues, slice)
	}
	return nil
}

func defaultArgInfo() argsInfo {
	return argsInfo{
		filedNameIsEqual: defaultFieldMatcher,
		args:             []*argTypeInfo{},
	}
}

var defaultFieldMatcher = func(fieldName, inputName string) bool {
	return strings.EqualFold(strings.ReplaceAll(fieldName, "_", ""),
		strings.ReplaceAll(inputName, "_", ""))
}

//SetGlobalResponse Customizing the global return
func SetGlobalFieldMatcher(fieldMatcher func(fieldName, inputName string) bool) {
	if fieldMatcher == nil {
		log.Panic("set global fieldMatcher can't null")
	}
	defaultFieldMatcher = fieldMatcher
}
