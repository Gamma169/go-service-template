package main


type SubModel struct {
    Id             string     `json:"id" jsonapi:"primary,sub-models"`
    FoobarModelId  string     `json:"foobarModelId" jsonapi:"attr,foobar-model-id"`
    Value          string     `json:"value" jsonapi:"attr,value"`
    ValueInt       int        `json:"valueInt" jsonapi:"attr,value-int"`
    TempID         string     `json:"__id__" jsonapi:"attr,__id__"`
}


func initSubModelPreparedStatements() {
	
}