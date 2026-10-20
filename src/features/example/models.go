package example

type ExampleRequest struct {
	Name string `json:"name"`
}

type ExampleResponse struct {
	Message string `json:"message"`
}
