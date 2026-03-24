package collaboration

type BroadcastCollaborationRequest struct {
	Event any `json:"event"`
}

type BroadcastCollaborationResponse struct {
	Err error `json:"err,omitzero"`
}

func (r BroadcastCollaborationResponse) Failed() error { return r.Err }
