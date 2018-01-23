package models

import "qiniu.ai/lib/model"

var (
	mongo *model.Model
)

func SetupModel(model *model.Model) {
	mongo = model
}

func Model() *model.Model {
	return mongo
}
