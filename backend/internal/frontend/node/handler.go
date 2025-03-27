package frontendnode

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/gorilla/mux"
)

type FrontendNodeHandler struct {
	FrontendNodeService *FrontendNodeService
}

// NewFrontendAgentHandler initializes the handler
func NewFrontendNodeHandler(frontendNodeServices *FrontendNodeService) *FrontendNodeHandler {
	return &FrontendNodeHandler{
		FrontendNodeService: frontendNodeServices,
	}
}

func (f *FrontendNodeHandler) GetComponent(w http.ResponseWriter, r *http.Request) {
	componentType := r.URL.Query().Get("type")

	utils.Logger.Info(fmt.Sprintf("Received request to get all components of type: %s", componentType))

	resp, err := f.FrontendNodeService.GetComponents(componentType)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error occured while getting components: %v", err))
		utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, resp)
}

func (f *FrontendNodeHandler) GetComponentSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	utils.Logger.Info(fmt.Sprintf("Received request to get schema for component: %s", name))

	schema, err := f.FrontendNodeService.GetComponentSchemaByName(name)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("Error occurred while getting schema for %s: %v", name, err))

		if errors.Is(err, sql.ErrNoRows) {
			utils.SendJSONError(w, http.StatusOK, "Schema not found")
		} else {
			utils.SendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, schema)
}
