package admin

import (
	"net/http"

	"github.com/go-playground/form"
	"github.com/google/uuid"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/pkg/http/pagination"
	"github.com/devian2011/msgchute/pkg/http/response"
	"github.com/devian2011/msgchute/pkg/http/sort"
)

type MessageFilterRequest struct {
	Status    []dto.MessageStatus `json:"status"`
	IDs       []uuid.UUID         `json:"ids"`
	Recipient []string            `json:"recipient"`
	SenderIDs []string            `json:"sender_ids"`
	Code      []string            `json:"code"`
	Transport []string            `json:"transport"`
}

type MessageFinderResponse struct {
	Messages   []dto.FullMessageInfo `json:"messages"`
	Pagination pagination.PageData   `json:"pagination"`
}

type messageFinderHandler interface {
	Handle(filter *dto.MessageFilter) ([]dto.FullMessageInfo, int, error)
}

type MessageFinderEndpoint struct {
	h       messageFinderHandler
	decoder *form.Decoder
}

func NewMessageFinderEndpoint(h messageFinderHandler) *MessageFinderEndpoint {
	return &MessageFinderEndpoint{
		h:       h,
		decoder: form.NewDecoder(),
	}
}

// ServeHTTP finds and lists messages using query filter, sorting, and pagination parameters
//
//	@Summary		List and filter messages
//	@Description	Retrieves a paginated list of messages filtered by status, IDs, recipients, senders, codes, or transports.
//	@Tags			admin-messages
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Param			page		query		int						false	"Page number for pagination"	default(1)
//	@Param			per_page	query		int						false	"Number of items per page"		default(10)
//	@Param			sort		query		string					false	"Field to sort by"
//	@Param			order		query		string					false	"Sort order direction (asc/desc)"
//	@Param			status		query		[]string				false	"Filter by message status array"	collectionFormat(multi)
//	@Param			ids			query		[]string				false	"Filter by message UUID array"		collectionFormat(multi)
//	@Param			recipient	query		[]string				false	"Filter by recipient array"			collectionFormat(multi)
//	@Param			sender_ids	query		[]string				false	"Filter by sender ID array"			collectionFormat(multi)
//	@Param			code		query		[]string				false	"Filter by message code array"		collectionFormat(multi)
//	@Param			transport	query		[]string				false	"Filter by transport type array"	collectionFormat(multi)
//	@Success		200			{object}	MessageFinderResponse	"List of filtered messages with pagination metadata"
//	@Failure		400			{object}	response.Response		"Invalid query parameters or form parsing failure"
//	@Failure		500			{object}	response.Response		"Internal server error while searching for records"
//	@Router			/api/admin/v1/message [get]
func (e *MessageFinderEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page, perPage := pagination.GetPageDataFromRequest(r)
	limit, offset := pagination.GetLimitsByPage(page, perPage)
	respSort := sort.GetSortFromRequest(r)

	filter := &dto.MessageFilter{
		Limit:     limit,
		Offset:    offset,
		SortField: &respSort.Field,
		SortOrder: &respSort.Order,
	}

	if parseFormErr := r.ParseForm(); parseFormErr != nil {
		response.WriteErrorResponse(w, http.StatusBadRequest, parseFormErr)
		return
	}
	var filterRequest MessageFilterRequest
	if err := e.decoder.Decode(&filterRequest, r.Form); err != nil {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	filter.Status = filterRequest.Status
	filter.IDs = filterRequest.IDs
	filter.Recipient = filterRequest.Recipient
	filter.SenderIDs = filterRequest.SenderIDs
	filter.Transport = filterRequest.Transport
	filter.Code = filterRequest.Code

	messages, totalCnt, getErr := e.h.Handle(filter)
	if getErr != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, getErr)
		return
	}
	response.WriteSuccessResponse(w, http.StatusOK, MessageFinderResponse{
		Messages: messages,
		Pagination: pagination.PageData{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       uint64(totalCnt),
			TotalPages:  uint64(pagination.GetPageCount(int(perPage), totalCnt)),
		},
	})
}

type messageFinderByIDHandler interface {
	Handle(ID uuid.UUID) (*dto.FullMessageInfo, error)
}

type MessageFinderByIDEndpoint struct {
	h messageFinderByIDHandler
}

func NewMessageFinderByIDEndpoint(h messageFinderByIDHandler) *MessageFinderByIDEndpoint {
	return &MessageFinderByIDEndpoint{h: h}
}

// ServeHTTP finds a single message by its unique ID path parameter
//
//	@Summary		Get message by ID
//	@Description	Retrieves full message details for a specific record via its UUID.
//	@Tags			admin-messages
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string				true	"Unique message UUID identifier"	format(uuid)
//	@Success		200	{object}	dto.FullMessageInfo	"Detailed structural message metadata"
//	@Failure		400	{object}	response.Response	"Invalid or malformed UUID parameter format"
//	@Failure		500	{object}	response.Response	"Internal server error or record resolution failure"
//	@Router			/api/admin/v1/message/{id} [get]
func (e *MessageFinderByIDEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ID, getIDErr := response.GetUUIDParam("id", w, r)
	if getIDErr != nil {
		return
	}

	msg, msgGetErr := e.h.Handle(ID)
	if msgGetErr != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, msgGetErr)
		return
	}
	response.WriteSuccessResponse(w, http.StatusOK, msg)
}
