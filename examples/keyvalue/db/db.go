package db

import (
	"examples/keyvalue/model"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/arriqaaq/flashdb"
	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/aschey/bubbleprompt/executor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var dbMutex sync.Mutex
var db *flashdb.FlashDB

func LoadDb() error {
	config := &flashdb.Config{Path: "./.bin"}
	var err error
	dbMutex.Lock()
	defer dbMutex.Unlock()
	db, err = flashdb.New(config)
	return err
}

func init() {
	err := LoadDb()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}

func getKeys(args []string, toComplete string, numAllowedArgs int, keyFunc func(tx *flashdb.Tx) []string) (keys []string, directive cobra.ShellCompDirective) {
	directive = cobra.ShellCompDirectiveDefault
	if numAllowedArgs > -1 && len(args) > numAllowedArgs-1 {
		return
	}
	dbMutex.Lock()
	defer dbMutex.Unlock()
	db.View(func(tx *flashdb.Tx) error {
		keys = cprompt.FilterShellCompletions(keyFunc(tx), toComplete)
		return nil
	})
	return
}

func GetKeys(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return getKeys(args, toComplete, 1, func(tx *flashdb.Tx) []string { return tx.Keys() })
}

func HGetKeys(_ *cobra.Command, args []string, toComplete string) (keys []string, directive cobra.ShellCompDirective) {
	return getKeys(args, toComplete, 1, func(tx *flashdb.Tx) []string { return tx.HKeys() })
}

func ZGetKeys(_ *cobra.Command, args []string, toComplete string) (keys []string, directive cobra.ShellCompDirective) {
	return getKeys(args, toComplete, 1, func(tx *flashdb.Tx) []string { return tx.ZKeys() })
}

func SGetKeys(_ *cobra.Command, args []string, toComplete string) (keys []string, directive cobra.ShellCompDirective) {
	return getKeys(args, toComplete, 1, func(tx *flashdb.Tx) []string { return tx.SKeys() })
}

func SGetKeysN(numAllowedArgs int) func(_ *cobra.Command, args []string, toComplete string) (keys []string, directive cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) (keys []string, directive cobra.ShellCompDirective) {
		return getKeys(args, toComplete, numAllowedArgs, func(tx *flashdb.Tx) []string { return tx.SKeys() })
	}
}

func GetExecCommand(methodName string) func(cmd *cobra.Command, args []string) error {
	return getExecCommand(methodName, false)
}

func GetListExecCommand(methodName string) func(cmd *cobra.Command, args []string) error {
	return getExecCommand(methodName, true)
}

func getExecCommand(methodName string, returnList bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		retVals := []string{}
		dbMutex.Lock()
		defer dbMutex.Unlock()
		err := db.Update(func(tx *flashdb.Tx) error {
			method, found := reflect.TypeOf(tx).MethodByName(methodName)
			if !found {
				return fmt.Errorf("command not found")
			}
			expectedParams := method.Type.NumIn()
			isVariadic := method.Type.In(expectedParams-1).Kind() == reflect.Slice
			if (isVariadic && len(args) < expectedParams-1) || (!isVariadic && len(args) != expectedParams-1) {
				// Subtract one for the tx object
				return fmt.Errorf("expected %d params but got %d", expectedParams-1, len(args))
			}
			paramVals, err := getReflectParams(args, tx, method.Type)
			if err != nil {
				return err
			}

			out := method.Func.Call(paramVals)

			for _, outVal := range out {
				if outVal.CanInterface() {
					iface := outVal.Interface()
					if iface == nil {
						continue
					}
					switch ifaceVal := iface.(type) {
					case error:
						return fmt.Errorf(ifaceVal.Error())
					case []interface{}:
						strVals := []string{}
						for _, s := range ifaceVal {
							strVals = append(strVals, fmt.Sprintf("%v", s))
						}
						retVals = append(retVals, strVals...)
					case []string:
						retVals = append(retVals, ifaceVal...)
					case string:
						retVals = append(retVals, ifaceVal)
					case bool:
						retVals = append(retVals, strconv.FormatBool(ifaceVal))
					case int64:
						retVals = append(retVals, strconv.FormatInt(ifaceVal, 10))
					case int:
						retVals = append(retVals, strconv.FormatInt(int64(ifaceVal), 10))
					case float64:
						retVals = append(retVals, strconv.FormatFloat(float64(ifaceVal), 'f', 3, 64))
					case float32:
						retVals = append(retVals, strconv.FormatFloat(float64(ifaceVal), 'f', 3, 32))

					}
				} else {
					retVals = append(retVals, outVal.String())
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		var retModel tea.Model
		if returnList {
			retModel = model.NewList(retVals)
		} else {
			retModel = executor.NewStringModel(strings.Join(retVals, ","))
		}

		return cprompt.ExecModel(cmd, retModel)
	}

}

func getReflectParams(params []string, tx *flashdb.Tx, methodType reflect.Type) ([]reflect.Value, error) {
	paramVals := []reflect.Value{reflect.ValueOf(tx)}
	for i, p := range params {
		var reflectVal any
		var err error
		methodParam := methodType.In(i + 1)
		switch methodParam.Kind() {
		case reflect.Int:
			var intVal int64
			intVal, err = strconv.ParseInt(p, 10, 32)
			reflectVal = int(intVal)
		case reflect.Int64:
			reflectVal, err = strconv.ParseInt(p, 10, 64)
		case reflect.Float32:
			reflectVal, err = strconv.ParseFloat(p, 32)
		case reflect.Float64:
			reflectVal, err = strconv.ParseFloat(p, 64)
		case reflect.String:
			reflectVal = p
		case reflect.Slice:
			for j := i; j < len(params); j++ {
				paramVals = append(paramVals, reflect.ValueOf(params[j]))
			}
			return paramVals, nil
		}
		if err != nil {
			return nil, err
		}
		paramVals = append(paramVals, reflect.ValueOf(reflectVal))

	}
	return paramVals, nil
}
