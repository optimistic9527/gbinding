package gbinding

import (
	"context"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

var rsWriterType = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()

var httpRequestType = reflect.TypeOf(&http.Request{})
var fileHeaderType = reflect.TypeOf(&multipart.FileHeader{})
var multiFileType = reflect.TypeOf(&multipart.Form{})

// errorType error 的反射类型
var errorType = reflect.TypeOf((*error)(nil)).Elem()

// contextType context.Context 的反射类型
var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

type callFunc struct {
	rsInfo response
	asInfo argsInfo

	callFnType  reflect.Type
	callFnValue reflect.Value
}

type CallOption func(c *callFunc)

func BindingAndInvoke(invokeFunc interface{}, ops ...CallOption) gin.HandlerFunc {
	c := &callFunc{
		asInfo: defaultArgInfo(),
	}
	for i := range ops {
		ops[i](c)
	}
	invokeFuncType := reflect.TypeOf(invokeFunc)
	if invokeFuncType == nil {
		log.Panic("arg invokeFunc expect a func bug get nil")
	}
	c.callFnType = invokeFuncType
	c.callFnValue = reflect.ValueOf(invokeFunc)
	if invokeFuncType.Kind() != reflect.Func {
		log.Panicf("arg invokeFunc expect a func bug get %s", invokeFuncType.String())
	}

	checkFuncArg(c, invokeFuncType)
	checkFuncReturn(c, invokeFuncType)
	return c.handlerFunc
}

func (c *callFunc) handlerFunc(gctx *gin.Context) {
	argsNum := len(c.asInfo.args)
	argValues := make([]reflect.Value, 0, argsNum)
	first := c.asInfo.args[0]
	var firstValue reflect.Value
	if first.IsHttpRequest() {
		firstValue = reflect.ValueOf(gctx.Request)
	} else {
		firstValue = reflect.ValueOf(gctx.Request.Context())
	}
	argValues = append(argValues, firstValue)
	var second *argTypeInfo

	if argsNum == 2 {
		second = c.asInfo.args[1]
		if second.IsResponseWriter() { //当两个参数是 要么是http.ResponseWriter 要么是需要绑定参数
			argValues = append(argValues, reflect.ValueOf(gctx.Writer.(http.ResponseWriter)))
		} else if err := c.asInfo.binding(gctx, second, argValues); err != nil {
			c.rsInfo.Return(gctx, nil, err)
			return
		}
	}

	if argsNum == 3 { //三个参数必定是第二个参数是http.ResponseWriter，第三个参数是要绑定的参数
		argValues = append(argValues, reflect.ValueOf(gctx.Writer.(http.ResponseWriter)))
		if err := c.asInfo.binding(gctx, c.asInfo.args[2], argValues); err != nil {
			c.rsInfo.Return(gctx, nil, err)
			return
		}
	}

	result := c.callFnValue.Call(argValues)
	/*//对于函数调用的结果不做任何处理
	if c.rsInfo.skipGlobalResponse {
		gctx.Next()
		return
	}*/

	errValue := result[0]
	//产生错误的情况下，统一返回。
	if !errValue.IsNil() {
		c.rsInfo.Return(gctx, nil, errValue.Interface().(error))
		return
	}
	//用户已经绑定了Writer的情况下，返回数据，就用全局Response，没有返回数据代表了，用户已经自定义返回数据了
	if second != nil && second.IsResponseWriter() && !c.rsInfo.hasData {
		gctx.Next()
		return
	}
	var data interface{}
	if c.rsInfo.hasData {
		data = result[1].Interface()
	}
	c.rsInfo.Return(gctx, data, nil)
}

func checkFuncReturn(c *callFunc, funcType reflect.Type) {
	out := funcType.NumOut()
	if out == 0 || out > 2 {
		log.Panicf("func return arg must err or (anyData,error)")
	}
	if out == 1 {
		in := funcType.In(0)
		if in != errorType {
			log.Panicf("invokeFunc return on arg must error")
		}
	}

	if out == 2 {
		in1 := funcType.In(0)
		in2 := funcType.In(1)
		if in1 == errorType || in2 != errorType {
			log.Panicf("func return arg must (anyData,error) on two arg return")
		}
		c.rsInfo.hasData = true
	}
}

func checkFuncArg(c *callFunc, invokeFuncType reflect.Type) {
	numIn := invokeFuncType.NumIn()
	argTypes := make([]reflect.Type, 0, numIn)
	for i := 0; i < numIn; i++ {
		argTypes = append(argTypes, invokeFuncType.In(i))
	}

	if numIn == 0 || numIn > 3 {
		log.Panicf("expect func args --->  *http.Request|context.Context [http.ResponseWriter] Struct{}|*Struct{}|[]basicType|basicType , but get %s", toJoinName(argTypes))
	}

	first := toArgTypeEnum(argTypes[0])
	if !first.ValidFirstArgType() {
		log.Panicf("expect invoke func first arg %s or %s but get %s", httpRequestType.String(), contextType.String(), first.String())
	}
	c.asInfo.args = append(c.asInfo.args, first)

	//参数只有一个情况下，就不需要绑定参数了
	if numIn == 1 {
		return
	}

	if numIn == 2 {
		secondArgType := argTypes[1]
		second := toArgTypeEnum(secondArgType)
		if !second.ValidSecondArgType() {
			log.Panicf("second arg must one of (%s),but get %s", "[http.ResponseWriter|Struct{}|*Struct{}|[]basicType|basicType]", second.argTypeEnum)
		}
		if !second.IsResponseWriter() {
			c.asInfo.checkBindValue(second)
		}
		c.asInfo.args = append(c.asInfo.args, second)
	}

	if numIn == 3 {
		second := toArgTypeEnum(argTypes[1])
		if !second.IsResponseWriter() {
			log.Panicf("second arg must http.ResponseWriter on three arg")
		}
		c.asInfo.args = append(c.asInfo.args, second)
		three := toArgTypeEnum(argTypes[2])
		if !three.ValidSecondArgType() {
			log.Panicf("three arg must one of (%s),but get %s", "[Struct{}|*Struct{}|[]basicType|basicType]", three.argTypeEnum)
		}
		c.asInfo.checkBindValue(three)
		c.asInfo.args = append(c.asInfo.args, three)
	}
}

func toJoinName(allType []reflect.Type) string {
	builder := strings.Builder{}
	for i := range allType {
		builder.WriteString(allType[i].String() + " ")
	}
	return builder.String()
}

//WithFileName 当参数中是绑定一个文件时使用
func WithFileName(fileName string) CallOption {
	return func(c *callFunc) {
		c.asInfo.fileName = fileName
	}
}

//WithHeaderNames 当超过2个时，必须在结构体中获取到对应的名称的字段，一个时可以通过在header上获取到数据自动绑定
func WithHeaderNames(headerNames ...string) CallOption {
	return func(c *callFunc) {
		c.asInfo.headerNames = headerNames
	}
}

//WithPathNames 当超过2个时，必须在结构体中获取到对应的名称的字段，一个时可以通过在url路径上获取到数据自动绑定
func WithPathNames(pathNames ...string) CallOption {
	return func(c *callFunc) {
		c.asInfo.pathNames = pathNames
	}
}

//WithCookieNames 当超过2个时，必须在结构体中获取到对应的名称的字段，一个时可以通过在cookie上获取到数据自动绑定
func WithCookieNames(cookieNames ...string) CallOption {
	return func(c *callFunc) {
		c.asInfo.cookieNames = cookieNames
	}
}

//WithQueryName queryName 只有当绑定一个从url后面获取或者post的form获取一个参数的时候使用，如果绑定的是slice那么这个字段就是必须设定的
func WithQueryName(queryName string) CallOption {
	return func(c *callFunc) {
		c.asInfo.queryName = queryName
	}
}
