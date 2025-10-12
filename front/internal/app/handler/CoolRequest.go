package handler

import (
	"RIP/internal/app/ds"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const hardcodedUserID = 1

func (h *Handler) AddComponentToCoolRequest(c *gin.Context) {
	componentID, err := strconv.Atoi(c.Param("component_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	Request, err := h.Repository.GetDraftCoolRequest(hardcodedUserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newRequest := ds.CoolRequest{
			CreatorID: hardcodedUserID,
			Status:    ds.StatusDraft,
		}
		if createErr := h.Repository.CreateCoolRequest(&newRequest); createErr != nil {
			h.errorHandler(c, http.StatusInternalServerError, createErr)
			return
		}
		Request = &newRequest
	} else if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	if err = h.Repository.AddComponentToCoolRequest(Request.ID, uint(componentID)); err != nil {
	}

	c.Redirect(http.StatusFound, "/CoolSystems")
}

func (h *Handler) GetCoolRequest(c *gin.Context) {
	CoolRequestID, err := strconv.Atoi(c.Param("CoolRequest_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	CoolRequest, err := h.Repository.GetCoolRequestWithComponents(uint(CoolRequestID))
	if err != nil {
		var deletedRequest ds.CoolRequest
		c.HTML(http.StatusOK, "coolRequest.html", gin.H{
			"CoolRequest": deletedRequest,
			"Error":       err,
		})
		//h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	if len(CoolRequest.ComponentLink) == 0 {
		//CoolRequestCount := len(CoolRequest.ComponentLink)
		c.HTML(http.StatusOK, "coolRequest.html", CoolRequest)

		//h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty cool-request page, add component first"))
		return
	}

	// c.HTML(http.StatusOK, "coolRequest.html", H.gin{
	// 	CoolRequest
	// }
	// )
	c.HTML(http.StatusOK, "coolRequest.html", CoolRequest)
}

func (h *Handler) DeleteCoolRequest(c *gin.Context) {
	CoolRequestID, _ := strconv.Atoi(c.Param("CoolRequest_id"))

	if err := h.Repository.LogicallyDeleteCoolRequest(uint(CoolRequestID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/CoolSystems")
}
