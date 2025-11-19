package entity

import (
	"fmt"

	"github.com/buger/jsonparser"

	"github.com/fastschema/fastschema/pkg/utils"
)

// MarshalJSON implements the json.Marshaler interface.
func (e *Entity) MarshalJSON() ([]byte, error) {
	return e.data.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Entity) UnmarshalJSON(data []byte) (err error) {
	return jsonparser.ObjectEach(
		data,
		func(keyData []byte, valueData []byte, dataType jsonparser.ValueType, offset int) error {
			if dataType == jsonparser.Object {
				e.Set(string(keyData), utils.Must(NewEntityFromJSON(string(valueData))))
				return nil
			}

			// Process array data
			if dataType == jsonparser.Array {
				entityValues := []*Entity{}
				nonEntityValues := []any{}

				// Process array data
				if _, err := jsonparser.ArrayEach(valueData, func(
					itemValue []byte,
					itemDataType jsonparser.ValueType,
					itemOffset int,
					err error,
				) {
					if itemDataType == jsonparser.Object {
						entityValues = append(entityValues, utils.Must(NewEntityFromJSON(string(itemValue))))
						return
					}

					iArrayItemValue := utils.Must(UnmarshalJSONValue(itemValue, itemDataType))
					nonEntityValues = append(nonEntityValues, iArrayItemValue)
				}); err != nil {
					return err
				}

				/** Process array data **/
				if len(entityValues) > 0 && len(nonEntityValues) > 0 {
					return fmt.Errorf(
						"cannot mix entities and non-entities in a slice: %s=%v",
						string(keyData),
						string(valueData),
					)
				}

				if len(entityValues) > 0 {
					e.Set(string(keyData), entityValues)
					return nil
				}

				if len(nonEntityValues) > 0 {
					e.Set(string(keyData), nonEntityValues)
					return nil
				}

				// Fallback for empty array
				e.Set(string(keyData), nil)

				/** End process array data **/

				return nil
			}

			e.Set(string(keyData), utils.Must(UnmarshalJSONValue(valueData, dataType)))

			return nil
		},
	)
}
