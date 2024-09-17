package plugins

import (
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/buffer"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/process"
	"github.com/dop251/goja_nodejs/url"
	"github.com/fastschema/fastschema/pkg/utils"
)

type VM struct {
	*goja.Runtime
	poolKey string
}

func (vm *VM) RunWithSets(program *Program, sets ...map[string]any) (any, error) {
	for _, set := range sets {
		for k, v := range set {
			if err := vm.Set(k, v); err != nil {
				return nil, err
			}
		}
	}

	return vm.RunProgram(program.program)
}

type VMProps struct {
	function string
	program  *Program
	set      map[string]any
}

func CreateVMProps(function string, program *Program, set map[string]any) *VMProps {
	return &VMProps{
		function: function,
		program:  program,
		set:      set,
	}
}

func (p *VMProps) Key() string {
	poolKey := p.program.key + "." + p.function
	if len(p.set) > 0 {
		setKeys := utils.GetMapKeys(p.set)
		poolKey += "_" + strings.Join(setKeys, ".")
	}

	return poolKey
}

type VMPool struct {
	sync.Pool
}

func (p *VMPool) Get() *VM {
	return p.Pool.Get().(*VM)
}

type VMPools struct {
	sync.Mutex
	pools map[string]*VMPool
}

func (p *VMPools) Get(props *VMProps) *VMPool {
	p.Lock()
	defer p.Unlock()

	poolKey := props.Key()
	pool, ok := p.pools[poolKey]
	if !ok {
		pool = &VMPool{
			Pool: sync.Pool{
				New: func() interface{} {
					// fmt.Println("> Creating new VM for pool:", poolKey)
					gojaVM := goja.New()
					gojaVM.SetFieldNameMapper(goja.TagFieldNameMapper("json", false))
					Require.Enable(gojaVM)
					console.Enable(gojaVM)
					buffer.Enable(gojaVM)
					process.Enable(gojaVM)
					url.Enable(gojaVM)

					vm := &VM{
						Runtime: gojaVM,
						poolKey: poolKey,
					}

					if _, err := vm.RunWithSets(props.program, props.set); err != nil {
						panic(err)
					}

					return vm
				},
			},
		}
		p.pools[poolKey] = pool
	}

	return pool
}

var Pools = &VMPools{
	pools: map[string]*VMPool{},
}
