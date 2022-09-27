package db

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/arriqaaq/flashdb"
	cprompt "github.com/aschey/bubbleprompt-cobra"
	"github.com/aschey/bubbleprompt/executor"
	"github.com/spf13/cobra"
)

var db *flashdb.FlashDB

func init() {
	config := &flashdb.Config{Path: "./.bin"}
	var err error
	db, err = flashdb.New(config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func GetExecCommand(methodName string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		outStr := ""

		err := db.Update(func(tx *flashdb.Tx) error {
			var err error

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
			retVals := []string{}
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
						retVals = append(retVals, strings.Join(strVals, ","))
					case []string:
						retVals = append(retVals, strings.Join(ifaceVal, ","))
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
			outStr = strings.Join(retVals, " ")
			return nil
		})

		if err != nil {
			return err
		}

		model := executor.NewStringModel(outStr)
		return cprompt.ExecModel(cmd, model)

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
