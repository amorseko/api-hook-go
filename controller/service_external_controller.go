package controller

import (
	"api-hook/utils"
	"net/http"
)

func ExternalService(w http.ResponseWriter, r *http.Request) {
	_, isObject, data, dataArray := utils.CallExternalAPI(r)
	if isObject {
		utils.Respond(w, data)
	} else {
		utils.RespondArray(w, dataArray)
	}
}
