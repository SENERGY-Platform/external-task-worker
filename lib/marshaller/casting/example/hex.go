package example

const Hex = "example_hex"

//color concept uses hex -> do nothing
func init() {
	conceptToCharacteristic.Set(Hex, func(concept interface{}) (out interface{}, err error) {
		return concept, nil
	})

	characteristicToConcept.Set(Hex, func(in interface{}) (concept interface{}, err error) {
		return in, nil
	})
}
