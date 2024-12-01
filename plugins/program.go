package plugins

import (
	"errors"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/fastschema/fastschema/pkg/utils"
)

type Program struct {
	program *goja.Program
	key     string
}

func NewProgram(program *goja.Program, key string) *Program {
	return &Program{
		program: program,
		key:     key,
	}
}

func (p *Program) VerifyJsFunc(functionValue goja.Value) (string, error) {
	functionObject, ok := functionValue.(*goja.Object)
	if !ok {
		return "", fmt.Errorf("functionValue must be an object: %v", functionValue)
	}

	functionNameValue := functionObject.Get("__virtual_name")
	if functionNameValue == nil {
		functionNameValue = functionObject.Get("name")
		if functionNameValue == nil {
			return "", fmt.Errorf("[jsvm] function name is nil")
		}
	}

	exportedFunctionName := functionNameValue.Export()
	functionName, ok := exportedFunctionName.(string)
	if !ok {
		return "", fmt.Errorf("[jsvm] function name is not a string: %v", functionValue)
	}

	if functionName == "" {
		return "", fmt.Errorf("[jsvm] function name is empty. If you are using an inline function, try using a named function instead: %v", functionValue)
	}

	if !IsValidJSFuncName(functionName) {
		return "", fmt.Errorf("[jsvm] invalid function name: %s", functionName)
	}

	vmPool := Pools.Get(CreateVMProps(functionName, p, nil))
	vm := vmPool.Get()
	defer vmPool.Put(vm)

	functionFn := vm.Get(functionName)
	if functionFn == nil {
		return "", fmt.Errorf("[jsvm] function %s is not found", functionName)
	}

	if _, ok = goja.AssertFunction(functionFn); !ok {
		return "", fmt.Errorf("[jsvm] %s is not a function", functionName)
	}

	return functionName, nil
}

func (p *Program) WithFuncName(functionValue goja.Value, cb func(string)) error {
	functionName, err := p.VerifyJsFunc(functionValue)
	if err != nil {
		return err
	}

	cb(functionName)
	return nil
}

func (p *Program) CallFunc(functionName string, set map[string]any, args ...any) (any, error) {
	vmPool := Pools.Get(CreateVMProps(functionName, p, set))
	vm := vmPool.Get()
	defer vmPool.Put(vm)

	functionFn := vm.Get(functionName)
	if functionFn == nil {
		return nil, fmt.Errorf("[jsvm] function %s is not found", functionName)
	}

	fn, ok := goja.AssertFunction(functionFn)
	if !ok {
		return nil, fmt.Errorf("[jsvm] %s is not a function", functionName)
	}

	output, err := fn(goja.Undefined(), utils.Map(args, func(arg any) goja.Value {
		return vm.ToValue(arg)
	})...)

	if err != nil {
		if jserr, ok := err.(*goja.Exception); ok {
			return nil, fmt.Errorf("[jsvm] Exception: %s", jserr.String())
		}

		return nil, err
	}

	if promise, ok := output.Export().(*goja.Promise); ok {
		output = promise.Result()
		if promise.State() == goja.PromiseStateRejected {
			return nil, fmt.Errorf("[jsvm] Promise Rejected: %s", output.String())
		}
	}

	return output.Export(), nil
}

func CreateGoJaProgram(file string, script []byte) (program *goja.Program, content string, err error) {
	if file == "" && len(script) == 0 {
		return nil, "", errors.New("createvm: file or script is required")
	}

	if file != "" && len(script) == 0 {
		if script, err = os.ReadFile(file); err != nil {
			return nil, "", err
		}
	}

	program, err = goja.Compile(file, string(script), true)
	if err != nil {
		return nil, "", err
	}

	return program, string(script), nil
}
