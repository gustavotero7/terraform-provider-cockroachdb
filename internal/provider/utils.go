package provider

func sliceInterfacesToStrings(slice []interface{}) []string {
	res := make([]string, len(slice))
	for i, v := range slice {
		res[i] = v.(string)
	}
	return res
}

func sliceStringsToInterfaces(slice []string) []interface{} {
	res := make([]interface{}, len(slice))
	for i, v := range slice {
		res[i] = v
	}
	return res
}
