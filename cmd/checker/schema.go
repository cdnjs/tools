package main

// validate against schema
// // should be valid json by now
// func test(ctx context.Context) {
// 	testbytes, errrr := ioutil.ReadFile("test.json")
// 	util.Check(err)

// 	schemabytes, errrr := ioutil.ReadFile("schema.json")
// 	util.Check(err)

// 	s, errrr := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemabytes))
// 	util.Check(err)

// 	res, errrr := s.Validate(gojsonschema.NewBytesLoader(testbytes))
// 	util.Check(err) // should be valid json by now

// 	for _, resErr := range res.Errors() {
// 		err(ctx, resErr.String())
// 	}
// }
